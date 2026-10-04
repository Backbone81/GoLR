package counterexample

// action is the action of the product parser which leads from a ProductConfiguration to a successor, see section 5.3.
type action uint8

const (
	// initialAction leads to the initial configuration, which is the parent itself.
	initialAction action = iota
	transitionAction
	productionStepAction
	emptyDerivationAction
	reductionAction
	reverseTransitionAction
	reverseProductionStepAction
)

// PendingConfiguration is a ProductConfiguration in the queue, given by the configuration it succeeds and the action
// which leads to it. The search queues many more configurations than it processes, so a configuration is only built
// when it is removed from the queue, see unifyingBuilder.build.
//
// The queue holds many more of them than the search processes, so the fields are as narrow as they can be, see
// NewLookupTables for the items. The parent comes first, as the alignment of the pointer would pad the fields in front
// of it.
type PendingConfiguration struct {
	parent *ProductConfiguration

	// Cost is the cost of the ProductConfiguration.
	Cost int32

	// itemIdxs holds the item the action appends or prepends per parser.
	itemIdxs [2]int32

	action action

	// parserIdx is the parser the action changes, unless it changes both.
	parserIdx uint8
}

// parserEdit is the change an action makes to the items and the depth of one parser.
type parserEdit struct {
	changed bool

	// dropped is the number of items removed from the end, before appended is appended. appended and prepended are
	// noItemIdx when the action appends or prepends no item.
	dropped   int
	appended  int
	prepended int
	depth     int
}
