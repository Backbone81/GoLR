package parser_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// seeds returns the grammar files of the repository which are small enough for fuzzing, and broken variants of a small
// grammar which lead the parser through its error recovery.
func seeds() ([][]byte, error) {
	var sources [][]byte
	paths, err := filepath.Glob("../../../../backendtest/golden/*/spec.golr")
	if err != nil {
		return nil, err
	}
	paths = append(paths,
		"../../../../../examples/calculator/calculator.golr",
		"../spec/golr.golr",
	)
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	for _, seed := range []string{
		"",
		"@scanner{A:\"a\";}@parser{@start s; s:A;}",
		"@scanner{A:\"a\" B;}@parser{@start s; s:A;}",
		"@scanner{A:;}@parser{@start s; s:A|;}",
		"@scanner{A:\"a\";}@parser{@start s; @precedence{@left:A;@right B;} s:A @name(x)|@empty;}",
		"@scanner{A:\"a\";}@parser{@start ; s:A ;; t:: ;}",
		"@scanner{A:/[a-z]+/ @skip; B:\"b\" @fragment;} // comment\n@parser{s:A B}",
		"@scanner{\"unterminated}",
		"@parser{s:A;}@scanner{A:\"a\";}",
		"\xff\x00@scanner{}",
	} {
		sources = append(sources, []byte(seed))
	}
	return sources, nil
}

// token is a token as the scanner reported it.
type token struct {
	token  parser.Token
	offset int
	length int
}

// scanAll reads the remaining tokens of the scanner and checks that each one is the part of the source it claims to
// be. Without skipping, the tokens follow each other without a gap, starting at the given offset.
func scanAll(scanner parser.TokenSource, source []byte, offset int, skipping bool) ([]token, error) {
	var tokens []token
	end := offset
	for scanner.Next() {
		if len(tokens) > len(source) {
			return nil, fmt.Errorf("more tokens than bytes in %q", source)
		}
		current := token{token: scanner.Token(), offset: scanner.ByteOffset(), length: scanner.ByteLength()}
		if current.length <= 0 || current.offset < end || (!skipping && current.offset != end) ||
			len(source) < current.offset+current.length {
			return nil, fmt.Errorf(
				"token %v at %d with length %d after %d in %q",
				current.token, current.offset, current.length, end, source,
			)
		}
		lexeme := source[current.offset : current.offset+current.length]
		if !bytes.Equal(scanner.Lexeme(), lexeme) || !bytes.Equal(scanner.Text(current.offset, current.length), lexeme) {
			return nil, fmt.Errorf(
				"token %v at %d has lexeme %q and text %q instead of %q",
				current.token, current.offset, scanner.Lexeme(), scanner.Text(current.offset, current.length), lexeme,
			)
		}
		if skipping && current.token.IsSkipped() {
			return nil, fmt.Errorf("skipped token %v at %d", current.token, current.offset)
		}
		tokens = append(tokens, current)
		end = current.offset + current.length
	}
	if scanner.Token() != parser.EndToken || scanner.ByteOffset() != len(source) || scanner.ByteLength() != 0 {
		return nil, fmt.Errorf(
			"scan ends with token %v at %d with length %d in %q",
			scanner.Token(), scanner.ByteOffset(), scanner.ByteLength(), source,
		)
	}
	return tokens, nil
}

