package interpreter

import (
	"sort"

	"github.com/backbone81/golr/internal/backendtest"
	parserbackend "github.com/backbone81/golr/internal/parsergen/backend"
	parsertable "github.com/backbone81/golr/internal/parsergen/backend/table"
	scannerbackend "github.com/backbone81/golr/internal/scannergen/backend"
)

// eofTerminalIdx is the terminal index of the end of input symbol. frontend.AugmentGrammar always inserts it as the
// first terminal, so the index is fixed at 0 for every augmented grammar.
const eofTerminalIdx = 0

// noTerminalIdx is the terminal index a token gets which is no terminal of the grammar. It is not a terminal index at
// all but the one value which table.TerminalColumn turns into table.NoTerminalColumn, the column no state of the parser
// has an entry in. A lookup in that column therefore finds nothing and falls through to the default action of the
// state, which is exactly what a generated parser does with a token it does not know.
const noTerminalIdx = -1

// invalidTerminalName is the name the trace uses for a token which no scanner rule matched. It is spelled with the same
// leading dollar sign as frontend.SymbolEOF and frontend.SymbolError, which no terminal a frontend can read is allowed
// to carry, so it can not collide with the name of a rule. Every runner prints this name for whatever its own scanner
// calls the invalid token.
const invalidTerminalName = "$invalid"

// errorSymbolName is what the trace calls the error symbol, matching frontend.SymbolError.Name.
const errorSymbolName = "$error"

// errorRecoveryShifts is the number of tokens which have to be shifted after a syntax error before errors are reported
// again. Suppressing the errors in between keeps a single mistake in the input from producing an avalanche of messages
// which are all consequences of the first one. It is the number of tokens yacc and GNU Bison use, see section 7 "Error
// Handling" of the yacc report, and the number the generated drivers use.
const errorRecoveryShifts = 3

// token is one token as the parser sees it: which terminal of the grammar it is, what to call it in the trace, and the
// extent of the input it covers.
//
// The name is the name of the scanner rule which matched and not the name of the terminal, because that is what a
// generated parser has at hand: it prints the token its scanner delivered. The two are the same name for every terminal
// the grammar and the scanner share, and only a token which is no terminal of the grammar has a name the grammar does
// not know.
type token struct {
	// terminalIdx is the terminal of the grammar the token stands for, or noTerminalIdx for a token which is none.
	terminalIdx int

	// name is what the trace calls the token.
	name string

	// start is the byte offset the token begins at, and the offset an error on it is reported at.
	start int

	// end is the byte offset one past the last byte of the token.
	end int
}

// nodeKind tells the three kinds of node of a parse tree apart. There is one trace event per kind, and the kind is
// what a walk of the tree decides on, so the two are deliberately the same three cases.
//
// A terminal and a nonterminal cannot be told apart by their children, which is why this exists at all: a production
// with an empty right hand side reduces to a node with no children, exactly like a leaf.
type nodeKind int

const (
	// nodeKindTerminal is a leaf, standing for a token the parser shifted.
	nodeKindTerminal nodeKind = iota

	// nodeKindNonterminal is an internal node, standing for a production the parser reduced.
	nodeKindNonterminal

	// nodeKindErrorSymbol is the leaf error recovery pushed in the state it resumed in. It stands for the part of the
	// input which was dropped, so unlike a terminal it names no token of its own.
	nodeKindErrorSymbol
)

// node mirrors one entry of a generated parser's node stack, carrying only what a trace line needs.
type node struct {
	kind nodeKind

	// name is what the trace calls the symbol: a scanner rule, a nonterminal, or errorSymbolName.
	name string
}

