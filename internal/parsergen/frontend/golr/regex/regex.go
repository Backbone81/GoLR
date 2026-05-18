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
}

func (p *Parser) Parse(input []byte) (*frontend.Node, error) {
	if len(input) < 2 || input[0] != '/' || input[len(input)-1] != '/' {
		return nil, fmt.Errorf("expected / around regular expression")
	}
	p.input = input[1 : len(input)-1]
	p.pos = 0
	p.currRune = 0

	if !p.next() {
		return nil, fmt.Errorf("unexpected empty regular expression")
	}

	node, more, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}
	if more {
		return nil, fmt.Errorf("unexpected character %q", p.rune())
	}
	return node, nil
}

func (p *Parser) next() bool {
	if p.pos >= len(p.input) {
		return false
	}
	if p.input[p.pos] < utf8.RuneSelf {
		p.currRune = rune(p.input[p.pos])
		p.pos++
	} else {
		nextRune, size := utf8.DecodeRune(p.input[p.pos:])
		p.currRune = nextRune
		p.pos += size
	}
	return true
}

func (p *Parser) rune() rune {
	return p.currRune
}

func (p *Parser) parseAlternation() (*frontend.Node, bool, error) {
	var children []*frontend.Node

	more := true
	for {
		child, localMore, err := p.parseConcatenation()
		if err != nil {
			return nil, false, err
		}
		more = localMore

		children = append(children, child)

		if !more || p.rune() != '|' {
			break
		}
		if !p.next() {
			return nil, false, errors.New("unexpected end of alternation")
		}
	}
	if len(children) == 1 {
		return children[0], more, nil
	}
	return dsl.Or(children...), more, nil
}

func (p *Parser) parseConcatenation() (*frontend.Node, bool, error) {
	var children []*frontend.Node

	more := true
	for p.rune() != '|' && p.rune() != ')' && more {
		child, localMore, err := p.parseQuantified()
		if err != nil {
			return nil, false, err
		}
		more = localMore
		children = append(children, child)
	}

	if len(children) == 0 {
		return nil, false, errors.New("unexpected end of concatenation")
	}
	if len(children) == 1 {
		return children[0], more, nil
	}
	return dsl.Concat(children...), more, nil
}

func (p *Parser) parseQuantified() (*frontend.Node, bool, error) {
	atom, more, err := p.parseAtom()
	if err != nil {
		return nil, false, err
	}
	if !more {
		return atom, false, nil
	}

	switch p.rune() {
	case '*':
		return dsl.ZeroOrMore(atom), p.next(), nil
	case '+':
		return dsl.OneOrMore(atom), p.next(), nil
	case '?':
		return dsl.Optional(atom), p.next(), nil
	case '{':
		node, more, err := p.parseRepetition(atom)
		if err != nil {
			return nil, false, err
		}
		return node, more, nil
	default:
		return atom, true, nil
	}
}

func (p *Parser) parseAtom() (*frontend.Node, bool, error) {
	switch p.rune() {
	case '.':
		return dsl.Any(), p.next(), nil
	case '(':
		if !p.next() {
			return nil, false, errors.New("unexpected end of group")
		}

		node, more, err := p.parseAlternation()
		if err != nil {
			return nil, false, err
		}
		if !more || p.rune() != ')' {
			return nil, false, fmt.Errorf("expected end of group")
		}

		return node, p.next(), nil
	case '[':
		return p.parseCharClass()
	case '\\':
		escapedChar, more, err := p.parseEscapeSequence()
		if err != nil {
			return nil, false, err
		}
		return dsl.Literal(string(escapedChar)), more, nil
	case '*', '+', '?', '{', '|', ')', ']', '}', '^', '$':
		return nil, false, fmt.Errorf("unescaped metacharacter %q", p.rune())
	default:
		return dsl.Literal(string(p.rune())), p.next(), nil
	}
}