// FuzzScanner scans arbitrary bytes. The tokens have to cover the source, scanning again from any token reproduces
// the tokens from there on, the skipper hands on the tokens which are not skipped, and positions count lines and
// characters.
func FuzzScanner(f *testing.F) {
	sources, err := seeds()
	if err != nil {
		f.Fatal(err)
	}
	for _, source := range sources {
		f.Add(source)
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		scanner := parser.NewScanner(source, "fuzz.golr")
		tokens, err := scanAll(scanner, source, 0, false)
		if err != nil {
			t.Fatal(err)
		}

		// Every token boundary when there are few, a spread of them otherwise, so the check stays linear.
		step := max(1, len(tokens)/8)
		for idx := 0; idx < len(tokens); idx += step {
			scanner.Reset(source, tokens[idx].offset)
			rescanned, err := scanAll(scanner, source, tokens[idx].offset, false)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(rescanned) != fmt.Sprint(tokens[idx:]) {
				t.Fatalf("scanning from %d gives %v instead of %v", tokens[idx].offset, rescanned, tokens[idx:])
			}
		}

		var unskipped []token
		for _, current := range tokens {
			if !current.token.IsSkipped() {
				unskipped = append(unskipped, current)
			}
		}
		skipper := parser.NewTokenSkipper(parser.NewScanner(source, "fuzz.golr"))
		skipped, err := scanAll(skipper, source, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(skipped) != fmt.Sprint(unskipped) {
			t.Fatalf("skipper gives %v instead of %v", skipped, unskipped)
		}

		line, column := 1, 1
		checked := 0
		for offset := 0; offset <= len(source); offset++ {
			if checked < len(tokens) && tokens[checked].offset == offset || offset == len(source) {
				checked++
				want := parser.Position{FilePath: "fuzz.golr", ByteOffset: offset, Line: line, Column: column}
				if got := scanner.Position(offset); got != want {
					t.Fatalf("position of %d is %+v instead of %+v", offset, got, want)
				}
			}
			if offset == len(source) {
				break
			}
			switch {
			case source[offset] == '\n':
				line, column = line+1, 1
			case source[offset]&0xc0 != 0x80:
				column++
			}
		}
		if scanner.Position(-1) != scanner.Position(0) || scanner.Position(len(source)+1) != scanner.Position(len(source)) {
			t.Fatalf("positions outside of the source are not clamped")
		}
	})
}

// FuzzParser parses arbitrary bytes. Every error has to be a syntax error at a token, a tree has to come with every
// parse which reached the end and only then, its nodes have to nest inside each other and cover the tokens, and the
// same input has to parse the same with a parser used before and with tracing enabled.
func FuzzParser(f *testing.F) {
	sources, err := seeds()
	if err != nil {
		f.Fatal(err)
	}
	for _, source := range sources {
		f.Add(source)
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		scanner := parser.NewTokenSkipper(parser.NewScanner(source, "fuzz.golr"))
		tokens, err := scanAll(scanner, source, 0, true)
		if err != nil {
			t.Fatal(err)
		}

		golrParser := parser.NewParser()
		var traced strings.Builder
		golrParser.Trace = func(line string) {
			traced.WriteString(line)
		}
		scanner.Reset(source, 0)
		root, parseErr := golrParser.Parse(scanner)
		if err := checkParse(scanner, source, tokens, root, parseErr); err != nil {
			t.Fatal(err)
		}
		if traced.Len() == 0 {
			t.Fatalf("no trace")
		}

		golrParser.Trace = nil
		scanner.Reset(source, 0)
		againRoot, againErr := golrParser.Parse(scanner)
		if err := compareParses(root, parseErr, againRoot, againErr); err != nil {
			t.Fatal(err)
		}
	})
}

// checkParse checks the result of a parse.
func checkParse(
	scanner parser.TokenSource,
	source []byte,
	tokens []token,
	root parser.Node,
	parseErr error,
) error {
	errorCount := 0
	if parseErr != nil {
		joined, ok := parseErr.(interface{ Unwrap() []error })
		if !ok {
			return fmt.Errorf("error is not joined: %w", parseErr)
		}
		for _, single := range joined.Unwrap() {
			var parseError *parser.Error
			if !errors.As(single, &parseError) || !errors.Is(single, parser.ErrSyntax) ||
				errors.Is(single, parser.ErrInternal) {
				return fmt.Errorf("error is no syntax error: %w", single)
			}
			position := scanner.Position(parseError.ByteOffset)
			if parseError.ByteOffset != position.ByteOffset || parseError.Line != position.Line ||
				parseError.Column != position.Column || parseError.FilePath != position.FilePath ||
				!bytes.Equal(parseError.Lexeme, scanner.Text(parseError.ByteOffset, len(parseError.Lexeme))) {
				return fmt.Errorf(
					"error %w at %d does not match position %+v", single, parseError.ByteOffset, position,
				)
			}
			errorCount++
		}
	}

	if root.Symbol == 0 && root.Children == nil {
		if parseErr == nil {
			return errors.New("no tree and no error")
		}
		return nil
	}
	if nonterminal, ok := root.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalFile {
		return fmt.Errorf("root is %v", root.Symbol)
	}

	checker := treeChecker{source: source, tokenByOffset: map[int]token{}}
	for _, current := range tokens {
		checker.tokenByOffset[current.offset] = current
	}
	if err := checker.check(root); err != nil {
		return err
	}

	if (errorCount == 0) != (checker.errorNodes == 0) {
		return fmt.Errorf("%d errors with %d error nodes", errorCount, checker.errorNodes)
	}
	// The leaves are ordered like the tokens, so one pass finds the leaf for every token.
	leafIdx := 0
	for _, current := range tokens {
		for leafIdx < len(checker.leaves) && !covers(checker.leaves[leafIdx], current) &&
			checker.leaves[leafIdx].ByteOffset <= current.offset {
			leafIdx++
		}
		if leafIdx == len(checker.leaves) || !covers(checker.leaves[leafIdx], current) {
			return fmt.Errorf(
				"token %v at %d is neither in the tree nor in an error node", current.token, current.offset,
			)
		}
	}
	return nil
}

