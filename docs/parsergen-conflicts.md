# Conflicts

A conflict is a state of the parser where more than one action is possible on the same terminal: a shift and a
reduction (shift/reduce conflict), or two reductions (reduce/reduce conflict). The parser generator has to decide on
one action for every such state.

## How conflicts are decided

The rules apply in this order:

1. **Precedence and associativity.** The declarations of the `@precedence` section (see the
   [GoLR frontend](parsergen-frontend-golr.md#precedence-section)) decide first. They are part of the grammar, so they
   always apply, and a conflict they decide is not reported.
2. **Shift over reduce.** A shift wins over every reduction.
3. **Earliest production.** Of several reductions, the production declared first in the grammar wins.

That is the behavior of GNU Bison and Yacc, and it is the default. Rules 2 and 3 are rules of last resort. Each of the
flags below switches one of them off, and a conflict which is left open then fails the run:

| Flag | Shift over reduce | Earliest production |
|---|---|---|
| none | applies | applies |
| `--fail-on-sr-conflicts` | fails | applies |
| `--fail-on-rr-conflicts` | applies | fails |
| `--fail-on-conflicts` | fails | fails |

`--fail-on-conflicts` is the same as passing both of the other flags.

A conflict is classified by the actions left once the rules have run, not by the actions which competed at first. A
shift competing with two reductions is decided by shift over reduce alone, so `--fail-on-rr-conflicts` accepts it.
Under `--fail-on-sr-conflicts`, earliest production still removes the later reduction, and the shift/reduce conflict
between the shift and the earliest reduction fails the run. Under `--fail-on-conflicts`, nothing removes any of the
three, so the conflict is both a shift/reduce and a reduce/reduce conflict.

The IELR(1) core builds the parser tables under the selected rules. With a fail flag it keeps states apart which it
would otherwise merge, so the report only names conflicts a canonical LR(1) parser has as well. For a grammar which
builds successfully, the flags do not change the parser.

With a fail flag, the `lalr1` core can fail where the `ielr1` and `lr1` cores succeed. LALR(1) merges states which
canonical LR(1) keeps apart, and a merged state can have a reduce/reduce conflict which canonical LR(1) does not have.

The `*-bison` cores leave conflict resolution to GNU Bison and do not support the flags. They fail with an error when one
is passed.

## The conflict report

`golr parser` writes the report to stderr. It starts with one line per kind of conflict, which counts the conflicts
the rules of last resort decided and the ones left unresolved:

```text
5 shift/reduce conflicts resolved
```

A grammar without such conflicts prints nothing. `--verbose` adds every resolved conflict with the state it occurs in
and the chosen action, and `--with-state-number` adds the state number to every state.

A failing run writes the counts, then every unresolved conflict, and exits with a non-zero exit code without writing a
parser:

```text
1 shift/reduce conflict unresolved

state:
  statement -> "if" expression "then" statement •
  statement -> "if" expression "then" statement • "else" statement

  shift/reduce conflict on terminal "else":
    shift
    reduce: statement -> "if" expression "then" statement

Error: 1 unresolved conflict
```

A failing run does not list the resolved conflicts. Fix the unresolved ones first.

## Productions which are never reduced

Deciding a conflict can remove every reduction of a production. The parser then never reduces it, and the parser
generator warns about it:

```
@parser {
    expression
        : variable
        | function
        ;

    variable
        : NAME
        ;

    function
        : NAME
        ;
}
```

```text
warning: production function -> NAME is never reduced after conflict resolution
1 reduce/reduce conflict resolved
```

Earliest production decides the reduce/reduce conflict on `NAME` for `variable`, so `function` never matches. A
precedence declaration can have the same effect when it decides every conflict of a production against it.

The production stays in the grammar and in the generated parser. Only the production which lost its reductions is
reported, not the productions which can no longer be reduced as a consequence, like one which contains `function`.

## Keeping conflicts under control

Depending on how restrictive you want to deal with conflicts in your grammar, choose one of the following approaches:

- A grammar which should have no conflicts beyond those decided by precedence builds with `--fail-on-conflicts`.
- A grammar which relies on shift over reduce, like the dangling `else`, but should never have a reduce/reduce conflict
  builds with `--fail-on-rr-conflicts`.
- A grammar which should not lose any production to conflict resolution builds with `--fail-on-warnings`, which also
  fails on the other [warnings](parsergen-grammar-checks.md#warnings).
- To track the accepted conflicts themselves, check the output of `--verbose` into source control and diff it in CI.
  The report is stable across grammar changes which do not touch the conflicts, so a diff shows exactly which conflict
  was added or removed.