// parseEscapeSequence consumes an escaped character.
//
// It expects to be on the backslash of the escape sequence.
// After the call, we are on the rune following the escaped character.
// Returns the unescaped character, an indicator if more runes are available and an error.
func (p *Parser) parseEscapeSequence() (rune, bool, error) {
	// We are on '\' right now.
	if !p.next() {
		return 0, false, errors.New("unexpected end of escape sequence")
	}

	switch p.rune() {
	case 'n':
		return '\n', p.next(), nil
	case 'r':
		return '\r', p.next(), nil
	case 't':
		return '\t', p.next(), nil
	default:
		return p.rune(), p.next(), nil
	}
}

// parseCharClass consumes a character class.
//
// It expects to sit on the first opening bracket.
// After the call, we are on the first rune after the closing bracket.
// It returns the frontend node, an indicator if there are more runes available and an error.
func (p *Parser) parseCharClass() (*frontend.Node, bool, error) {
	// We are on '[' right now.
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

	if negate {
		return dsl.NegCharClass(charRanges...), more, nil
	}
	return dsl.CharClass(charRanges...), more, nil
}

func (p *Parser) parseCharRanges() ([]frontend.CharRange, bool, error) {
	var charRanges []frontend.CharRange

	// 0: no character seen
	// 1: first range character seen
	// 2: dash seen
	var state int
	var nextCharRange frontend.CharRange
	for {
		switch state {
		case 0:
			// no character seen
			switch p.rune() {
			case '\\':
				escapedChar, more, err := p.parseEscapeSequence()
				if err != nil {
					return nil, false, err
				}
				if !more {
					return nil, false, errors.New("unexpected end of character class")
				}
				nextCharRange.Low = escapedChar
				nextCharRange.High = escapedChar
				state = 1

				// We need to prevent a double advance at the end of the loop.
				continue
			case ']':
				return charRanges, p.next(), nil
			default:
				nextCharRange.Low = p.rune()
				nextCharRange.High = p.rune()
				state = 1
			}
		case 1:
			// first range character seen
			switch p.rune() {
			case '\\':
				charRanges = append(charRanges, nextCharRange)
				escapedChar, more, err := p.parseEscapeSequence()
				if err != nil {
					return nil, false, err
				}
				if !more {
					return nil, false, errors.New("unexpected end of character class")
				}
				nextCharRange.Low = escapedChar
				nextCharRange.High = escapedChar

				// We need to prevent a double advance at the end of the loop.
				continue
			case '-':
				state = 2
			case ']':
				charRanges = append(charRanges, nextCharRange)
				return charRanges, p.next(), nil
			default:
				charRanges = append(charRanges, nextCharRange)
				nextCharRange.Low = p.rune()
				nextCharRange.High = p.rune()
			}
		case 2:
			// dash seen
			switch p.rune() {
			case '\\':
				escapedChar, more, err := p.parseEscapeSequence()
				if err != nil {
					return nil, false, err
				}
				if !more {
					return nil, false, errors.New("unexpected end of character class")
				}
				nextCharRange.High = escapedChar
				charRanges = append(charRanges, nextCharRange)
				state = 0

				// We need to prevent a double advance at the end of the loop.
				continue
			case ']':
				charRanges = append(charRanges, nextCharRange)
				charRanges = append(charRanges, dsl.CharRange('-', '-'))
				return charRanges, p.next(), nil
			default:
				nextCharRange.High = p.rune()
				charRanges = append(charRanges, nextCharRange)
				state = 0
			}
		default:
			return nil, false, fmt.Errorf("unexpected character class state: %d", state)
		}

		if !p.next() {
			return nil, false, errors.New("unexpected end of character class")
		}
	}
}

// parseRepetition consumes a repetition statement like {3}, {3,}, {,3} or {2,3}.
//
// It expects to be on the first open curly braces.
// After the call we are on the first rune following the closing curly braces.
// It returns a frontend node, an indicator if more runes are available and an error.
func (p *Parser) parseRepetition(child *frontend.Node) (*frontend.Node, bool, error) {
	// We are on '{' right now.
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

	if !hasMin && !hasMax {
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