// compareParses returns an error showing both parses if parsing again gave a different tree or error than the first
// parse.
func compareParses(root parser.Node, parseErr error, againRoot parser.Node, againErr error) error {
	if reflect.DeepEqual(againRoot, root) && reflect.DeepEqual(againErr, parseErr) {
		return nil
	}
	tree, err := json.MarshalIndent(root, "", " ") //nolint:musttag
	if err != nil {
		return err
	}
	againTree, err := json.MarshalIndent(againRoot, "", " ") //nolint:musttag
	if err != nil {
		return err
	}
	return fmt.Errorf("parsing again gives\n%s\n%w\ninstead of\n%s\n%w", againTree, againErr, tree, parseErr)
}

// treeChecker checks the nodes of a parse tree.
type treeChecker struct {
	source        []byte
	tokenByOffset map[int]token

	// leaves are the terminal nodes in the order of the tree, the error nodes included.
	leaves     []parser.Node
	errorNodes int
}

// check checks the given node and the nodes below it.
func (c *treeChecker) check(node parser.Node) error {
	if node.ByteOffset < 0 || node.ByteLength < 0 || len(c.source) < node.ByteOffset+node.ByteLength {
		return fmt.Errorf(
			"%v at %d with length %d is outside of %d bytes",
			node.Symbol, node.ByteOffset, node.ByteLength, len(c.source),
		)
	}
	if terminal, ok := node.Symbol.Terminal(); ok {
		if len(node.Children) != 0 || node.Production != parser.NoProduction {
			return fmt.Errorf("terminal %v at %d has children or a production", terminal, node.ByteOffset)
		}
		if terminal == parser.ErrorToken {
			c.errorNodes++
		} else if current, ok := c.tokenByOffset[node.ByteOffset]; !ok || current.token != terminal ||
			current.length != node.ByteLength {
			return fmt.Errorf(
				"terminal %v at %d with length %d is no token", terminal, node.ByteOffset, node.ByteLength,
			)
		}
		c.leaves = append(c.leaves, node)
		return nil
	}

	if node.Production == parser.NoProduction {
		return fmt.Errorf("nonterminal %v at %d has no production", node.Symbol, node.ByteOffset)
	}
	if len(node.Children) == 0 {
		if node.ByteLength != 0 {
			return fmt.Errorf("empty %v at %d has length %d", node.Symbol, node.ByteOffset, node.ByteLength)
		}
		return nil
	}
	first, last := node.Children[0], node.Children[len(node.Children)-1]
	if node.ByteOffset != first.ByteOffset || node.ByteOffset+node.ByteLength != last.ByteOffset+last.ByteLength {
		return fmt.Errorf("%v at %d with length %d does not span its children", node.Symbol, node.ByteOffset,
			node.ByteLength)
	}
	for idx, child := range node.Children {
		if idx != 0 {
			previous := node.Children[idx-1]
			if child.ByteOffset < previous.ByteOffset+previous.ByteLength {
				return fmt.Errorf("children %d and %d of %v at %d overlap", idx-1, idx, node.Symbol, node.ByteOffset)
			}
		}
		if err := c.check(child); err != nil {
			return err
		}
	}
	return nil
}

