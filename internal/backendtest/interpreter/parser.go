package interpreter

import (
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

// Parser is the reference parser of the backend test harness. It reads the compressed tables the same way a generated
// table driven parser reads them, takes its tokens from the reference scanner, and records what it did as the canonical
// trace.
//
// Everything which distinguishes one state from another is in the tables, so the whole automaton is the loop in Parse:
// look up the action of the state on top of the stack for the current token, and shift, reduce, accept or recover. The
// order in which a lookup consults the tables is what decides where an error is detected, see CompressedParser.Action.
type Parser struct {
	parser     parserbackend.Parser
	compressed parsertable.CompressedParser
	scanner    *Scanner

	// terminalIdxByRuleIdx translates a scanner rule into the terminal of the grammar it stands for, or into
	// noTerminalIdx for a rule the grammar never mentions.
	terminalIdxByRuleIdx []int

	// stateStack mirrors the state stack of a generated parser. It starts with the start state 0. There is no node
	// stack next to it, because a trace records no parse tree: the two stacks of a generated parser grow and shrink
	// together, so the state stack alone says everything a trace reports.
	stateStack []int

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

	// trace collects the events of the parse in the order they happened.
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
		stateStack:           []int{0},
		maxSteps:             maxSteps(len(scanner.source), len(parser.States)),
	}
	// A generated parser reads the first token before it enters its loop, so that the loop always has one to decide
	// on.
	result.advanceToken()
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

// Parse runs the whole parse and returns the canonical trace of it. It can only be called once, because a parse
// consumes the scanner it was created with.
//
// A parse which succeeds ends in the accept event. A parse which does not ends wherever the error recovery gave up, so
// the last event of such a trace is the error, the discarded token or the states the recovery popped in vain. There is
// no event for the end of the input, unlike in a scanner trace, because a parser reaching the end of the input is not
// an outcome of its own - it either accepts there or reports an error.
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
			p.trace = append(p.trace, backendtest.Accept{})
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
// position again after an error.
func (p *Parser) shift(stateIdx int) {
	p.trace = append(p.trace, backendtest.Shift{
		TerminalName: p.token.name,
		Start:        p.token.start,
		End:          p.token.end,
	})
	p.stateStack = append(p.stateStack, stateIdx)
	p.advanceToken()

	if p.errorRecoveryShiftsRemaining > 0 {
		// Getting tokens of the input shifted again is what makes the parser trust its position again.
		p.errorRecoveryShiftsRemaining--
	}
}

// reduce replaces the right hand side of the given production on the stack with the nonterminal on its left hand side,
// and continues in the state the goto of the uncovered state leads to.
//
// The trace names the production by its left hand side and the length of its right hand side rather than by its index,
// because production numbering is a property of the core which built the tables and not of the grammar.
func (p *Parser) reduce(productionIdx int) {
	production := p.parser.Grammar.Productions[productionIdx]
	popCount := len(production.SymbolRefs)

	p.trace = append(p.trace, backendtest.Reduce{
		NonterminalName:     p.parser.Grammar.Nonterminals[production.NonterminalIdx].Name,
		RightHandSideLength: popCount,
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
// discard, pop, resume.
//
// The token which caused the error is kept for the resumed state to look at, because that state is usually waiting for
// exactly it - in a production like "{" @error "}" it is the closing brace which ends the recovery. Only when the parse
// fails on that very same token again is the token discarded, which is what an untouched countdown of tokens to shift
// tells us. Popping and discarding in the same round is what guarantees progress: every round either gets the parse
// going again or consumes one token of the input.
func (p *Parser) recoverFromError() bool {
	if p.errorRecoveryShiftsRemaining == 0 {
		// While recovering, the parser is not in sync with the input, so the errors it runs into there are most
		// likely consequences of the error which is already reported. Reporting those as well is the avalanche of
		// messages the countdown exists to prevent. Only the report is suppressed, never the recovery itself, so
		// which errors a trace does not carry is as much a part of it as which errors it does.
		p.trace = append(p.trace, backendtest.ParserError{Offset: p.token.start})
	}

	if p.errorRecoveryShiftsRemaining == errorRecoveryShifts {
		// Nothing was shifted since the last error, so the parser is failing on the token it already failed on and
		// keeping it would only lead here again.
		if p.token.terminalIdx == eofTerminalIdx {
			// The end of the input is the one token which can not be discarded, so there is nothing left to try.
			return false
		}
		p.trace = append(p.trace, backendtest.Discard{
			TerminalName: p.token.name,
			Start:        p.token.start,
			End:          p.token.end,
		})
		p.advanceToken()
	}
	p.errorRecoveryShiftsRemaining = errorRecoveryShifts

	return p.popToErrorState()
}

// popToErrorState drops states off the stack until one of them can shift the error symbol, and shifts it there. It
// reports whether it found such a state.
//
// The states which can shift the error symbol are the places the grammar marked to resume at. Everything the popped
// states had parsed is dropped with them. A grammar which marks no place to resume at has no such state anywhere, so
// the recovery unwinds the whole stack and the parse is over - which is the behavior of every parser for a grammar
// without error recovery, and is why the trace still reports what was popped.
//
// The number of popped states is reported as one event rather than one per state, and it is reported even when nothing
// was popped. Every round of recovery therefore has exactly one of these events, which keeps the number of events a
// runner has to produce independent of how far the recovery had to unwind.
func (p *Parser) popToErrorState() bool {
	popCount := 0
	for {
		if stateIdx, ok := p.compressed.ErrorShiftStateIdx(p.stateStack[len(p.stateStack)-1]); ok {
			p.trace = append(p.trace, backendtest.Pop{Count: popCount})
			p.stateStack = append(p.stateStack, stateIdx)
			p.trace = append(p.trace, backendtest.Resync{})
			return true
		}
		if len(p.stateStack) == 1 {
			// Only the state the parse started in is left and it can not shift the error symbol either, so no
			// place the grammar marked to resume at covers the position of the error.
			p.trace = append(p.trace, backendtest.Pop{Count: popCount})
			return false
		}
		p.stateStack = p.stateStack[:len(p.stateStack)-1]
		popCount++
	}
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
