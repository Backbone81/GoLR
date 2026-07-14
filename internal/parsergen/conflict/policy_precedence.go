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
		if survivors.IsEmpty() {
			// The terminal is rejected in this state, so every action for it is removed.
			return ContributionSet{}
		}
		if survivors.Contains(candidate) {
			// The reduction either beat the shift or was not decided against it, so it stays a candidate.
			result.Add(candidate)
		}
		if !survivors.Contains(shift) {
			// The reduction beat the shift. The shift is gone for good once a single reduction beat it, even when
			// another reduction does not decide against it.
			shiftRemoved = true
		}
	}
	if !shiftRemoved {
		result.Add(shift)
	}
	return result
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
