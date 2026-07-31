package selftest

// Progress describes progress information of the self test. It can reflect progress of a single grammar being tested,
// progress of a single worker, or overall progress since the start of the self-test.
type Progress struct {
	// Generated is the number of random grammars produced, which is Compared plus Skipped.
	Generated int64

	// Compared is the number of grammars whose tables were all built and checked against each other.
	Compared int64

	// Skipped is the number of grammars whose canonical LR(1) oracle exceeded the addressable state limit and could
	// therefore not be built. A run where this dominates is exploring grammars too large for the oracle and proves
	// little, so it is worth watching.
	Skipped int64

	// Failed is the number of grammars which failed the check. Anything above zero is a finding: the grammar is in
	// FailureDir and the run should be reported.
	Failed int64

	// Discriminating is the number of grammars where LALR(1) has a conflict canonical LR(1) does not. These are the
	// non-LALR shapes where the IELR(1) splitting under test earns its keep, so this is the coverage number which
	// matters most.
	Discriminating int64

	// SplittingFired is the number of grammars where the IELR(1) table came out larger than the LALR(1) table, i.e.
	// phase 3 actually split a state.
	SplittingFired int64

	// SentencesParsed is the total number of sentences the two tables were driven through.
	SentencesParsed int64
}

// Add combines the other progress with the current one.
func (p *Progress) Add(other Progress) {
	p.Generated += other.Generated
	p.Compared += other.Compared
	p.Skipped += other.Skipped
	p.Failed += other.Failed
	p.Discriminating += other.Discriminating
	p.SplittingFired += other.SplittingFired
	p.SentencesParsed += other.SentencesParsed
}
