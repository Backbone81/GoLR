package golr

import "github.com/backbone81/golr/internal/parsergen/backend"

// Item is an LR(1) item. It consists of a core, which is a production together with a position within that production,
// and the set of terminals which may follow the production once it has been reduced.
type Item struct {
	// Core is the production and the position within that production.
	Core backend.Core

	// LookaheadSet is the set of terminals which may follow the production of the core.
	LookaheadSet backend.LookaheadSet
}