// Parser is the reference parser of the backend test harness. It reads the compressed tables the way a generated table
// driven parser does, takes its tokens from the reference scanner, and emits the canonical trace, one line per action.
// The order in which CompressedParser.Action consults the tables is what decides where an error is detected.
type Parser struct {
	parser     parserbackend.Parser
	compressed parsertable.CompressedParser
	scanner    *Scanner

	// terminalIdxByRuleIdx translates a scanner rule into the terminal of the grammar it stands for, or into
	// noTerminalIdx for a rule the grammar never mentions.
	terminalIdxByRuleIdx []int

	// lineStarts holds the byte offset each line begins at, so lineCol can turn an offset into a line and column.
	lineStarts []int

	// stateStack mirrors the state stack of a generated parser. It starts with the start state 0.
	stateStack []int

	// nodeStack mirrors the node stack of a generated parser. It carries one node per symbol which was shifted or
	// reduced, so it is always one shorter than stateStack, and the two grow and shrink together. Dropping a state
	// during error recovery drops the node with it.
	nodeStack []node

	// token is the token the parser currently looks at, which a shift and a discard advance past.
	token token

	// errorRecoveryShiftsRemaining counts down the tokens which still have to be shifted before syntax errors are
	// reported again. It is zero while the parser is in sync with the input and errorRecoveryShifts right after the
	// error symbol was shifted.
	errorRecoveryShiftsRemaining int

	// stepCount and maxSteps bound the number of actions, so that a table which makes the parser loop fails the
	// suite instead of hanging it. A well-formed table over a finite input stays far below the bound.
	stepCount int
	maxSteps  int

	// trace collects the events in the order they happened. It is the whole output of a parse.
	trace backendtest.Trace
}

// NewParser creates a parser for the given parser tables which takes its tokens from the given scanner. It compresses
// the tables itself, instead of taking the compressed ones, so that the names it reports and the tables it reads can
// never come from two different parsers. The compression is lossless, which the exhaustive equivalence test of the
// table package proves for every state, terminal and nonterminal.
func NewParser(parser parserbackend.Parser, scanner *Scanner) *Parser {
	result := &Parser{
		parser:               parser,
		compressed:           parsertable.NewCompressedParser(parser),
		scanner:              scanner,
		terminalIdxByRuleIdx: newTerminalIdxByRuleIdx(parser, scanner),
		lineStarts:           newLineStarts(scanner.source),
		stateStack:           []int{0},
		maxSteps:             maxSteps(len(scanner.source), len(parser.States)),
	}
	// A generated parser reads the first token before it enters its loop, so that the loop always has one to decide
	// on.
	result.advanceToken()
	return result
}

// newLineStarts returns the byte offset each line of the source begins at.
func newLineStarts(source []byte) []int {
	result := []int{0}
	for offset, value := range source {
		if value == '\n' {
			result = append(result, offset+1)
		}
	}
	return result
}

// newTerminalIdxByRuleIdx returns, for every rule of the scanner, the terminal of the grammar the rule stands for, or
// noTerminalIdx for a rule which is no terminal of the grammar.
//
// The two are matched by name, which is what the generated code does as well: the parser backend emits the terminals of
// the grammar under the names of the token constants of the scanner, and lets the compiler bring the two together. A
// scanner may well have rules the grammar never mentions, which is the case this translation exists for.
func newTerminalIdxByRuleIdx(parser parserbackend.Parser, scanner *Scanner) []int {
	terminalIdxByName := make(map[string]int, len(parser.Grammar.Terminals))
	for terminalIdx, terminal := range parser.Grammar.Terminals {
		terminalIdxByName[terminal.Name] = terminalIdx
	}

	result := make([]int, len(scanner.rules))
	for ruleIdx, rule := range scanner.rules {
		terminalIdx, ok := terminalIdxByName[rule.Name]
		if !ok {
			terminalIdx = noTerminalIdx
		}
		result[ruleIdx] = terminalIdx
	}
	return result
}

// maxSteps returns the number of actions a parse may take before it is given up as looping. Every shift consumes a
// token, of which there can be no more than one per byte of the input, and the reductions between two shifts are
// bounded by the size of the automaton. The bound is loose on purpose: it exists to turn a malformed or cyclic table
// into a failing test rather than a hanging one, and never to cut a parse short.
func maxSteps(sourceLen int, stateCount int) int {
	return (sourceLen+2)*(stateCount+1)*4 + 1024
}

// Parse runs the whole parse and returns its canonical trace, one line per action. It can only be called once, because
// a parse consumes the scanner it was created with.
func (p *Parser) Parse() backendtest.Trace {
	for {
		p.stepCount++
		if p.stepCount > p.maxSteps {
			return p.trace
		}

		switch action := p.action(); action.Kind() {
		case parsertable.ActionKindShift:
			p.shift(action.StateIdx())
		case parsertable.ActionKindReduce:
			p.reduce(action.ProductionIdx())
		case parsertable.ActionKindAccept:
			line, column := lineCol(p.lineStarts, p.token.start)
			p.trace = append(p.trace, backendtest.Accept{Line: line, Column: column})
			return p.trace
		case parsertable.ActionKindError:
			if !p.recoverFromError() {
				return p.trace
			}
		}
	}
}

