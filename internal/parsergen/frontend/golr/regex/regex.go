package regex

import (
	"errors"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/pkg/scannergen/frontend/dsl"
)

// Parse parses a regular expression enclosed in / delimiters and returns the corresponding Node tree.
func Parse(input []byte) (*frontend.Node, error) {
	var p Parser
	return p.Parse(input)
}

type Parser struct {
	input    []byte
	pos      int
	currRune rune
	atEOF    bool
}

func (p *Parser) Parse(input []byte) (*frontend.Node, error) {
	if len(input) < 2 || input[0] != '/' || input[len(input)-1] != '/' {
		return nil, fmt.Errorf("regular expression must be enclosed in /")
	}
	p.input = input[1 : len(input)-1]
	p.pos = 0
	p.atEOF = false
	p.currRune = 0

	if !p.next() {
		return nil, fmt.Errorf("empty regular expression")
	}

	node, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}
	if !p.atEOF {
		return nil, fmt.Errorf("unexpected character %q", p.rune())
	}
	return node, nil
}

func (p *Parser) next() bool {
	if p.pos >= len(p.input) {
		p.atEOF = true
		return false
	}
	if p.input[p.pos] < utf8.RuneSelf {
		p.currRune = rune(p.input[p.pos])
		p.pos++
	} else {
		r, size := utf8.DecodeRune(p.input[p.pos:])
		p.currRune = r
		p.pos += size
	}
	return true
}

func (p *Parser) rune() rune {
	return p.currRune
}

