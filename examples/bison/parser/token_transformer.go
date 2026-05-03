package parser

import "slices"

// TokenTransformer provides a transformer which modifies the returned token to be of a different type when some other
// token comes next. This is needed because the Bison grammar has some situations which require to look at the following
// token to decide the current one (like ID_COLON is generated for an ID which is followed by a COLON).
type TokenTransformer struct {
	Scanner             ParserScanner
	tokenQueue          []TokenSnapshot
	percentPercentCount int
}

// TokenTransformer implements ParserScanner
var _ ParserScanner = (*TokenTransformer)(nil)

type TokenSnapshot struct {
	Token      Token
	ByteOffset int
	Line       int
	Column     int
	Lexeme     []byte
}

func (t *TokenTransformer) Err() error {
	if len(t.tokenQueue) == 0 {
		return t.Scanner.Err()
	}
	return nil
}

func (t *TokenTransformer) Token() Token {
	return t.tokenQueue[0].Token
}

func (t *TokenTransformer) ByteOffset() int {
	return t.tokenQueue[0].ByteOffset
}

func (t *TokenTransformer) Line() int {
	return t.tokenQueue[0].Line
}

func (t *TokenTransformer) Column() int {
	return t.tokenQueue[0].Column
}

func (t *TokenTransformer) Lexeme() []byte {
	return t.tokenQueue[0].Lexeme
}

func (t *TokenTransformer) Next() bool {
	// Discard the first token in our queue
	if len(t.tokenQueue) > 0 {
		t.tokenQueue = t.tokenQueue[1:]
	}

	t.ensureQueuedTokens(2)
	if len(t.tokenQueue) >= 2 && t.tokenQueue[0].Token == TokenId && t.tokenQueue[1].Token == TokenColon {
		t.tokenQueue[0].Token = TokenIdColon
	}
	if len(t.tokenQueue) >= 1 && t.tokenQueue[0].Token == TokenPercentPercent {
		t.percentPercentCount++
		if t.percentPercentCount == 2 {
			t.tokenQueue = slices.Insert(t.tokenQueue, 1, TokenSnapshot{
				Token:      TokenEpilogue,
				ByteOffset: t.tokenQueue[0].ByteOffset + len(t.tokenQueue[0].Lexeme),
				Line:       t.tokenQueue[0].Line + 1,
				Column:     1,
				Lexeme:     []byte(""),
			})
		}
	}
	return len(t.tokenQueue) > 0 && t.tokenQueue[0].Token != EndToken
}

func (t *TokenTransformer) ensureQueuedTokens(count int) {
	for len(t.tokenQueue) < count {
		if len(t.tokenQueue) > 0 && t.tokenQueue[len(t.tokenQueue)-1].Token == EndToken {
			// We do not read past TokenEnd.
			return
		}
		t.Scanner.Next()
		t.tokenQueue = append(t.tokenQueue, TokenSnapshot{
			Token:      t.Scanner.Token(),
			ByteOffset: t.Scanner.ByteOffset(),
			Line:       t.Scanner.Line(),
			Column:     t.Scanner.Column(),
			Lexeme:     t.Scanner.Lexeme(),
		})
	}
}

func (t *TokenTransformer) FilePath() string {
	return t.Scanner.FilePath()
}
