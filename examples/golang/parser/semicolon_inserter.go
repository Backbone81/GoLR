package parser

import "bytes"

// SemicolonInserter wraps a Scanner and inserts synthetic TokenSemicolon tokens
// as specified by https://go.dev/ref/spec#Semicolons.
//
// A semicolon is inserted between two tokens whenever:
//   - a line break sits between them (the previous token was the last on its line), and
//   - the previous token is one of the trigger tokens listed in the spec.
//
// A trailing semicolon is also inserted at end of file if the last token is a trigger.
//
// An inserted semicolon has an empty lexeme and the offset of the end of the token in front of it.
type SemicolonInserter struct {
	Scanner *TokenSkipper

	insertSemicolon bool
	bufferedTokens  []Token
	bufferedResult  bool

	bufferedByteOffset int
}

func (s *SemicolonInserter) Reset(source []byte, offset int) {
	s.Scanner.Reset(source, offset)
	s.insertSemicolon = false
	s.bufferedTokens = s.bufferedTokens[:0]
}

func (s *SemicolonInserter) Next() bool {
	// A semicolon inserted after the current token goes at its end.
	previousEnd := s.Scanner.ByteOffset() + len(s.Scanner.Lexeme())

	var result bool
	if len(s.bufferedTokens) > 0 {
		s.bufferedTokens = s.bufferedTokens[:0]
		result = s.bufferedResult
	} else {
		result = s.Scanner.Next()
	}

	if !result && s.insertSemicolon {
		s.insertSemicolon = false
		s.bufferSemicolon(previousEnd, result)
		return true
	}

	switch {
	case s.insertSemicolon && s.crossesLine(previousEnd):
		s.insertSemicolon = false
		s.bufferSemicolon(previousEnd, result)
		return true
	case isSemicolonTrigger(s.Scanner.Token()):
		s.insertSemicolon = true
	default:
		s.insertSemicolon = false
	}

	return result
}

// bufferSemicolon holds a semicolon back, so that the next call hands on the token the scanner has already moved to.
func (s *SemicolonInserter) bufferSemicolon(byteOffset int, result bool) {
	s.bufferedTokens = append(s.bufferedTokens, TokenSemicolon)
	s.bufferedByteOffset = byteOffset
	s.bufferedResult = result
}

// crossesLine reports whether a line break sits between the end of the previous token and the start of the current one.
// Everything in between was skipped, so a line feed anywhere in that gap means the previous token was the last on its
// line.
func (s *SemicolonInserter) crossesLine(previousEnd int) bool {
	gap := s.Scanner.Text(previousEnd, s.Scanner.ByteOffset()-previousEnd)
	return bytes.IndexByte(gap, '\n') >= 0
}

func isSemicolonTrigger(tok Token) bool {
	//nolint:exhaustive // We are only interested in these few tokens. No need to list all.
	switch tok {
	case TokenIdentifier,
		TokenIntLit, TokenFloatLit, TokenImaginaryLit, TokenRuneLit, TokenStringLit,
		TokenBreak, TokenContinue, TokenFallthrough, TokenReturn,
		TokenIncrement, TokenDecrement,
		TokenRightParen, TokenRightBracket, TokenRightBrace:
		return true
	}
	return false
}

func (s *SemicolonInserter) Token() Token {
	if len(s.bufferedTokens) > 0 {
		return s.bufferedTokens[0]
	}
	return s.Scanner.Token()
}

func (s *SemicolonInserter) ByteOffset() int {
	if len(s.bufferedTokens) > 0 {
		return s.bufferedByteOffset
	}
	return s.Scanner.ByteOffset()
}

func (s *SemicolonInserter) Line() int {
	return s.Position(s.ByteOffset()).Line
}

func (s *SemicolonInserter) Column() int {
	return s.Position(s.ByteOffset()).Column
}

func (s *SemicolonInserter) Lexeme() []byte {
	if len(s.bufferedTokens) > 0 {
		return nil
	}
	return s.Scanner.Lexeme()
}

func (s *SemicolonInserter) Position(byteOffset int) Position {
	return s.Scanner.Position(byteOffset)
}

func (s *SemicolonInserter) Text(byteOffset int, byteLength int) []byte {
	return s.Scanner.Text(byteOffset, byteLength)
}

func (s *SemicolonInserter) FilePath() string {
	return s.Scanner.FilePath()
}
