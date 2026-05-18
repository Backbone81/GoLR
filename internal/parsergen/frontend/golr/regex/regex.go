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
		return nil, errors.New("expected / around regular expression")
	}
	p.input = input[1 : len(input)-1]
	p.pos = 0
	p.currRune = 0

	if !p.next() {
		return nil, errors.New("unexpected empty regular expression")
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

// parseAlternation consumes an alternation of one or more concatenations separated by |.
//
// It expects to be on the first rune of the first concatenation.
// After the call, if more is true, p.rune() is on the first rune following the alternation,
// which is ) when the alternation is inside a group.
// It returns the frontend node, an indicator if more runes are available and an error.
func (p *Parser) parseAlternation() (*frontend.Node, bool, error) {
	var children []*frontend.Node

	var more bool
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

// parseConcatenation consumes a concatenation of one or more quantified atoms.
//
// It expects to be on the first rune of the first atom.
// After the call, if more is true, p.rune() is on the first rune that terminated the concatenation,
// which is | to continue an alternation or ) to close a group.
// It returns the frontend node, an indicator if more runes are available and an error.
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
	children = p.mergeLiterals(children)
	if len(children) == 1 {
		return children[0], more, nil
	}
	return dsl.Concat(children...), more, nil
}

// parseQuantified consumes a quantified atom.
//
// It expects to be on the first rune of the atom.
// After the call, if more is true, p.rune() is on the first rune following the quantified atom.
// It returns the frontend node, an indicator if more runes are available and an error.
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

// parseAtom consumes a single atom: a literal character, escape sequence, character class, group, or the any operator.
//
// It expects to be on the first rune of the atom.
// After the call, if more is true, p.rune() is on the first rune following the atom.
// It returns the frontend node, an indicator if more runes are available and an error.
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
			return nil, false, errors.New("expected end of group")
		}

		return node, p.next(), nil
	case '[':
		return p.parseCharClass()
	case '\\':
		escapedChar, more, err := p.parseEscapeSequence()
		if err != nil {
			return nil, false, err
		}

		switch escapedChar {
		case 'd', 'w', 's':
			return dsl.CharClass(p.shorthandCharRanges(escapedChar)...), more, nil
		case 'D', 'W', 'S':
			return dsl.NegCharClass(p.shorthandCharRanges(escapedChar)...), more, nil
		default:
			return dsl.Literal(string(escapedChar)), more, nil
		}
	case '*', '+', '?', '{', '|', ')', ']', '}', '^', '$':
		return nil, false, fmt.Errorf("unescaped metacharacter %q", p.rune())
	default:
		return dsl.Literal(string(p.rune())), p.next(), nil
	}
}

// shorthandCharRanges creates the relevant character ranges for a given character class shorthand. If an unknown
// shorthand is given, nil is returned.
func (p *Parser) shorthandCharRanges(r rune) []frontend.CharRange {
	switch r {
	case 'd', 'D':
		return []frontend.CharRange{dsl.CharRange('0', '9')}
	case 'w', 'W':
		return []frontend.CharRange{
			dsl.CharRange('A', 'Z'),
			dsl.CharRange('a', 'z'),
			dsl.CharRange('0', '9'),
			dsl.CharRange('_', '_'),
		}
	case 's', 'S':
		return []frontend.CharRange{
			dsl.CharRange('\t', '\t'),
			dsl.CharRange('\n', '\n'),
			dsl.CharRange('\r', '\r'),
			dsl.CharRange(' ', ' '),
			dsl.CharRange('\f', '\f'),
			dsl.CharRange('\v', '\v'),
		}
	default:
		return nil
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
	case 'v':
		return '\v', p.next(), nil
	case 'f':
		return '\f', p.next(), nil
	case '0':
		return 0, p.next(), nil
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

// parseCharClassChar consumes a single character inside a character class, handling escape sequences.
//
// It expects to be on the first rune of the character, which is either \ for an escape sequence or a literal rune.
// After the call, if more is true, p.rune() is on the first rune following the character.
// It returns the unescaped character, an indicator if the character is an escaped one, an indicator if more runes
// are available and an error.
func (p *Parser) parseCharClassChar() (rune, bool, bool, error) {
	if p.rune() == '\\' {
		r, more, err := p.parseEscapeSequence()
		return r, true, more, err
	}
	r := p.rune()
	return r, false, p.next(), nil
}

// parseCharRanges consumes the character ranges inside a character class until the closing bracket.
//
// It expects to be on the first rune of the first character range, after [ and an optional ^.
// After the call, if more is true, p.rune() is on the first rune following the closing ].
// A trailing - before ] is treated as a literal - character.
// It returns the character ranges, an indicator if more runes are available and an error.
func (p *Parser) parseCharRanges() ([]frontend.CharRange, bool, error) {
	var charRanges []frontend.CharRange
	for p.rune() != ']' {
		low, escaped, more, err := p.parseCharClassChar()
		if err != nil {
			return nil, false, err
		}
		if !more {
			return nil, false, errors.New("unexpected end of character class")
		}
		if escaped {
			switch low {
			case 'd', 'w', 's':
				charRanges = append(charRanges, p.shorthandCharRanges(low)...)
				continue
			case 'D', 'W', 'S':
				return nil, false, fmt.Errorf("unsupported character class shorthand %q", low)
			}
		}
		if p.rune() != '-' {
			charRanges = append(charRanges, dsl.CharRange(low, low))
			continue
		}

		if !p.next() {
			return nil, false, errors.New("unexpected end of character class")
		}
		if p.rune() == ']' {
			charRanges = append(charRanges, dsl.CharRange(low, low))
			charRanges = append(charRanges, dsl.CharRange('-', '-'))
			break
		}

		high, _, more, err := p.parseCharClassChar()
		if err != nil {
			return nil, false, err
		}
		if !more {
			return nil, false, errors.New("unexpected end of character class")
		}
		if low > high {
			return nil, false, fmt.Errorf("invalid character range order: %q-%q", low, high)
		}
		charRanges = append(charRanges, dsl.CharRange(low, high))
	}
	return charRanges, p.next(), nil
}

// parseRepetition consumes a repetition statement like {3}, {3,}, {,3} or {2,3}.
//
// It expects to be on the first open curly braces.
// After the call we are on the first rune following the closing curly braces.
// It returns a frontend node, an indicator if more runes are available and an error.
//
//nolint:cyclop,funlen // This is difficult to simplify.
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
		if maxRepetition < minRepetition {
			return nil, false, fmt.Errorf("wrong order for repetition %d-%d", minRepetition, maxRepetition)
		}
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
		number = min(number, math.MaxUint16)
		if !p.next() {
			return number, false
		}
	}
	return number, true
}

// mergeLiterals is a helper function which collapses multiple successive literals together. This makes it easier for
// tests and debugging, because /foo/ is one literal "foo" instead of cone concatenation with three literals "f", "o"
// and "o".
func (p *Parser) mergeLiterals(children []*frontend.Node) []*frontend.Node {
	result := make([]*frontend.Node, 0, len(children))
	for _, child := range children {
		if len(result) > 0 &&
			result[len(result)-1].Kind == frontend.KindLiteral &&
			child.Kind == frontend.KindLiteral {
			result[len(result)-1].Literal.Text += child.Literal.Text
			continue
		}
		result = append(result, child)
	}
	return result
}
