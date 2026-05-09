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
