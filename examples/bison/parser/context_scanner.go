package parser

// ContextScanner is responsible for collecting blocks of code into one single token.
type ContextScanner struct {
	Scanner             *Scanner
	percentPercentCount int
}

// ContextScanner implements ParserScanner
var _ ParserScanner = (*ContextScanner)(nil)

func (c *ContextScanner) Err() error {
	return c.Scanner.Err()
}

func (c *ContextScanner) Token() Token {
	return c.Scanner.Token()
}

func (c *ContextScanner) ByteOffset() int {
	return c.Scanner.ByteOffset()
}

func (c *ContextScanner) Line() int {
	return c.Scanner.Line()
}

func (c *ContextScanner) Column() int {
	return c.Scanner.Column()
}

func (c *ContextScanner) Lexeme() []byte {
	return c.Scanner.Lexeme()
}

func (c *ContextScanner) Next() bool {
	if c.percentPercentCount == 2 {
		c.Scanner.ReadEpilogue()

		// We want to prevent the next call to again read the epilogue.
		c.percentPercentCount++
		return true
	}

	result := c.Scanner.Next()
	switch c.Token() {
	case TokenPercentPercent:
		c.percentPercentCount++
	case TokenTagStart:
		c.Scanner.ReadTag()
	case TokenPrologueStart:
		c.Scanner.ReadPrologue()
	case TokenBracedCodeStart:
		c.Scanner.ReadBracedCode()
	case TokenBracedPredicateStart:
		c.Scanner.ReadBracedPredicate()
	}
	return result
}

func (c *ContextScanner) FilePath() string {
	return c.Scanner.FilePath()
}
