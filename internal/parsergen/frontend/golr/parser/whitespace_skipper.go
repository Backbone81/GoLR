package parser

// WhitespaceSkipper provides a scanner which is skipping whitespaces and comments. Use this to provide only the
// relevant tokens to the parser.
type WhitespaceSkipper struct {
	Scanner ParserScanner
}

// WhitespaceSkipper implements ParserScanner.
var _ ParserScanner = (*WhitespaceSkipper)(nil)

func (s *WhitespaceSkipper) Token() Token {
	return s.Scanner.Token()
}

func (s *WhitespaceSkipper) ByteOffset() int {
	return s.Scanner.ByteOffset()
}

func (s *WhitespaceSkipper) Line() int {
	return s.Scanner.Line()
}

func (s *WhitespaceSkipper) Column() int {
	return s.Scanner.Column()
}

func (s *WhitespaceSkipper) Lexeme() []byte {
	return s.Scanner.Lexeme()
}

func (s *WhitespaceSkipper) Next() bool {
	for {
		tokenAvailable := s.Scanner.Next()
		if !tokenAvailable {
			return false
		}
		if s.Scanner.Token() == TokenWhitespace || s.Scanner.Token() == TokenComment {
			continue
		}
		return true
	}
}

func (s *WhitespaceSkipper) FilePath() string {
	return s.Scanner.FilePath()
}
