package parser

// ReadEpilogue is an extension of the generated scanner providing functionality to consume the epilogue.
func (s *Scanner) ReadEpilogue() {
	s.err = nil
	s.tokenStart = s.tokenEnd
	s.runeReader = s.tokenEnd

	// Consume all runes until the end of source
	for s.runeReader.Next() {
	}
	s.tokenEnd = s.runeReader
	s.token = TokenEpilogue
}

// ReadTag is an extension to the generated scanner providing functionality to consume tags.
func (s *Scanner) ReadTag() {
	s.err = nil
	var nesting int
	var previousRune rune
	for {
		currRune := s.runeReader.Rune()
		switch currRune {
		case '<':
			nesting++
		case '>':
			// We do not count a C++ '->' as a closing tag.
			if previousRune != '-' {
				if nesting == 0 {
					// Advance the rune reader to the next rune
					s.runeReader.Next()
					s.tokenEnd = s.runeReader
					s.token = TokenTag
					return
				}
				nesting--
			}
		}
		previousRune = currRune
		if !s.runeReader.Next() {
			break
		}
	}

	// The tag was not closed.
	s.token = InvalidToken
}

// ReadPrologue is an extension to the generated scanner providing functionality to consume the prologue.
func (s *Scanner) ReadPrologue() {
	s.err = nil
	var previousRune rune
	for {
		currRune := s.runeReader.Rune()
		switch {
		case previousRune == '%' && currRune == '}':
			s.runeReader.Next()
			s.tokenEnd = s.runeReader
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
		previousRune = s.runeReader.Rune()
		if !s.runeReader.Next() {
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
	s.err = nil
	var previousRune rune
	var nestingLevel int
	for {
		currRune := s.runeReader.Rune()
		switch {
		// <% is the C digraph for {
		case currRune == '{' || (previousRune == '<' && currRune == '%'):
			nestingLevel++
		// %> is the C digraph for }
		case currRune == '}' || (previousRune == '%' && currRune == '>'):
			if nestingLevel == 0 {
				s.runeReader.Next()
				s.tokenEnd = s.runeReader
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

		previousRune = s.runeReader.Rune()
		if !s.runeReader.Next() {
			break
		}
	}

	// The braced content was not closed.
	s.token = InvalidToken
}

func (s *Scanner) skipBlockComment() {
	var previousRune rune
	for {
		if !s.runeReader.Next() {
			return
		}
		currRune := s.runeReader.Rune()

		if previousRune == '*' && currRune == '/' {
			return
		}

		previousRune = currRune
	}
}

func (s *Scanner) skipLineComment() {
	for {
		if !s.runeReader.Next() {
			return
		}
		if s.runeReader.Rune() == '\n' {
			return
		}
	}
}

func (s *Scanner) skipString(quote rune) {
	var previousRune rune
	for {
		if !s.runeReader.Next() {
			return
		}
		currRune := s.runeReader.Rune()

		if currRune == quote && previousRune != '\\' {
			return
		}

		previousRune = currRune
	}
}
