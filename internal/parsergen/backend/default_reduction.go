package backend

// ApplyDefaultReductions picks a default reduce action for every state which does not already have one, compressing the
// reduce actions of the parser tables. It is the equivalent of GNU Bison's default behavior.
//
// The pass is a table transformation on the finished parser: it must run after the conflicts have been resolved, when
// no state has more than one action for any one terminal, so that the reduce actions it reads are the reduce actions
// the parser will actually take. It is idempotent - a state which already carries a default reduction is left
// untouched, so running it over the Bison backed tables, which already have their defaults, changes nothing.
func ApplyDefaultReductions(parser *Parser) {
	for stateIdx := range parser.States {
		applyDefaultReduction(&parser.States[stateIdx])
	}
}

// applyDefaultReduction turns the reduce action with the widest lookahead set of the state into its default reduction:
// the action taken for any lookahead not claimed by a shift or by one of the remaining, explicit reduce actions. The
// chosen reduce action is then removed from the explicit reduce actions, because the default arm covers all of its
// lookaheads now and a backend would otherwise emit it twice.
//
// # Which reduction is chosen
//
// The default reduction is the one which covers the most lookahead terminals, so that it eliminates the most explicit
// reduce entries; ties go to the lowest production index for a reproducible result. This is Bison's documented `most`
// heuristic. ReduceActionSet is ordered by production index and there is at most one reduce action per production in a
// state. Iterating the set therefore visits the productions in strictly ascending order with each production's full
// lookahead coverage in hand, so a running best kept with a strict greater-than comparison lands on the widest set and,
// among equally wide sets, on the one seen first, which is the one with the lowest production index.
//
// # Why applying it everywhere is correct
//
// A default reduction never changes the language the parser accepts nor the parse it produces, so the tables stay
// equivalent to the ones without it. This rests on the defining property of an LR parser: it never shifts on an
// erroneous token. Where the explicit tables would report an error on some lookahead, the compressed tables reduce
// instead - possibly several times as the reductions expose further default reductions - but every state reached this
// way is reached without shifting the offending token, so the parser still reports the error before it ever consumes
// the token. The only observable difference is the moment the error is reported, which is delayed to the next state
// that has no default reduction, exactly as in Bison. No input is accepted that the explicit tables would reject, and
// no parse changes for accepted input.
//
// The compression is applied to every state that has a reduction, including states which also shift. A shift is always
// an explicit action keyed on its terminal, so it takes precedence over the default arm and is never replaced by it;
// only the reduce-or-error lookaheads of the state fall through to the default.
//
// The accept action is left alone. The GoLR cores encode it as a reduce of the accept production with an empty
// reduction lookahead set, which the backend already renders as the unconditional default of its state. Such an action
// covers no terminal, so it is never chosen as a default reduction here and never removed; skipping the empty lookahead
// sets keeps the accept out of this compression and keeps a reduce action no terminal can trigger from becoming a
// state's default.
func applyDefaultReduction(state *State) {
	if state.DefaultReduceProductionIdx != nil {
		// The state already has a default reduction, most likely because it came from a Bison backed core. Leaving it
		// untouched keeps the pass idempotent.
		return
	}

	bestProductionIdx := -1
	bestLookaheadCount := 0
	var bestReduceAction ReduceAction

	for _, reduceAction := range state.ReduceActions.All() {
		lookaheadCount := reduceAction.LookaheadSet.Length()
		if lookaheadCount == 0 {
			// The accept action, and any reduce action no terminal can trigger, cover no lookahead and must not become
			// the default reduction of the state.
			continue
		}

		// A strict greater-than keeps the first, and therefore lowest indexed, production among equally wide lookahead
		// sets, because the reduce actions are visited in ascending production index order.
		if lookaheadCount > bestLookaheadCount {
			bestProductionIdx = reduceAction.ProductionIdx
			bestLookaheadCount = lookaheadCount
			bestReduceAction = reduceAction
		}
	}

	if bestProductionIdx == -1 {
		// The state has no reduce action which any terminal triggers, so there is nothing to turn into a default.
		return
	}

	// The pointer target has to outlive this function, so it escapes to the heap.
	productionIdx := bestProductionIdx
	state.DefaultReduceProductionIdx = &productionIdx

	// The default arm now covers every lookahead of the chosen reduce action, so its explicit entry has to go or the
	// backend would emit the same reduction twice.
	state.ReduceActions.Remove(bestReduceAction)
}