// covers reports whether the leaf is the token or an error node the token lies inside of.
func covers(leaf parser.Node, current token) bool {
	if terminal, _ := leaf.Symbol.Terminal(); terminal == parser.ErrorToken {
		return leaf.ByteOffset <= current.offset && current.offset+current.length <= leaf.ByteOffset+leaf.ByteLength
	}
	return leaf.ByteOffset == current.offset
}

// tokenSequence is a token source which hands on a token of one byte for every byte of its source, so the fuzzer
// chooses the tokens directly instead of the bytes the scanner makes tokens of.
type tokenSequence struct {
	source []byte
	offset int
	length int
}

// tokenSequence implements TokenSource.
var _ parser.TokenSource = (*tokenSequence)(nil)

// sequenceTokens are the tokens a byte of the source of tokenSequence stands for: every token a parser can be handed
// besides the end of input.
var sequenceTokens = newSequenceTokens()

// newSequenceTokens returns the tokens for sequenceTokens.
func newSequenceTokens() []parser.Token {
	var tokens []parser.Token
	for current := parser.InvalidToken; current <= parser.TokenString; current++ {
		if current != parser.EndToken && current != parser.ErrorToken && !current.IsSkipped() {
			tokens = append(tokens, current)
		}
	}
	return tokens
}

func (s *tokenSequence) Token() parser.Token {
	if s.offset == len(s.source) {
		return parser.EndToken
	}
	return sequenceTokens[int(s.source[s.offset])%len(sequenceTokens)]
}

func (s *tokenSequence) ByteOffset() int {
	return s.offset
}

func (s *tokenSequence) ByteLength() int {
	return s.length
}

func (s *tokenSequence) Lexeme() []byte {
	return s.source[s.offset : s.offset+s.length]
}

func (s *tokenSequence) Position(byteOffset int) parser.Position {
	byteOffset = min(max(byteOffset, 0), len(s.source))
	return parser.Position{FilePath: "tokens", ByteOffset: byteOffset, Line: 1, Column: byteOffset + 1}
}

func (s *tokenSequence) Text(byteOffset int, byteLength int) []byte {
	startIdx := min(max(byteOffset, 0), len(s.source))
	return s.source[startIdx:min(max(startIdx, startIdx+byteLength), len(s.source))]
}

func (s *tokenSequence) FilePath() string {
	return "tokens"
}

func (s *tokenSequence) Reset(source []byte, offset int) {
	s.source = source
	s.offset = min(offset, len(source))
	s.length = 0
}

func (s *tokenSequence) Next() bool {
	s.offset += s.length
	s.length = min(1, len(s.source)-s.offset)
	return s.length != 0
}

// FuzzParserTokens parses arbitrary sequences of tokens, which reaches states of the parser no source reaches through
// the scanner. The parse has to hold to the same as in FuzzParser.
func FuzzParserTokens(f *testing.F) {
	// The seeds are the tokens the scanner makes of the seed sources.
	sources, err := seeds()
	if err != nil {
		f.Fatal(err)
	}
	for _, source := range sources {
		scanner := parser.NewTokenSkipper(parser.NewScanner(source, "seed.golr"))
		var sequence []byte
		for scanner.Next() {
			sequence = append(sequence, byte(slices.Index(sequenceTokens, scanner.Token())))
		}
		f.Add(sequence)
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		sequence := &tokenSequence{}
		sequence.Reset(source, 0)
		var tokens []token
		for sequence.Next() {
			tokens = append(tokens, token{token: sequence.Token(), offset: sequence.ByteOffset(), length: 1})
		}

		golrParser := parser.NewParser()
		sequence.Reset(source, 0)
		root, parseErr := golrParser.Parse(sequence)
		if err := checkParse(sequence, source, tokens, root, parseErr); err != nil {
			t.Fatal(err)
		}

		sequence.Reset(source, 0)
		againRoot, againErr := golrParser.Parse(sequence)
		if err := compareParses(root, parseErr, againRoot, againErr); err != nil {
			t.Fatal(err)
		}
	})
}