// action returns what the state on top of the stack does with the current token.
//
// A state which has no action of its own for the token and no default action either reports no action at all, which the
// error action stands in for here. Both make the token a syntax error, and the generated tables carry the error action
// for such a state for the same reason: it is a value like any other and keeps the emitted table free of a negative
// entry.
func (p *Parser) action() parsertable.Action {
	action := p.compressed.Action(p.stateStack[len(p.stateStack)-1], p.token.terminalIdx)
	if action == parsertable.NoAction {
		return parsertable.NewErrorAction()
	}
	return action
}

// shift consumes the current token, continues in the given state, and moves the parser one token closer to trusting its
// position again after an error. The synthetic shift of the end of input symbol goes through here too, so it appears in
// the trace like any other shift.
func (p *Parser) shift(stateIdx int) {
	line, column := lineCol(p.lineStarts, p.token.start)
	p.trace = append(p.trace, backendtest.Shift{
		Line:         line,
		Column:       column,
		TerminalName: p.token.name,
		Lexeme:       string(p.scanner.source[p.token.start:p.token.end]),
	})

	p.nodeStack = append(p.nodeStack, node{kind: nodeKindTerminal, name: p.token.name})
	p.stateStack = append(p.stateStack, stateIdx)
	p.advanceToken()

	if p.errorRecoveryShiftsRemaining > 0 {
		// Getting tokens of the input shifted again is what makes the parser trust its position again.
		p.errorRecoveryShiftsRemaining--
	}
}

// reduce replaces the right hand side of the given production on the stacks with the nonterminal on its left hand
// side, and continues in the state the goto of the uncovered state leads to. The right hand side is read off the node
// stack for the trace before it is cut back.
func (p *Parser) reduce(productionIdx int) {
	production := p.parser.Grammar.Productions[productionIdx]
	popCount := len(production.SymbolRefs)
	leftHandSide := p.parser.Grammar.Nonterminals[production.NonterminalIdx].Name

	rightHandSide := make([]string, popCount)
	for i, child := range p.nodeStack[len(p.nodeStack)-popCount:] {
		rightHandSide[i] = child.name
	}
	line, column := lineCol(p.lineStarts, p.token.start)
	p.trace = append(p.trace, backendtest.Reduce{
		Line:          line,
		Column:        column,
		LeftHandSide:  leftHandSide,
		RightHandSide: rightHandSide,
	})

	p.nodeStack = append(p.nodeStack[:len(p.nodeStack)-popCount], node{
		kind: nodeKindNonterminal,
		name: leftHandSide,
	})

	p.stateStack = p.stateStack[:len(p.stateStack)-popCount]
	uncoveredStateIdx := p.stateStack[len(p.stateStack)-1]
	p.stateStack = append(p.stateStack, p.compressed.Goto(uncoveredStateIdx, production.NonterminalIdx))
}

// recoverFromError reports the syntax error the parser ran into and puts it back into a state where it can continue on
// the remaining input. It reports whether that succeeded. Once it did not, the parse is given up.
//
// This is the panic mode recovery of section 9 "Error Recovery" of "LR Parsing" by Aho and Johnson, in the shape which
// section 7 "Error Handling" of the yacc report describes, and in the order the generated drivers perform it: report,
// discard, pop, resume. Every step is its own trace line, so the whole recovery is visible.
//
// The token which caused the error is kept for the resumed state to look at, because that state is usually waiting for
// exactly it - in a production like "{" @error "}" it is the closing brace which ends the recovery. Only when the parse
// fails on that very same token again is the token discarded, which is what an untouched countdown of tokens to shift
// tells us. Popping and discarding in the same round is what guarantees progress: every round either gets the parse
// going again or consumes one token.
func (p *Parser) recoverFromError() bool {
	line, column := lineCol(p.lineStarts, p.token.start)
	p.trace = append(p.trace, backendtest.ParserError{
		Line:   line,
		Column: column,
		Detail: "unexpected token " + p.token.name,
		// A generated parser hides errors it hits while still recovering; the trace keeps them, marked.
		Suppressed: p.errorRecoveryShiftsRemaining != 0,
	})

	if p.errorRecoveryShiftsRemaining == errorRecoveryShifts {
		// Nothing was shifted since the last error, so the parser is failing on the token it already failed on and
		// keeping it would only lead here again.
		if p.token.terminalIdx == eofTerminalIdx {
			// The end of the input is the one token which can not be discarded, so there is nothing left to try.
			p.trace = append(p.trace, backendtest.Fail{Line: line, Column: column})
			return false
		}
		p.trace = append(p.trace, backendtest.Discard{
			Line:         line,
			Column:       column,
			TerminalName: p.token.name,
			Lexeme:       string(p.scanner.source[p.token.start:p.token.end]),
		})
		p.advanceToken()
	}
	p.errorRecoveryShiftsRemaining = errorRecoveryShifts

	return p.popToErrorState()
}

