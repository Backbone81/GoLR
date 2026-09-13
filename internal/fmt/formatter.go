package fmt

import (
	"bytes"

	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

type Formatter struct {
	config Config

	scanner     *golrparser.Scanner
	indentLevel int

	// lastScannerIdentifierLen is the length of the most recently emitted identifier inside a @scanner section,
	// i.e. the left hand side of the scanner rule whose ":" comes next.
	lastScannerIdentifierLen int

	// scannerRuleGroups collects scanner rule marks seen during the first pass, grouped by contiguous runs of
	// rules not separated by an explicit blank line. It always has at least one (possibly empty) group; a new
	// one is started as soon as an explicit blank line is seen inside a @scanner section, and it stays empty if
	// no rule follows before the next one. Consumed by alignScannerRules afterwards.
	scannerRuleGroups [][]scannerRuleRHS

	// indentNext reports if the next output needs to be indented or not.
	indentNext bool

	// pendingLinebreak reports if a linebreak was requested but not yet written to the output. Requesting a
	// linebreak is lazy so that a same-line trailing comment can still be emitted before it.
	pendingLinebreak bool

	// pendingBlankLine reports if a blank line before the next emitted content has already been scheduled, so a
	// second, independent reason to want one (e.g. automatic spacing between rules and a user's own blank line
	// around a comment coinciding) does not stack into two. Cleared once real content is emitted.
	pendingBlankLine bool

	// emitTight reports if the next emit should not write a whitespace to separate the previous token from the next
	// one.
	emitTight bool

	// context is a stack which describes the current nesting
	context []golrparser.Token

	// explicitBlankLine reports if the whitespace just consumed contained a blank line, i.e. the user separated the
	// surrounding tokens by an empty line of their own.
	explicitBlankLine bool

	// explicitNewline reports if the whitespace just consumed contained a linebreak, i.e. the next token starts on a
	// new source line instead of trailing the previous token.
	explicitNewline bool

	// explicitComment reports if the token just emitted was a comment, so a following explicit blank line can be
	// attributed to it instead of to whatever token comes after the whitespace.
	explicitComment bool

	output *bytes.Buffer
}

// scannerRuleRHS notes down where a scanner rule's right hand side starts, so a second pass can align it with
// the other rules in the same group.
type scannerRuleRHS struct {
	// offset is the byte position in the output right after the rule's ":", where padding is inserted.
	offset int

	// ruleNameLen is the length of the rule's identifier.
	ruleNameLen int
}

func NewFormatter() *Formatter {
	return &Formatter{
		config: DefaultConfig,
	}
}

// Format pretty prints source in two passes: token-by-token emission with lazily scheduled line breaks (see
// linebreak/blankLine/emit), followed by alignScannerRules, which pads scanner rule ":" columns using the marks
// collected during the first pass.
//
//nolint:cyclop,funlen // This method cannot be simplified any further.
func (f *Formatter) Format(source []byte, filePath string) []byte {
	f.indentLevel = 0
	f.indentNext = true
	f.pendingLinebreak = false
	f.pendingBlankLine = false
	f.emitTight = false
	f.context = f.context[:0]
	f.explicitBlankLine = false
	f.explicitNewline = false
	f.explicitComment = false
	f.lastScannerIdentifierLen = 0
	f.scannerRuleGroups = [][]scannerRuleRHS{nil}

	f.output = bytes.NewBuffer(make([]byte, 0, len(source)))
	f.scanner = golrparser.NewScanner(source, filePath)
	for f.scanner.Next() {
		// explicitComment must be reset before every token, not just at the bottom of the loop like
		// explicitBlankLine/explicitNewline, so it reflects only the token from the immediately preceding
		// iteration; TokenWhitespace below still needs that old value, so grab a copy first.
		explicitComment := f.explicitComment
		f.explicitComment = false

		//nolint:exhaustive // We are only interested in a few special tokens
		switch f.scanner.Token() {
		case golrparser.TokenWhitespace:
			f.onTokenWhitespace(explicitComment)
			continue
		case golrparser.TokenComment:
			f.onTokenComment()
		case golrparser.TokenColon:
			f.onTokenColon()
		case golrparser.TokenPipe:
			f.onTokenPipe()
		case golrparser.TokenSemi:
			f.onTokenSemi()
		case golrparser.TokenLbrace:
			f.onTokenLbrace()
		case golrparser.TokenRbrace:
			f.onTokenRbrace()
		case golrparser.TokenLparen:
			f.onTokenLparen()
		case golrparser.TokenRparen:
			f.onTokenRparen()
		case golrparser.TokenIdentifier:
			f.onTokenIdentifier()
		case golrparser.TokenScanner:
			f.onTokenScanner()
		case golrparser.TokenParser:
			f.onTokenParser()
		case golrparser.TokenPrecedence:
			f.onTokenPrecedence()
		default:
			f.emit(f.scanner.Lexeme())
		}

		f.explicitBlankLine = false
		f.explicitNewline = false
		// explicitComment is intentionally not reset here; see the reset at the top of the loop.
	}
	f.linebreak()
	return f.alignScannerRules(f.output.Bytes())
}

func (f *Formatter) onTokenWhitespace(explicitComment bool) {
	// We are looking for linebreaks and explicit blank lines by the user.
	newlines := bytes.Count(f.scanner.Lexeme(), []byte("\n"))
	if newlines >= 1 {
		f.explicitNewline = true
	}
	if newlines >= 2 {
		f.explicitBlankLine = true
		if explicitComment {
			// The user separated the comment we just emitted from what follows with a blank line;
			// keep it instead of collapsing the comment onto the next token.
			f.blankLine()
		}
		if f.currentContext() == golrparser.TokenScanner {
			// Start a new alignment group. It stays empty if no rule follows before the next blank
			// line, which is harmless.
			f.scannerRuleGroups = append(f.scannerRuleGroups, nil)
		}
	}
}

func (f *Formatter) onTokenComment() {
	isLineComment := bytes.HasPrefix(f.scanner.Lexeme(), []byte("//"))
	switch {
	case f.explicitNewline:
		// The comment starts on its own line rather than trailing the previous token.
		if f.explicitBlankLine {
			// The user separated the comment from the previous content with a blank line; keep it.
			// blankLine is idempotent, so this composes correctly with automatic spacing (e.g. between
			// top-level parser rules) that may already have scheduled the same blank line.
			f.blankLine()
		} else if !f.pendingLinebreak {
			// Only schedule if nothing is pending yet: linebreak writes immediately when a linebreak is
			// already pending, which would add an unwanted extra "\n" here.
			f.linebreak()
		}
		f.emit(f.scanner.Lexeme())
	case f.pendingLinebreak:
		// The comment trails the previous token, but a linebreak (and possibly a blank line) is already
		// scheduled to run after that token. Emit the comment before that instead of flushing it early.
		hadBlankLine := f.pendingBlankLine
		f.pendingLinebreak = false
		f.pendingBlankLine = false
		f.emit(f.scanner.Lexeme())
		f.pendingLinebreak = true
		f.pendingBlankLine = hadBlankLine
	default:
		f.emit(f.scanner.Lexeme())
	}
	if isLineComment {
		// Anything after "//" on the same line would otherwise be swallowed into the comment.
		f.linebreak()
	}
	// Read by the top of the loop on the next token, to detect a blank line right after this comment.
	f.explicitComment = true
}

func (f *Formatter) onTokenColon() {
	//nolint:exhaustive // we are not interested in all tokens
	switch f.currentContext() {
	case golrparser.TokenScanner:
		// Note down where the rule's right hand side starts, so the second pass can align it.
		mark := scannerRuleRHS{
			offset:      f.output.Len() + 1,
			ruleNameLen: f.lastScannerIdentifierLen,
		}
		lastGroup := len(f.scannerRuleGroups) - 1
		f.scannerRuleGroups[lastGroup] = append(f.scannerRuleGroups[lastGroup], mark)

		f.emitTight = true
		f.emit(f.scanner.Lexeme())
	case golrparser.TokenParser:
		// Indent the alternatives under the rule name; matching indentDec runs at the rule's ";".
		f.indentInc()
		f.linebreak()
		f.emit(f.scanner.Lexeme())
	default:
		f.emitTight = true
		f.emit(f.scanner.Lexeme())
	}
}

func (f *Formatter) onTokenPipe() {
	if !f.pendingLinebreak {
		// Only schedule if nothing is pending yet; see the same guard in TokenComment for why.
		f.linebreak()
	}
	f.emit(f.scanner.Lexeme())
}

func (f *Formatter) onTokenSemi() {
	//nolint:exhaustive // we are not interested in all tokens
	switch f.currentContext() {
	case golrparser.TokenParser:
		f.linebreak()
		f.emit(f.scanner.Lexeme())
		f.indentDec()
		// Separate top-level rules with a blank line; idempotent, so it composes with a blank line a
		// trailing comment on the next rule also wants (see the TokenComment case).
		f.blankLine()
	default:
		f.emitTight = true
		f.emit(f.scanner.Lexeme())
		f.linebreak()
	}
}

func (f *Formatter) onTokenLbrace() {
	f.emit(f.scanner.Lexeme())
	f.linebreak()
	f.indentInc()
}

func (f *Formatter) onTokenRbrace() {
	f.indentLevel = max(f.indentLevel-1, 0)
	// A blank line may still be pending to separate the previous rule from a following one (see
	// TokenSemi); it must not leak into a blank line before the closing brace itself.
	f.pendingBlankLine = false
	if !f.indentNext {
		// Skip if we're already at the start of a fresh line (e.g. an empty "{}" block), otherwise this
		// would add a spurious blank line before "}".
		f.linebreak()
	}
	f.emit(f.scanner.Lexeme())
	f.blankLine()

	// We remove any context we did push onto the context stack.
	f.popContext()
}

func (f *Formatter) onTokenLparen() {
	if f.currentContext() == golrparser.TokenPrecedence {
		// "@precedence" is ambiguous: "@precedence { ... }" opens a block, but "@precedence(...)" is
		// an inline alternative annotation with no matching "}". TokenPrecedence below pushes eagerly
		// as if it were the block form; seeing a "(" instead of a "{" proves it wasn't, so undo it.
		f.popContext()
	}

	// Left parenthesis always sit close to the previous token.
	f.emitTight = true
	f.emit(f.scanner.Lexeme())

	// The following token also needs to sit close to the left parenthesis.
	f.emitTight = true
}

func (f *Formatter) onTokenRparen() {
	// Right parenthesis always sit close to the previous token.
	f.emitTight = true
	f.emit(f.scanner.Lexeme())
}

func (f *Formatter) onTokenIdentifier() {
	//nolint:exhaustive,gocritic // we are not interested in all tokens
	switch f.currentContext() {
	case golrparser.TokenScanner:
		// This identifier is the left hand side of a scanner rule; remember its length for the ":"
		// that follows right after.
		f.lastScannerIdentifierLen = len(f.scanner.Lexeme())
		if f.explicitBlankLine {
			// Preserve the user's blank line between scanner rules.
			f.linebreak()
		}
	}
	f.emit(f.scanner.Lexeme())
}

func (f *Formatter) onTokenScanner() {
	// Track that we're inside @scanner; popped again on the matching "}".
	f.emit(f.scanner.Lexeme())
	f.pushContext(golrparser.TokenScanner)
}

func (f *Formatter) onTokenParser() {
	// Track that we're inside @parser; popped again on the matching "}".
	f.emit(f.scanner.Lexeme())
	f.pushContext(golrparser.TokenParser)
}

func (f *Formatter) onTokenPrecedence() {
	// Pushed eagerly as if opening a "@precedence { ... }" block; TokenLparen above repairs this if
	// it turns out to be the inline "@precedence(...)" annotation instead.
	f.emit(f.scanner.Lexeme())
	f.pushContext(golrparser.TokenPrecedence)
}

func (f *Formatter) indentInc() {
	f.indentLevel++
}

func (f *Formatter) indentDec() {
	f.indentLevel = max(f.indentLevel-1, 0)
}

// linebreak requests a linebreak before the next emitted content. It is lazy: the actual "\n" is written by emit,
// so that a same-line trailing comment can still be inserted before it. If a linebreak is already pending, this one
// is written immediately, keeping the pending flag around for the one emit will still owe.
func (f *Formatter) linebreak() {
	if f.pendingLinebreak {
		f.output.Write([]byte("\n"))
	}
	f.pendingLinebreak = true
	f.indentNext = true
}

// blankLine requests a blank line before the next emitted content. Like linebreak, it is lazy: the actual "\n\n" is
// written by emit. It is idempotent, so a second, independent reason to want a blank line (e.g. automatic spacing
// between rules and a user's own blank line around a comment coinciding) does not stack into two.
func (f *Formatter) blankLine() {
	f.pendingLinebreak = true
	f.pendingBlankLine = true
	f.indentNext = true
}

// emit flushes any pending linebreak or blank line, indents if needed, and writes data.
func (f *Formatter) emit(data []byte) {
	switch {
	case f.pendingBlankLine:
		// A pending blank line implies a pending linebreak too (see blankLine), so check it first.
		f.output.Write([]byte("\n\n"))
	case f.pendingLinebreak:
		f.output.Write([]byte("\n"))
	}
	f.pendingLinebreak = false
	f.pendingBlankLine = false
	if f.indentNext {
		for range f.indentLevel {
			f.output.Write([]byte(f.config.Indentation))
		}
		f.emitTight = true
	}
	if !f.emitTight {
		f.output.Write([]byte(" "))
	}
	f.indentNext = false
	f.emitTight = false
	f.output.Write(data)
}

func (f *Formatter) pushContext(token golrparser.Token) {
	f.context = append(f.context, token)
}

func (f *Formatter) popContext() {
	f.context = f.context[:max(len(f.context)-1, 0)]
}

func (f *Formatter) currentContext() golrparser.Token {
	if len(f.context) == 0 {
		return golrparser.InvalidToken
	}
	return f.context[len(f.context)-1]
}

// alignScannerRules is the second pass over the pretty printed output. It pads the ":" of scanner rules within
// each group (a run of rules not separated by an explicit blank line) so their right hand sides start in the same
// column.
func (f *Formatter) alignScannerRules(data []byte) []byte {
	result := make([]byte, 0, len(data))
	lastOffset := 0
	for _, group := range f.scannerRuleGroups {
		maxPrefixLen := 0
		for _, mark := range group {
			maxPrefixLen = max(maxPrefixLen, mark.ruleNameLen)
		}

		for _, mark := range group {
			result = append(result, data[lastOffset:mark.offset]...)
			for range maxPrefixLen - mark.ruleNameLen {
				result = append(result, ' ')
			}
			lastOffset = mark.offset
		}
	}
	result = append(result, data[lastOffset:]...)
	return result
}
