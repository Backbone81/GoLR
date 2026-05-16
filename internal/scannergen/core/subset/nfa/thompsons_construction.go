package nfa

import (
	"context"
	"errors"
	"fmt"
	"runtime/trace"

	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// ThompsonsConstruction is responsible for building the NFA from regular expressions.
// It is an implementation of Thompson's Construction as described in the paper "Programming Techniques: Regular
// expression search algorithm" by Ken Thompson (https://doi.org/10.1145/363347.363387).
type ThompsonsConstruction struct{}

// NewThompsonsConstruction creates a new builder instance.
func NewThompsonsConstruction() *ThompsonsConstruction {
	return &ThompsonsConstruction{}
}

// Build creates a new NFA from the given regular expression.
func (b *ThompsonsConstruction) Build(regexNode *frontend.Node, ruleIdx int) []State {
	defer trace.StartRegion(
		context.TODO(),
		"github.com/backbone81/golr/internal/scannergen/nfa/ThompsonsConstruction.Build()",
	).End()

	b.mustBeValidRegex(regexNode)
	return b.buildNFAFromRegexValidated(regexNode, ruleIdx, []State{})
}

// buildNFAFromRegexValidated constructs an NFA from a regular expression by applying Thomson's construction.
func (b *ThompsonsConstruction) buildNFAFromRegexValidated(
	regexNode *frontend.Node,
	ruleIdx int,
	states []State,
) []State {
	startStateIdx := len(states)
	result := b.buildNFAFromRegex(regexNode, ruleIdx, states)
	b.MustBeValidNFA(result[startStateIdx:])
	return result
}

// buildNFAFromRegex dispatches the NFA construction to a typed function for that regular expression node.
func (b *ThompsonsConstruction) buildNFAFromRegex(regexNode *frontend.Node, ruleIdx int, states []State) []State {
	switch regexNode.Kind {
	case frontend.KindOr:
		return b.fromOr(&regexNode.Or, ruleIdx, states)
	case frontend.KindAny:
		return b.fromAny(&regexNode.Any, ruleIdx, states)
	case frontend.KindCharClass:
		return b.fromCharClass(&regexNode.CharClass, ruleIdx, states)
	case frontend.KindConcat:
		return b.fromConcat(&regexNode.Concat, ruleIdx, states)
	case frontend.KindLiteral:
		return b.fromLiteral(&regexNode.Literal, ruleIdx, states)
	case frontend.KindOneOrMore:
		return b.fromOneOrMore(&regexNode.OneOrMore, ruleIdx, states)
	case frontend.KindOptional:
		return b.fromOptional(&regexNode.Optional, ruleIdx, states)
	case frontend.KindRepetition:
		return b.fromRepetition(&regexNode.Repetition, ruleIdx, states)
	case frontend.KindZeroOrMore:
		return b.fromZeroOrMore(&regexNode.ZeroOrMore, ruleIdx, states)
	default:
		// This is a logic error in the builder which justifies a panic here.
		panic(fmt.Sprintf("unexpected regex node kind: %T", regexNode.Kind))
	}
}

// mustBeValidRegex will panic if the given regular expression is not valid.
func (b *ThompsonsConstruction) mustBeValidRegex(regexNode *frontend.Node) {
	if err := regexNode.Validate(); err != nil {
		// The builder can assume to be only called with valid regular expressions. Invalid regular expressions
		// can be considered a logic error in the application. Therefore, calling panic is justified here.
		panic(err)
	}
}

// MustBeValidNFA will panic if the given NFA is not valid.
func (b *ThompsonsConstruction) MustBeValidNFA(states []State) {
	if err := b.validateThompsonNFA(states); err != nil {
		// The builder is expected to construct NFAs with certain restrictions. Failure to do so can be considered
		// a logic error in the builder. Therefore, calling panic is justified here.
		panic(err)
	}
}

// validateThompsonNFA makes sure that assumptions we work with in this package are satisfied.
func (b *ThompsonsConstruction) validateThompsonNFA(states []State) error {
	var acceptingStateCount int
	var acceptingStateIdx int
	for i := range states {
		if states[i].Accept {
			acceptingStateCount++
			acceptingStateIdx = i
		}
	}
	if acceptingStateCount != 1 {
		return errors.New("NFA requires exactly one accepting state")
	}
	if acceptingStateIdx == 0 {
		return errors.New("NFA start state is required to be different from accepting state")
	}
	return nil
}