// popToErrorState drops states off the stack until one of them can shift the error symbol, and shifts it there,
// reporting whether it found such a state. A grammar which marks no place to resume at unwinds the whole stack here.
func (p *Parser) popToErrorState() bool {
	for {
		line, column := lineCol(p.lineStarts, p.token.start)
		if stateIdx, ok := p.compressed.ErrorShiftStateIdx(p.stateStack[len(p.stateStack)-1]); ok {
			p.trace = append(p.trace, backendtest.Resync{Line: line, Column: column})
			// Shift the error symbol. Its node stands for the part of the input which was dropped.
			p.stateStack = append(p.stateStack, stateIdx)
			p.nodeStack = append(p.nodeStack, node{kind: nodeKindErrorSymbol, name: errorSymbolName})
			return true
		}
		if len(p.stateStack) == 1 {
			// Only the state the parse started in is left and it can not shift the error symbol either, so no place
			// the grammar marked to resume at covers the position of the error.
			p.trace = append(p.trace, backendtest.Fail{Line: line, Column: column})
			return false
		}
		// The two stacks hold one node per symbol and one state on top of that, so dropping one state drops one node.
		p.trace = append(p.trace, backendtest.Pop{
			Line:       line,
			Column:     column,
			SymbolName: p.nodeStack[len(p.nodeStack)-1].name,
		})
		p.stateStack = p.stateStack[:len(p.stateStack)-1]
		p.nodeStack = p.nodeStack[:len(p.nodeStack)-1]
	}
}

// lineCol turns a byte offset into a one based line and column, matching how the generated scanners count. lineStarts
// is the offset each line begins at, from newLineStarts.
func lineCol(lineStarts []int, offset int) (int, int) {
	line := sort.Search(len(lineStarts), func(i int) bool { return lineStarts[i] > offset })
	return line, offset - lineStarts[line-1] + 1
}

// advanceToken reads the next token the parser has to decide on, skipping the rules the scanner marks as skipped the
// way the token skipper of a generated scanner does.
//
// A token which no rule matched is not skipped but handed on, as the invalid token of a generated scanner is. It is no
// terminal of the grammar, so the parser takes the default action of its state for it and detects the error wherever
// that action leads - which is a place every language has to agree on.
//
// The end of the input is the EOF terminal at the offset one past the last byte, and stays there however often it is
// read, so the parser can keep deciding on it while it reduces.
func (p *Parser) advanceToken() {
	for {
		match, ok := p.scanner.Next()
		if !ok {
			p.token = token{
				terminalIdx: eofTerminalIdx,
				name:        p.parser.Grammar.Terminals[eofTerminalIdx].Name,
				start:       len(p.scanner.source),
				end:         len(p.scanner.source),
			}
			return
		}

		rule, ok := p.scanner.rule(match)
		if !ok {
			p.token = token{
				terminalIdx: noTerminalIdx,
				name:        invalidTerminalName,
				start:       match.Start,
				end:         match.End,
			}
			return
		}
		if rule.Skip {
			continue
		}

		p.token = token{
			terminalIdx: p.terminalIdxByRuleIdx[match.RuleIdx],
			name:        rule.Name,
			start:       match.Start,
			end:         match.End,
		}
		return
	}
}

// ParseTrace scans and parses the whole input and returns the canonical trace of the parse. This is one corpus case end
// to end: a grammar and an input go in, and the trace every backend has to reproduce comes out.
func ParseTrace(parser parserbackend.Parser, dfa scannerbackend.DFA, source []byte) backendtest.Trace {
	return NewParser(parser, NewScanner(dfa, source)).Parse()
}
