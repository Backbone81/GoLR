package parser

// ReadEpilogue is an extension of the generated scanner providing functionality to consume the epilogue.
func (s *Scanner) ReadEpilogue() {
	s.lexemeStartIdx = s.lexemeEndIdx

	// Consume all runes until the end of source
	s.lexemeEndIdx = len(s.source)
	s.token = TokenEpilogue
}

// ReadTag is an extension to the generated scanner providing functionality to consume tags.
func (s *Scanner) ReadTag() {
	var nesting int
	var previousRune byte
	peekIdx := s.lexemeEndIdx
	for peekIdx < len(s.source) {
		currRune := s.source[peekIdx]
		switch currRune {
		case '<':
			nesting++
		case '>':
			// We do not count a C++ '->' as a closing tag.
			if previousRune != '-' {
				if nesting == 0 {
					// Advance the rune reader to the next rune
					peekIdx++
					s.lexemeEndIdx = peekIdx
					s.token = TokenTag
					return
				}
				nesting--
			}
		}
		previousRune = currRune
		peekIdx++
	}

	// The tag was not closed.
	s.token = InvalidToken
}

// ReadPrologue is an extension to the generated scanner providing functionality to consume the prologue.
func (s *Scanner) ReadPrologue() {
	var previousRune byte
	peekIdx := s.lexemeEndIdx
	for peekIdx < len(s.source) {
		currRune := s.source[peekIdx]
		switch {
		case previousRune == '%' && currRune == '}':
			peekIdx++
			s.lexemeEndIdx = peekIdx
			s.token = TokenPrologue
			return
		case previousRune == '/' && currRune == '*':
			peekIdx = s.skipBlockComment(peekIdx)
		case previousRune == '/' && currRune == '/':
			peekIdx = s.skipLineComment(peekIdx)
		case currRune == '"':
			peekIdx = s.skipString(peekIdx, '"')
		case currRune == '\'':
			peekIdx = s.skipString(peekIdx, '\'')
		}

		// The skip helper methods might have moved the rune reader to the end of the source.
		if peekIdx >= len(s.source) {
			break
		}

		// We do not use currRune here, because the skip helper methods might move the rune reader forward.
		previousRune = s.source[peekIdx]
		peekIdx++
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
	peekIdx := s.lexemeEndIdx
	for peekIdx < len(s.source) {
		currRune := s.source[peekIdx]
		switch {
		// <% is the C digraph for {
		case currRune == '{' || (previousRune == '<' && currRune == '%'):
			nestingLevel++
		// %> is the C digraph for }
		case currRune == '}' || (previousRune == '%' && currRune == '>'):
			if nestingLevel == 0 {
				peekIdx++
				s.lexemeEndIdx = peekIdx
				s.token = token
				return
			}
			nestingLevel--
		case previousRune == '/' && currRune == '*':
			peekIdx = s.skipBlockComment(peekIdx)
		case previousRune == '/' && currRune == '/':
			peekIdx = s.skipLineComment(peekIdx)
		case currRune == '"':
			peekIdx = s.skipString(peekIdx, '"')
		case currRune == '\'':
			peekIdx = s.skipString(peekIdx, '\'')
		}

		// The skip helper methods might have moved the rune reader to the end of the source.
		if peekIdx >= len(s.source) {
			break
		}

		previousRune = s.source[peekIdx]
		peekIdx++
	}

	// The braced content was not closed.
	s.token = InvalidToken
}

func (s *Scanner) skipBlockComment(peekIdx int) int {
	var previousRune byte
	for {
		peekIdx++
		if peekIdx >= len(s.source) {
			return peekIdx
		}
		currRune := s.source[peekIdx]

		if previousRune == '*' && currRune == '/' {
			return peekIdx
		}

		previousRune = currRune
	}
}

func (s *Scanner) skipLineComment(peekIdx int) int {
	for {
		peekIdx++
		if peekIdx >= len(s.source) {
			return peekIdx
		}
		if s.source[peekIdx] == '\n' {
			return peekIdx
		}
	}
}

func (s *Scanner) skipString(peekIdx int, quote byte) int {
	for {
		peekIdx++
		if peekIdx >= len(s.source) {
			return peekIdx
		}
		currRune := s.source[peekIdx]

		if currRune == '\\' {
			// The escaped character is consumed by the increment at the top of the loop, so a quote or backslash
			// following a backslash can never be misread as a delimiter.
			peekIdx++
			continue
		}
		if currRune == quote {
			return peekIdx
		}
	}
}
