package parser

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

func (s *Scanner) ReadTag() {
	s.err = nil
	var nesting int
	previousRune := rune(0)
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
