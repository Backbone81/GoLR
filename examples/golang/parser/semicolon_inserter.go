package parser

// SemicolonInserter wraps a Scanner and inserts synthetic TokenSemicolon tokens
// as specified by https://go.dev/ref/spec#Semicolons.
//
// A semicolon is inserted between two tokens whenever:
//   - the line number increases (the previous token was the last on its line), and
//   - the previous token is one of the trigger tokens listed in the spec.
//
// A trailing semicolon is also inserted at end of file if the last token is a trigger.
type SemicolonInserter struct {
	Scanner *TokenSkipper

	insertSemicolon bool
	bufferedTokens  []Token
	bufferedResult  bool
}

func (s *SemicolonInserter) Reset(source []byte, offset int) {
	s.Scanner.Reset(source, offset)
	s.insertSemicolon = false
	s.bufferedTokens = s.bufferedTokens[:0]
}

func (s *SemicolonInserter) Next() bool {
	previousLine := s.Scanner.Line()
	if s.Scanner.Token() == TokenStringLit && s.Scanner.Lexeme()[0] == '`' {
		for _, b := range s.Scanner.Lexeme() {
			if b != '\n' {
				continue
			}
			previousLine++
		}
	}

	var result bool
	if len(s.bufferedTokens) > 0 {
		s.bufferedTokens = s.bufferedTokens[:0]
		result = s.bufferedResult
	} else {
		result = s.Scanner.Next()
	}

	if !result && s.insertSemicolon {
		s.insertSemicolon = false
		s.bufferedTokens = append(s.bufferedTokens, TokenSemicolon)
		s.bufferedResult = result
		return true
	}

	switch {
	case s.insertSemicolon && previousLine < s.Scanner.Line():
		s.insertSemicolon = false
		s.bufferedTokens = append(s.bufferedTokens, TokenSemicolon)
		s.bufferedResult = result
		return true
	case isSemicolonTrigger(s.Scanner.Token()):
		s.insertSemicolon = true
	default:
		s.insertSemicolon = false
	}

	return result
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
	return s.Scanner.ByteOffset()
}

func (s *SemicolonInserter) Line() int {
	return s.Scanner.Line()
}

func (s *SemicolonInserter) Column() int {
	return s.Scanner.Column()
}

func (s *SemicolonInserter) Lexeme() []byte {
	return s.Scanner.Lexeme()
}

func (s *SemicolonInserter) FilePath() string {
	return s.Scanner.FilePath()
}
