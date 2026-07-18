package conflict

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// PrecedencePolicy resolves a shift/reduce conflict from the precedence and the associativity which the grammar
// declares. The conflicted terminal is the terminal which would be shifted, and the production of a reduction inherits
// its precedence from a terminal, see ProductionPrecedence.
//
// The reduction wins when the production binds tighter than the terminal, and the shift wins when the terminal binds
// tighter than the production. When both bind equally tight, the associativity of the terminal decides: a left
// associative terminal reduces, a right associative terminal shifts, and a nonassociative terminal rejects the
// conflicted terminal, which removes both actions and makes the parser report an error.
//
// The policy only decides between a shift and a reduction, never between two reductions, which the grammar has no way
// of expressing with precedence declarations. It leaves a reduction and the shift untouched whenever the declarations
// do not decide between them, so a compound policy needs the policies behind this one to decide those conflicts.
type PrecedencePolicy struct {
	// grammar is where the precedence and the associativity of the terminals and the productions come from.
	grammar frontend.Grammar

	// lastResolveRejecters and lastResolveShiftBeaters record why the most recent Resolve decided the conflict as it did.
	// Resolve fills them while it narrows the candidates, so that ContributeSplitStability can judge whether the narrowing
	// is split-stable without deciding the conflict a second time. They are only meaningful right after a Resolve call and
	// hold the reasons of that call only; ContributeSplitStability is the single reader and reads them immediately after
	// the Resolve it triggers.
	//
	// lastResolveRejecters are the reductions which reject the terminal, which turns the conflict into an error action.
	// The error holds for every isocore only when a rejecting reduction is an always contribution.
	lastResolveRejecters []Contribution

	// lastResolveShiftBeaters are the reductions which beat the shift, so that the shift is removed. It is only meaningful
	// when there is no rejecter, and the shift removal holds for every isocore only when such a reduction is an always
	// contribution.
	lastResolveShiftBeaters []Contribution
}

// PrecedencePolicy implements Policy.
var _ Policy = (*PrecedencePolicy)(nil)

// NewPrecedencePolicy returns the policy which resolves a shift/reduce conflict from the precedence and the
// associativity declared in the grammar.
func NewPrecedencePolicy(grammar frontend.Grammar) *PrecedencePolicy {
	return &PrecedencePolicy{
		grammar: grammar,
	}
}

// Resolve decides every reduction of the candidates against the shift of the conflicted terminal.
//
// A conflict which holds more than one reduction is decided reduction by reduction: a reduction which loses against the
// shift is removed, and the shift only survives when it lost against no reduction at all. That can leave a single
// reduction which beat the shift, or several candidates which the policies behind this one have to decide between. As
// soon as one reduction asks for the terminal to be rejected, the whole conflict is decided that way, because a
// rejected terminal leaves no action for any other reduction to win.
func (p *PrecedencePolicy) Resolve(terminalIdx int, candidates ContributionSet) ContributionSet {
	// The reasons of any earlier Resolve are reset to zero length, so that they hold the reasons of this call only while
	// reusing the memory the earlier call already allocated.
	p.lastResolveRejecters = p.lastResolveRejecters[:0]
	p.lastResolveShiftBeaters = p.lastResolveShiftBeaters[:0]

	shift := NewShiftContribution()
	if !candidates.Contains(shift) {
		// This is a conflict between reductions only, which precedence declarations cannot express.
		return candidates
	}
	terminal := p.grammar.Terminals[terminalIdx]
	if IsNoPrecedence(terminal.Precedence) {
		// The terminal has no precedence declared, so there is nothing to compare the productions against.
		return candidates
	}

	shiftRemoved := false
	var result ContributionSet
	for _, candidate := range candidates.All() {
		if candidate.IsShiftAction() {
			// The shift is added at the end, once we know whether any reduction beat it.
			continue
		}

		survivors := p.resolveShiftReduce(terminal.Precedence, terminal.Associativity, shift, candidate)
		switch {
		case survivors.IsEmpty():
			// The reduction rejects the terminal. Every reduction is still decided so that ContributeSplitStability
			// learns all the rejecters, even though a single rejecter already turns the conflict into an error action.
			p.lastResolveRejecters = append(p.lastResolveRejecters, candidate)
		case !survivors.Contains(candidate):
			// The reduction lost against the shift, so it is removed by being left out of the result.
		case !survivors.Contains(shift):
			// The reduction beat the shift, so it stays a candidate and the shift is gone for good, even when another
			// reduction does not decide against it.
			p.lastResolveShiftBeaters = append(p.lastResolveShiftBeaters, candidate)
			result.Add(candidate)
			shiftRemoved = true
		default:
			// The declarations did not decide, so the reduction stays a candidate alongside the shift.
			result.Add(candidate)
		}
	}

	if len(p.lastResolveRejecters) > 0 {
		// The terminal is rejected in this state, so every action for it is removed.
		return ContributionSet{}
	}
	if !shiftRemoved {
		result.Add(shift)
	}
	return result
}

