package oracle

import "fmt"

// RunInLockstep drives two parser interpreters over their own inputs one action at a time and compares them action for
// action. It is the behavioral comparison at the heart of the differential test: because an IELR(1) parser accepts the
// same language and produces the same parses as canonical LR(1) under the same conflict-resolution policy, two parser
// tables for the same grammar and input must take the identical sequence of LR actions — with one allowance.
//
// The allowance is error-detection slack. An IELR(1) table is near-LALR, and a near-LALR table may perform extra
// reductions on a doomed input before it reports the error that canonical LR(1) reports earlier. Both still reject the
// input, so the language is the same; only the moment of rejection differs. So once one interpreter rejects, the other
// is drained until it rejects too, requiring every remaining action to be a reduction: a shift would mean it actually
// accepted the token the first table rejected, and an accept would mean the two disagree on the language — both real
// divergences. This mirrors the well-known way LALR(1) delays error detection relative to canonical LR(1).
//
// It returns nil when the two agree under that rule, and otherwise an error describing the step and the differing
// actions, which is all a differential test needs to report a reproducible failure. Each interpreter keeps its own state
// stack and input cursor, so a divergence surfaces as a differing action. The two interpreters are expected to have been
// built from the same input; feeding them different inputs is a programming error the comparison does not try to detect.
func RunInLockstep(a *ParserInterpreter, b *ParserInterpreter) error {
	for step := 0; ; step++ {
		progressedA := a.Next()
		progressedB := b.Next()

		if !progressedA && !progressedB {
			// Both interpreters terminated on the same step and agreed there, so the whole action sequence matched.
			return nil
		}

		if progressedA != progressedB {
			// One interpreter still has actions left while the other has already terminated. A rejection is handled
			// below on the step it happens, so reaching here means they disagreed on a terminal action; report it.
			return fmt.Errorf(
				"diverged at step %d: a=%s (input offset %d), b=%s (input offset %d)",
				step, a.Value(), a.Offset(), b.Value(), b.Offset(),
			)
		}

		actionA, actionB := a.Value(), b.Value()
		if actionA == actionB {
			continue
		}

		// The actions differ. The one legitimate reason is error-detection slack, which applies only when exactly one
		// interpreter rejected here — the other must then finish on reductions alone and reject too. Any other
		// disagreement (a differing shift or reduce while both continue, or one accepting while the other rejects) is a
		// real divergence. Two rejects on the same step are equal actions and never reach here.
		switch {
		case actionA.Kind == ParserActionReject:
			return drainReductionsUntilReject("b", b, actionB, step)
		case actionB.Kind == ParserActionReject:
			return drainReductionsUntilReject("a", a, actionA, step)
		default:
			return fmt.Errorf(
				"diverged at step %d: a=%s (input offset %d), b=%s (input offset %d)",
				step, actionA, a.Offset(), actionB, b.Offset(),
			)
		}
	}
}

// drainReductionsUntilReject consumes the remaining actions of the interpreter the other one has already rejected past,
// requiring every one to be a reduction until it rejects too. current is the action the interpreter already produced on
// the diverging step, and step and label identify it in an error. It returns nil once the interpreter rejects, and an
// error the moment it does anything but reduce — the error-detection slack an IELR(1) table is allowed is reductions
// only, never a shift (which would mean the token was valid after all) or an accept (which would mean the two tables
// disagree on the language).
func drainReductionsUntilReject(label string, interpreter *ParserInterpreter, current ParserAction, step int) error {
	for {
		switch current.Kind {
		case ParserActionReject:
			return nil
		case ParserActionReduce:
			// The allowed slack: a harmless reduction on the way to the same rejection. Keep draining.
		default:
			return fmt.Errorf(
				"diverged at step %d: %s took %s after the other table rejected, expected only reductions until it rejects too (input offset %d)",
				step, label, current, interpreter.Offset(),
			)
		}

		if !interpreter.Next() {
			// A reduction is never a terminal action, so Next keeps returning true until the interpreter rejects or
			// accepts. Running out here would mean a malformed interpreter, so report it rather than pass silently.
			return fmt.Errorf(
				"diverged at step %d: %s ran out of actions without rejecting after the other table rejected",
				step, label,
			)
		}
		current = interpreter.Value()
		step++
	}
}