// parseAlternation: concatenation ("|" concatenation)*
func (p *Parser) parseAlternation() (*frontend.Node, error) {
	first, err := p.parseConcatenation()
	if err != nil {
		return nil, err
	}
	if p.atEOF || p.rune() != '|' {
		return first, nil
	}
	children := []*frontend.Node{first}
	for !p.atEOF && p.rune() == '|' {
		p.next() // consume '|'
		if p.atEOF {
			return nil, fmt.Errorf("unexpected end of input after |")
		}
		child, err := p.parseConcatenation()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return &frontend.Node{
		Kind: frontend.KindOr,
		Or:   frontend.Or{Children: children},
	}, nil
}

// parseConcatenation: quantified+
func (p *Parser) parseConcatenation() (*frontend.Node, error) {
	first, err := p.parseQuantified()
	if err != nil {
		return nil, err
	}
	children := []*frontend.Node{first}
	for !p.atEOF && p.rune() != '|' && p.rune() != ')' {
		child, err := p.parseQuantified()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	if len(children) == 1 {
		return children[0], nil
	}
	return &frontend.Node{
		Kind:   frontend.KindConcat,
		Concat: frontend.Concat{Children: children},
	}, nil
}

// parseQuantified: atom quantifier?
func (p *Parser) parseQuantified() (*frontend.Node, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	if p.atEOF {
		return atom, nil
	}
	switch p.rune() {
	case '*':
		p.next()
		return &frontend.Node{Kind: frontend.KindZeroOrMore, ZeroOrMore: frontend.ZeroOrMore{Child: atom}}, nil
	case '+':
		p.next()
		return &frontend.Node{Kind: frontend.KindOneOrMore, OneOrMore: frontend.OneOrMore{Child: atom}}, nil
	case '?':
		p.next()
		return &frontend.Node{Kind: frontend.KindOptional, Optional: frontend.Optional{Child: atom}}, nil
	case '{':
		node, more, err := p.parseRepetition(atom)
		if err != nil {
			return nil, err
		}
		return node, nil
	}
	return atom, nil
}

func (p *Parser) parseAtom() (*frontend.Node, error) {
	switch p.rune() {
	case '.':
		p.next()
		return &frontend.Node{Kind: frontend.KindAny}, nil
	case '(':
		p.next() // consume '('
		node, err := p.parseAlternation()
		if err != nil {
			return nil, err
		}
		if p.atEOF || p.rune() != ')' {
			return nil, fmt.Errorf("expected closing )")
		}
		p.next() // consume ')'
		return node, nil
	case '[':
		return p.parseCharClass()
	case '\\':
		node, more, err := p.parseEscape()
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		r := p.rune()
		if isMeta(r) {
			return nil, fmt.Errorf("unescaped metacharacter %q", r)
		}
		p.next()
		return &frontend.Node{
			Kind:    frontend.KindLiteral,
			Literal: frontend.Literal{Text: string(r)},
		}, nil
	}
}

// parseEscape consumes an escaped character.
//
// It expects to be on the backslash of the escape sequence.
// After the call, we are on the rune following the escaped character.
// Returns the frontend node representing the escaped character, an indicator if more runes are available and an error.
func (p *Parser) parseEscape() (*frontend.Node, bool, error) {
	if !p.next() {
		return nil, false, errors.New("unexpected and of escape sequence")
	}
	switch p.rune() {
	case 'n':
		return dsl.Literal("\n"), p.next(), nil
	case 'r':
		return dsl.Literal("\r"), p.next(), nil
	case 't':
		return dsl.Literal("\t"), p.next(), nil
	default:
		return dsl.Literal(string(p.rune())), p.next(), nil
	}
}

// parseCharClass consumes a character class.
//
// It expects to sit on the first opening bracket.
// After the call, we are on the first rune after the closing bracket.
// It returns the frontend node, an indicator if there are more runes available and an error.
func (p *Parser) parseCharClass() (*frontend.Node, bool, error) {
	if !p.next() {
		return nil, false, errors.New("unexpected end of character class")
	}

	var negate bool
	if p.rune() == '^' {
		negate = true
		if !p.next() {
			return nil, false, errors.New("unexpected end of character class")
		}
	}

	charRanges, more, err := p.parseCharRanges()
	if err != nil {
		return nil, false, err
	}
	if !more {
		return nil, false, errors.New("unexpected end of character class")
	}

	if negate {
		return dsl.NegCharClass(charRanges), p.next(), nil
	}
	return dsl.CharClass(charRanges), p.next(), nil
}

func (p *Parser) parseCharRanges() ([]frontend.CharRange, bool, error) {
	var charRanges []frontend.CharRange

	// 0: no character seen
	// 1: first range character seen
	// 2: dash seen
	var state int
	var nextCharRange frontend.CharRange
	for {
		switch p.rune() {
		case '\\':
		case '-':
		case ']':
			return charRanges, p.next(), nil
		default:
			switch state {
			case 0:
				nextCharRange.Low = p.rune()
				nextCharRange.High = p.rune()
				state = 1
			case 1:
				charRanges = append(charRanges, nextCharRange)
				nextCharRange.Low = p.rune()
				nextCharRange.High = p.rune()
			case 2:
				nextCharRange.High = p.rune()
				state = 0
			}
		}

		if !p.next() {
			return nil, false, errors.New("unexpected end of character class")
		}
	}

	for !p.atEOF && p.rune() != ']' {
		low, err := p.parseClassChar()
		if err != nil {
			return nil, err
		}
		if !p.atEOF && p.rune() == '-' {
			p.next() // consume '-'
			if p.atEOF || p.rune() == ']' {
				// trailing dash is a literal
				charRanges = append(charRanges, frontend.CharRange{Low: low, High: low})
				charRanges = append(charRanges, frontend.CharRange{Low: '-', High: '-'})
			} else {
				high, err := p.parseClassChar()
				if err != nil {
					return nil, err
				}
				if low > high {
					return nil, fmt.Errorf("character range out of order: %q-%q", low, high)
				}
				charRanges = append(charRanges, frontend.CharRange{Low: low, High: high})
			}
		} else {
			charRanges = append(charRanges, frontend.CharRange{Low: low, High: low})
		}
	}
	return charRanges, nil
}

func (p *Parser) parseClassChar() (rune, error) {
	if p.atEOF || p.rune() == ']' {
		return 0, fmt.Errorf("unexpected end of character class")
	}
	if p.rune() != '\\' {
		r := p.rune()
		p.next()
		return r, nil
	}
	p.next() // consume '\'
	if p.atEOF {
		return 0, fmt.Errorf("unexpected end of input after \\ in character class")
	}
	r := p.rune()
	p.next()
	switch r {
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	default:
		return r, nil
	}
}

// parseRepetition consumes a repetition statement like {3}, {3,}, {,3} or {2,3}.
//
// It expects to be on the first open curly braces.
// After the call we are on the first rune following the closing curly braces.
// It returns a frontend node, an indicator if more runes are available and an error.
func (p *Parser) parseRepetition(child *frontend.Node) (*frontend.Node, bool, error) {
	if !p.next() {
		return nil, false, errors.New("unexpected end of repetition")
	}

	minRepetition := 0
	maxRepetition := math.MaxInt
	hasMin, hasComma, hasMax := false, false, false

	if '0' <= p.rune() && p.rune() <= '9' {
		hasMin = true
		repetition, more := p.parseNumber()
		if !more {
			return nil, false, errors.New("unexpected end of repetition")
		}
		minRepetition = repetition
	}

	if p.rune() == ',' {
		hasComma = true
		if !p.next() {
			return nil, false, errors.New("unexpected end of repetition")
		}
	}

	if '0' <= p.rune() && p.rune() <= '9' {
		hasMax = true
		repetition, more := p.parseNumber()
		if !more {
			return nil, false, errors.New("unexpected end of repetition")
		}
		maxRepetition = repetition
	}

	if p.rune() != '}' {
		return nil, false, errors.New("expected end of repetition")
	}

	if !hasMin && !hasComma && !hasMax {
		return nil, false, errors.New("empty repetition")
	}

	more := p.next()

	switch {
	case !hasComma:
		// {n}
		return dsl.Repetition(child, minRepetition, minRepetition), more, nil
	case hasMin && !hasMax:
		// {n,}
		switch minRepetition {
		case 0:
			return dsl.ZeroOrMore(child), more, nil
		case 1:
			return dsl.OneOrMore(child), more, nil
		default:
			exact := dsl.Repetition(child, minRepetition, minRepetition)
			rest := dsl.ZeroOrMore(child)
			return dsl.Concat(exact, rest), more, nil
		}
	case !hasMin && hasMax:
		// {,m}
		switch maxRepetition {
		case 1:
			return dsl.Optional(child), more, nil
		default:
			return dsl.Repetition(child, 0, maxRepetition), more, nil
		}
	default:
		// {n,m}
		return dsl.Repetition(child, minRepetition, maxRepetition), more, nil
	}
}

// parseNumber consumes a positive integer.
//
// It expects to be on the first digit already.
// It returns the parsed positive integer and an indicator if more runes are available.
// After the call, we are on the first rune following the number.
func (p *Parser) parseNumber() (int, bool) {
	var number int
	for '0' <= p.rune() && p.rune() <= '9' {
		number = 10*number + int(p.rune()-'0')
		if !p.next() {
			return number, false
		}
	}
	return number, true
}

func isMeta(r rune) bool {
	switch r {
	case '.', '*', '+', '?', '|', '(', ')', '[', ']', '{', '}', '^', '$':
		return true
	}
	return false
}
