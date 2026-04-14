package backend

import (
	intbackends "golr/internal/parsergen/backend"
)

// Core is the core of an LR(1) item consisting of a production index and a position within that production. The values
// for the production index and the position must be in the range of [0, 65535].
type Core = intbackends.Core

// NewCore creates a new core with the given production index and the position.
var NewCore = intbackends.NewCore

// CoreSet is an ordered set of cores.
type CoreSet = intbackends.CoreSet

// NewCoreSet creates a new core set.
var NewCoreSet = intbackends.NewCoreSet

// LookaheadSet is a set of terminal indexes.
type LookaheadSet = intbackends.LookaheadSet

// NewLookaheadSet creates a new lookahead set.
var NewLookaheadSet = intbackends.NewLookaheadSet

// Parser is a parser.
type Parser = intbackends.Parser

// ReduceAction is a reduce action of an LR(1) item consisting of a lookahead set of terminals and a production index
// to reduce for. The values for the production must be in the range of [0, 65535].
type ReduceAction = intbackends.ReduceAction

// NewReduceAction creates a new reduce action with the given lookahead set and the production index.
var NewReduceAction = intbackends.NewReduceAction

// ReduceActionSet is an ordered set of ReduceAction elements.
type ReduceActionSet = intbackends.ReduceActionSet

// NewReduceActionSet creates a new ordered reduce action set.
var NewReduceActionSet = intbackends.NewReduceActionSet

// State represents a LR(1) state. The structure of this state is derived from definition 3.1 of IELR(1).
type State = intbackends.State

// TransitionAction is a transition action of an LR(1) item consisting of a symbol index representing a terminal or
// nonterminal and a state index for the target state. The values for the symbol index and the state index must be in
// the range of [0, 65535]. The symbol index is the index of a terminal or the index of a nonterminal added to the index
// of the last terminal + 1.
type TransitionAction = intbackends.TransitionAction

// NewTransitionAction creates a new transition action with the given symbol index and the target state index.
var NewTransitionAction = intbackends.NewTransitionAction

// TransitionActionSet is an ordered set of transition actions.
type TransitionActionSet = intbackends.TransitionActionSet

// NewTransitionActionSet creates a new ordered transition action set.
var NewTransitionActionSet = intbackends.NewTransitionActionSet