// ContributeSplitStability resolves the conflict through Resolve, which both narrows the bookkeeping and records why it
// decided as it did, and then reads those reasons to decide whether the narrowing is split-stable.
//
// The classification of a reduction against the shift depends only on the precedence and associativity of that
// reduction and the terminal, never on the other reductions, so it is the same in every isocore. What varies between
// the isocores is only which reductions are present, so the split stability of each narrowing comes down to whether the
// reductions which cause it are always contributions:
//
//   - A reduction which loses to the shift is removed whether or not an isocore makes it, so removing it never
//     threatens split stability, which is why Resolve records no reason for it.
//   - A rejecting reduction turns the whole conflict into an error action. That happens in an isocore exactly when the
//     isocore makes a rejecting reduction, so it holds for every isocore only when a rejecting reduction is an always
//     contribution.
//   - A reduction which beats the shift removes it. The shift is removed in an isocore exactly when the isocore makes a
//     reduction which beats it, so it holds for every isocore only when such a reduction is an always contribution.
//
// This reasoning relies on the shift being an always contribution, which it always is: point 1 of definition 3.30 of
// IELR(1) makes a shift's contribution matrix row undefined, which is an always contribution by point 2(a) of
// definition 3.28, because splitting a state keeps its transitions so every isocore makes the shift. A potential shift
// would make precedence's narrowing itself conditional, because precedence only decides a conflict which has a shift:
// in the isocores which did not make the shift, the reductions it removed as losers would survive and could change the
// dominant contribution. Since a potential shift cannot occur, this is not handled.
func (p *PrecedencePolicy) ContributeSplitStability(terminalIdx int, splitStability *SplitStability) {
	// Resolve narrows the candidates and fills lastResolveRejecters and lastResolveShiftBeaters, which are read right
	// after with no other Resolve call in between.
	splitStability.remaining = p.Resolve(terminalIdx, splitStability.remaining)
	switch {
	case len(p.lastResolveRejecters) > 0:
		if !splitStability.anyAlways(p.lastResolveRejecters) {
			splitStability.markUnstable()
		}
	case len(p.lastResolveShiftBeaters) > 0:
		if !splitStability.anyAlways(p.lastResolveShiftBeaters) {
			splitStability.markUnstable()
		}
	}
}

// resolveShiftReduce decides a single shift/reduce conflict between the shift of the conflicted terminal and the
// reduction of a single production, and returns the contributions which survive that decision. The precedence level and
// the associativity of the conflicted terminal are what the reduction is decided against, and they are passed in
// separately because they decide different things: the levels decide which action binds tighter, and the associativity
// only comes into play once the levels turn out to be equal.
//
// This is the same narrowing a Policy does, applied to a conflict of exactly two contributions: both survive when the
// declarations do not decide, the winner survives alone when they do, and neither survives when the declarations reject
// the terminal.
func (p *PrecedencePolicy) resolveShiftReduce(
	terminalPrecedence int,
	terminalAssociativity frontend.Associativity,
	shift Contribution,
	reduce Contribution,
) ContributionSet {
	undecided := NewContributionSet(shift, reduce)

	productionPrecedence := ProductionPrecedence(p.grammar, reduce.ProductionIdx())
	if IsNoPrecedence(productionPrecedence) {
		// The production has no precedence declared, so there is nothing to compare the terminal against.
		return undecided
	}

	if productionPrecedence > terminalPrecedence {
		// The production binds tighter than the terminal, so what is on the stack is reduced before the terminal is
		// shifted.
		return NewContributionSet(reduce)
	}
	if productionPrecedence < terminalPrecedence {
		// The terminal binds tighter than the production, so the terminal is shifted into the production which follows.
		return NewContributionSet(shift)
	}

	// Both bind equally tight, so the way the terminal associates with itself decides.
	switch terminalAssociativity {
	case frontend.AssociativityLeft:
		// The terminal groups to the left, so what is on the stack is reduced before the terminal is shifted.
		return NewContributionSet(reduce)
	case frontend.AssociativityRight:
		// The terminal groups to the right, so the terminal is shifted and what is on the stack waits for it.
		return NewContributionSet(shift)
	case frontend.AssociativityNone:
		// The terminal does not associate at all, so an input which needs it to associate is an error.
		return ContributionSet{}
	case frontend.AssociativityUndeclared:
		// The terminal has a precedence, but no associativity was declared for it. There is nothing left to decide the
		// conflict with, so both actions stay candidates.
		return undecided
	}
	return undecided
}
