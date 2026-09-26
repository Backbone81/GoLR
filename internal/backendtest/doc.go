// Package backendtest contains the test harness which proves that the code the language backends emit reproduces the
// committed traces of the golden corpus. It tests the emitted code and the driver around it, not the construction of
// the tables, which is covered where those tables are built.
//
// The committed traces are the specification. The Go backend writes them with "make update-golden", and a person
// reviews the diff before it is committed. Every runner in every language prints the same canonical text for the same
// input, so comparing a backend against the corpus is a plain text diff instead of a language specific assertion.
//
// # Trace format
//
// A case has three traces: scanner.trace, parser.trace and tree.trace. Each is UTF-8 with one line per event, LF line
// endings and a trailing newline. An empty trace is the empty file. A line is the position padded to 7 characters, the
// keyword padded to 7 characters and the payload, separated by single spaces, with no trailing whitespace. A line
// without payload ends after the keyword.
//
// The position is "line:column", both one based. The column counts characters: a byte which continues a UTF-8 character
// does not count. An event is positioned where its token starts, and the end of the input one past the last byte.
//
// No line carries a state or production index, so a trace is a property of the grammar and the input. A production is
// named "lhs => rhs", or "lhs => ε" for an empty right hand side. The end of input symbol is "$end", the error symbol
// "$error", and a token no scanner rule matched "$invalid". The leading dollar sign keeps them apart from every rule
// name.
//
// The scanner trace reports every match, skipped rules included, because a runner is never told which rules are
// skipped:
//
//	TOKEN   rule "lexeme"      a rule matched
//	ERROR   "lexeme"           a stretch of bytes no rule matched
//	EOF                        the whole input was consumed
//
// The parser trace leaves the skipped tokens out:
//
//	SHIFT   terminal "lexeme"                 a terminal was shifted, "$end" with an empty lexeme included
//	REDUCE  lhs => rhs                        a production was reduced
//	ERROR   unexpected t, expecting a or b    a syntax error at the lookahead, see below for the message
//	ERROR   (suppressed) unexpected t, ...    one hit before three tokens were shifted after the previous error
//	POP     symbol                            recovery dropped a state and the symbol it had parsed
//	RESYNC                                    recovery shifted the error symbol and resumed
//	DISCARD terminal "lexeme"                 recovery dropped the lookahead it keeps failing on
//	ACCEPT                                    the input was accepted
//	FAIL                                      the parse was given up
//
// The message of an ERROR line names a terminal by its string alias, or by its name if it has none, and the end of
// input and a token no scanner rule matched as "end of input" and "invalid input". It lists the terminals the parser
// would have shifted instead, in the order of their columns, joined by "or". It lists at most four, and none if there
// are more.
//
// The tree trace is the parse tree in pre-order, one node per line, and empty for a parse which was given up. The
// keyword column holds the span of the input a node covers, as "byteOffset+byteLength", and the payload is the node
// indented by two spaces per level of depth: a terminal as `terminal "text"`, a nonterminal as its production, and an
// error node as "$error". The spans follow these rules:
//
//   - A terminal covers its lexeme.
//   - A nonterminal starts where its first child starts and ends where its last child ends.
//   - An ε-production has zero length at the end of the symbol to its left. At the bottom of the stack there is no such
//     symbol and it sits where the parser was looking.
//   - An error node runs from the start of what its recovery round threw away to the point where that round resumed. A
//     round which threw no bytes away leaves it empty at the end of the symbol to its left.
//
// # Escaping
//
// A lexeme is written between double quotes and escaped byte by byte: `\\`, `\"`, `\n`, `\r`, `\t`, the printable ASCII
// range 0x20 to 0x7E as it is, and every other byte as lower case `\xHH`. A multi byte UTF-8 character is therefore
// one escape per byte, because bytes are the only unit every target language agrees on without conversion. The escaping
// is a minimal C style string literal, so every target language writes it with a switch and no escaping library. Go's
// %q is not the same: it keeps UTF-8 characters whole and uses \a, \b, \f, \v and \u.
package backendtest
