package parser

const (
	TokenError Token = ^0
)

// ReadEpilogue is an extension of the generated scanner providing functionality to consume the epilogue.
func (s *Scanner) ReadEpilogue() {
	s.lexemeStartIdx = s.lexemeEndIdx
	s.lexemePeekIdx = s.lexemeEndIdx

	// Consume all runes until the end of source
	s.lexemePeekIdx = len(s.source)
	s.lexemeEndIdx = s.lexemePeekIdx
	s.token = TokenEpilogue
}

// ReadTag is an extension to the generated scanner providing functionality to consume tags.
func (s *Scanner) ReadTag() {
	var nesting int
	var previousRune byte
	for {
		currRune := s.source[s.lexemePeekIdx]
		switch currRune {
		case '<':
			nesting++
		case '>':
			// We do not count a C++ '->' as a closing tag.
			if previousRune != '-' {
				if nesting == 0 {
					// Advance the rune reader to the next rune
					s.lexemePeekIdx++
					s.lexemeEndIdx = s.lexemePeekIdx
					s.token = TokenTag
					return
				}
				nesting--
			}
		}
		previousRune = currRune
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			break
		}
	}

	// The tag was not closed.
	s.token = InvalidToken
}

// ReadPrologue is an extension to the generated scanner providing functionality to consume the prologue.
func (s *Scanner) ReadPrologue() {
	var previousRune byte
	for {
		currRune := s.source[s.lexemePeekIdx]
		switch {
		case previousRune == '%' && currRune == '}':
			s.lexemePeekIdx++
			s.lexemeEndIdx = s.lexemePeekIdx
			s.token = TokenPrologue
			return
		case previousRune == '/' && currRune == '*':
			s.skipBlockComment()
		case previousRune == '/' && currRune == '/':
			s.skipLineComment()
		case currRune == '"':
			s.skipString('"')
		case currRune == '\'':
			s.skipString('\'')
		}

		// We do not use currRune here, because the skip helper methods might move the rune reader forward.
		previousRune = s.source[s.lexemePeekIdx]
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			break
		}
	}

	// The prologue was not closed.
	s.token = InvalidToken
}

// ReadBracedCode is an extension to the generated scanner providing functionality to consume braced code.
func (s *Scanner) ReadBracedCode() {
	s.readBracedContent(TokenBracedCode)
}

// ReadBracedPredicate is an extension to the generated scanner providing functionality to consume braced predicates.
func (s *Scanner) ReadBracedPredicate() {
	s.readBracedContent(TokenBracedPredicate)
}

//nolint:cyclop // The complexity is fine. Changes to the code would make it harder to understand.
func (s *Scanner) readBracedContent(token Token) {
	var previousRune byte
	var nestingLevel int
	for {
		currRune := s.source[s.lexemePeekIdx]
		switch {
		// <% is the C digraph for {
		case currRune == '{' || (previousRune == '<' && currRune == '%'):
			nestingLevel++
		// %> is the C digraph for }
		case currRune == '}' || (previousRune == '%' && currRune == '>'):
			if nestingLevel == 0 {
				s.lexemePeekIdx++
				s.lexemeEndIdx = s.lexemePeekIdx
				s.token = token
				return
			}
			nestingLevel--
		case previousRune == '/' && currRune == '*':
			s.skipBlockComment()
		case previousRune == '/' && currRune == '/':
			s.skipLineComment()
		case currRune == '"':
			s.skipString('"')
		case currRune == '\'':
			s.skipString('\'')
		}

		previousRune = s.source[s.lexemePeekIdx]
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			break
		}
	}

	// The braced content was not closed.
	s.token = InvalidToken
}

func (s *Scanner) skipBlockComment() {
	var previousRune byte
	for {
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			return
		}
		currRune := s.source[s.lexemePeekIdx]

		if previousRune == '*' && currRune == '/' {
			return
		}

		previousRune = currRune
	}
}

func (s *Scanner) skipLineComment() {
	for {
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			return
		}
		if s.source[s.lexemePeekIdx] == '\n' {
			return
		}
	}
}

func (s *Scanner) skipString(quote byte) {
	var previousRune byte
	for {
		s.lexemePeekIdx++
		if s.lexemePeekIdx >= len(s.source) {
			return
		}
		currRune := s.source[s.lexemePeekIdx]

		if currRune == quote && previousRune != '\\' {
			return
		}

		previousRune = currRune
	}
}
