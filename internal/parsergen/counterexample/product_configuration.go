package counterexample

// ProductConfiguration is a configuration of the product parser, see figure 8: the two parsers it simulates, the first
// one reducing the reduce item and the second one shifting with the other item or reducing it.
type ProductConfiguration struct {
	Parsers [2]SimulatedParser

	// TerminalShifted reports if the parsers shifted the conflict terminal behind the dot.
	TerminalShifted bool

	// Cost is the sum of the costs of the actions which led to the ProductConfiguration.
	Cost int
}

// Successor returns a copy of the ProductConfiguration with the cost added. The sequences are immutable, so the copy
// shares them.
func (c *ProductConfiguration) Successor(cost int) ProductConfiguration {
	result := *c
	result.Cost += cost
	return result
}

// pendingSuccessor returns the successor of the ProductConfiguration by the action on one parser with the item, at the
// cost of the action.
func (c *ProductConfiguration) pendingSuccessor(
	cost int,
	action action,
	parserIdx int,
	itemIdx int,
) PendingConfiguration {
	result := PendingConfiguration{
		parent:    c,
		Cost:      int32(c.Cost + cost), //nolint:gosec // The costs of a search stay far below.
		action:    action,
		parserIdx: uint8(parserIdx), //nolint:gosec // There are two parsers.
	}
	result.itemIdxs[parserIdx] = int32(itemIdx) //nolint:gosec // See NewLookupTables.
	return result
}

// Hash calculates a hash over both parsers and if the conflict terminal was shifted, which tells configurations apart
// that the search treats as different, see SimulatedParser.Hash.
func (c *ProductConfiguration) Hash() uint64 {
	return hashConfiguration([2]uint64{
		c.Parsers[0].Hash(),
		c.Parsers[1].Hash(),
	}, c.TerminalShifted)
}
