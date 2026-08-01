# Correctness

Generating LR(1) parser tables is not the kind of code where a bug announces itself. A reduction lookahead set which is
one terminal too large silently turns a well-formed grammar into one with a conflict. A lookahead set which is one
terminal too small produces a parser that builds and runs and rejects a sentence the grammar clearly derives. Nothing
crashes, no test of the generator itself notices, and the damage only shows up much later in whoever uses the generated
parser.

IELR(1) is especially exposed to this. It is a five phase algorithm which deliberately produces a parser table that is
*not* structurally comparable to any table you could easily derive by hand: it starts from LALR(1), computes annotations
describing which lookaheads a state contributes to a conflict elsewhere, and then splits exactly those states where the
LALR(1) merge was harmful. The result has a state count somewhere between LALR(1) and canonical LR(1), a state numbering
of its own, and a splitting granularity nobody can predict by inspection. So "diff it against the expected table" is not
available as a general test strategy.

What IELR(1) does guarantee is *behavioral*, and that is what the verification is built on:

> An IELR(1) parser accepts the same language and produces the same parses as a canonical LR(1) parser for the same
> grammar under the same conflict resolution policy — with fewer states.

That statement can be tested exhaustively, and the sections below describe the layers which do it. They deliberately
overlap: each layer catches a different class of bug, and the cheap ones run on every `make test` while the expensive
ones run for hours.

## Layer 1: The Grammars from the Paper

The [IELR(1) paper](https://doi.org/10.1016/j.scico.2009.08.001) contains a handful of small grammars, each constructed
to isolate one specific phenomenon: the unambiguous grammar of figure 1, the ambiguous grammar of figure 2, and the goto
follows grammars of figures 5 and 6. These are kept as test grammars, together with a grammar which has the "mysterious"
reduce/reduce conflict — a conflict LALR(1) has and canonical LR(1) does not.

For these grammars the tests do not compare behavior, they pin down internals against the definitions in the paper:

- The **full parser table** of the reduce/reduce grammar is pinned, with default reductions switched off so every reduce
  action is listed with its own lookahead. This is the sharpest available check on phase 4: the two isocores of the
  split state must end up with exactly the mirrored one-terminal lookahead sets their respective predecessors generate.
  A behavioral test cannot see a lookahead set that is too *small* as easily, because a missing terminal produces no
  conflict — it just makes the parser reject valid input at a place a random sentence may never reach.
- The **follow kernel items** of definition 3.16 are pinned for grammars which exercise both halves of the reflexive
  transitive closure the definition asks for: a goto depending on a kernel item of its own state, and a goto depending
  on kernel items reached through a chain of internal relations.
- The **annotations** of phase 2 are pinned for the grammar of figure 5, item lookahead sets and inadequacies included,
  along with the annotation being discarded again when conflict resolution removes the conflict it was tracking, and the
  iteration terminating on cyclic lanes.
- The **division of labor between the phases** is pinned: for the ambiguous grammar of figure 2, the raw table produced
  by the splitting phases must still carry the genuine conflict, and only the table which went through phase 5 conflict
  resolution is free of it. Splitting must not "fix" a conflict canonical LR(1) has too.

## Layer 2: Real-World Grammars Cross-Checked Against GNU Bison

Small grammars from a paper do not reach the scale where an algorithm's complexity shows. The `testdata` directory
therefore holds the full grammar files of a set of real languages:

| Grammar             | Terminals | Nonterminals | Productions | LALR(1)? |
|---------------------|-----------|--------------|-------------|----------|
| GNU Bison 3.8.2     | 59        | 38           | 119         | yes      |
| GCC 2.95.3 C        | 82        | 117          | 364         | yes      |
| GCC 2.95.3 Obj-C    | 82        | 162          | 502         | yes      |
| GCC 3.3.6 C++       | 112       | 238          | 871         | **no**   |
| GCC 4.2.4 Java      | 110       | 153          | 505         | yes      |
| Go 1.5.4            | 74        | 127          | 337         | yes      |
| PHP 8.6.7           | 184       | 177          | 623         | yes      |
| PostgreSQL 18.4     | 540       | 733          | 3434        | **no**   |

Every one of them is built four ways in the test suite: with the GoLR LALR(1) core, the GoLR IELR(1) core, the Bison
LALR(1) core and the Bison IELR(1) core. The Bison-backed cores are not a reimplementation — they shell out to GNU
Bison, the reference implementation whose authors wrote the IELR(1) paper, so the comparison is against the real thing.

The assertions depend on whether the grammar is LALR(1):

- For a **LALR(1) grammar** there is nothing for IELR(1) to split, so all four tables must have the identical state
  count. This catches both directions of failure at once: GoLR's LALR(1) core disagreeing with Bison's, and GoLR's
  IELR(1) core splitting a state which does not need splitting.
- For a **non-LALR(1) grammar** — GCC's C++ grammar and the PostgreSQL grammar — splitting must actually happen in both
  implementations, and GoLR's IELR(1) state count must be at least Bison's and within 2% of it.

That 2% tolerance is honest about a known gap rather than hiding it: GoLR's IELR(1) does not yet reproduce Bison's state
count exactly on the largest non-LALR(1) grammars. The excess states are a *quality* difference, not a correctness one —
the tables agree on the language and the parses, GoLR's is merely a little less minimal than it could be. The
suboptimum state merging of section 3.8 of the paper is the likely place where the remaining difference lives.

These grammars double as the benchmark corpus, so a performance regression on a real grammar shows up in the same place.

## Layer 3: Differential Testing Against Canonical LR(1)

This is the main line of defense and the one which finds bugs nobody thought to write a test for.

### Canonical LR(1) as the oracle

Canonical LR(1) is the right oracle because it is *simple*. Building it needs nothing but item closures and first sets —
no goto follows digraph, no annotations, no isocore compatibility test, none of the machinery which makes IELR(1) hard
to get right. The two implementations share almost nothing, so they are very unlikely to be wrong in the same way. It is
far too expensive to ship as a general purpose core for large grammars, which is precisely why IELR(1) exists, but for
small generated grammars it is perfectly practical.

The comparison is behavioral, for the reason given at the top: the two tables are not supposed to look alike. Both are
built with conflicts resolved under the same default policy, and both with the default-reduction compaction switched
off, so what is compared is the canonical resolved table of each side.

### Random grammars aimed at the hard shapes

A random grammar generator would, left to itself, produce mostly grammars which are trivially LALR(1) and exercise none
of the splitting logic. The generator therefore builds each production from one of a set of weighted *scenarios*, chosen
because each one is known to stress a specific relation of the algorithm:

- **Empty** right hand sides, which make nonterminals nullable and drive the always-follows and successor relations of
  the DeRemer–Pennello algorithm.
- **Nonterminal-only** right hand sides, which build the chains that goto follows propagate along.
- **Recursive** productions, which create cycles and reuse a nonterminal in more than one context.
- **Nullable suffix** productions of the form `B -> alpha A gamma` with a guaranteed nullable, non-empty `gamma`. This
  is exactly the *includes* relation, along which a lookahead set is propagated backwards across a nullable gap. Random
  productions produce that relation only by accident; this scenario makes it a reliable part of every corpus.
- **Shared nonterminal**, which reaches one nonterminal from two contexts with different following terminals. This is
  the situation where canonical LR(1) and LALR(1) genuinely diverge, because LALR(1) merges the isocore and unions the
  two lookahead sets.
- **Shared nonterminal with a nullable gap**, the same situation but with a nullable nonterminal between the shared
  nonterminal and the terminal which distinguishes the contexts, so the distinguishing lookahead has to reach the shared
  nonterminal across the gap through the reads and includes relations instead of as a direct read.
- **Reduce/reduce**, which constructs the mysterious reduce-reduce conflict: two nonterminals with a common core
  reduction reached from two contexts with swapped following terminals, which LALR(1) merges into a conflict canonical
  LR(1) does not have.

### Sentences the grammar actually derives

Inputs are not random token streams. Each sentence is produced by a random leftmost derivation of the grammar itself, so
every sentence is by construction a member of the language, and an early rejection from *either* table is a signal in
itself. Termination is guaranteed for any productive grammar: the generator precomputes the minimum derivation height
per nonterminal, and once a shared expansion budget is spent it only chooses productions which achieve that minimum,
which bounds the sentence length no matter how the grammar branches.

### Lockstep comparison, with one allowance

Both tables are driven through the same sentence one LR action at a time, and every action must match: same shift, same
reduction by the same production, same accept, same rejection at the same point.

There is exactly one permitted deviation, and it is a documented property of near-LALR tables rather than a hole in the
test. A near-LALR table may perform some additional harmless reductions on a doomed input before it reports the error
that canonical LR(1) reports slightly earlier. Both still reject the sentence, so the language is unchanged; only the
moment of detection differs. The comparison allows for it in the narrowest possible way: once one side rejects, the
other is drained and **every remaining action must be a reduction** until it rejects too. A shift would mean it actually
accepted a token the other side rejected, and an accept would mean the two disagree on the language — both are reported
as real divergences.

When a divergence is found, both tables are replayed with tracing enabled, so the failure carries the offending
sentence and the full action trace of each side. Reading the two traces against each other pins down the state and the
lookahead where they first parted ways.

### The size invariant

Independently of behavior, every compared grammar must satisfy

```text
|LALR(1) states| <= |IELR(1) states| <= |canonical LR(1) states|
```

Conflict resolution never adds or removes states, so the resolved tables can be compared directly. A table below the
lower bound means a required split was lost; one above the upper bound means splitting ran away past what canonical
LR(1) itself needs. This catches a class of bug the behavioral comparison is blind to, because over-splitting produces a
table which is bigger than it should be but still parses everything correctly.

### Guarding against a corpus that passes vacuously

A differential test over grammars which are all trivially LALR(1) would pass without exercising a single line of the
splitting logic, and would look exactly like a healthy test run. The corpus therefore measures itself and fails if it
degrades:

- More than half of the generated grammars must actually be compared, rather than skipped for exceeding the canonical
  LR(1) state limit.
- The corpus must keep clearing a floor of *discriminating* grammars — grammars where LALR(1) has a conflict canonical
  LR(1) does not. Those are the grammars where splitting earns its keep. The generator yields roughly 65 of them per
  thousand grammars; the floor sits several standard deviations below that, so it is not flaky but still trips long
  before the generator could degrade into producing only trivial grammars.

The number of grammars where splitting actually fired is reported on every run as well.

### Phase 0 has its own oracle

The LALR(1) foundation the whole algorithm is built on is verified separately and *structurally*, because for LALR(1) a
structural comparison is available. The builder under test derives its reduction lookaheads by propagating goto follows
along the digraph relations of DeRemer and Pennello. The oracle reaches the same table by an entirely different route:
it constructs the canonical LR(1) automaton and then merges the states which agree on their cores. Both the kernel items
and the full states — every action with its lookahead — must match exactly, over the same kind of random grammar corpus.

Getting phase 0 wrong is otherwise a nasty failure mode, because every later phase would faithfully compute annotations
and splits from a broken starting table.

## Layer 4: The `golr selftest` Soak Test

The test suite has to finish in seconds. Finding the grammar which breaks a subtle piece of the splitting logic can take
hundreds of thousands of grammars. The `selftest` sub-command exists for that: it runs the very same comparison code as
the behavioral differential test — not a reimplementation of it — across every CPU core, for as long as it is given.

```shell
golr selftest --duration 8h --failure-dir ./selftest-failures
```

A run saturates all cores, reports progress as it goes, and can be interrupted at any point without losing the summary
of what it checked so far. Each grammar is built from a single seed which reconstructs both the grammar and the
sentences it was checked with, so a failure found in hour six of a run on sixteen workers is reproducible on its own.
With `--failure-dir`, a failing grammar is dumped to disk together with the action traces of both tables. Runs of
millions of grammars are routine.

This is a soak test for GoLR itself. It is not a step in generating a parser, and it never runs as part of one.

## Layer 5: Testing the Tests

A test suite that passes proves nothing unless the tests would fail if the code were wrong. For the differential test
this was measured directly, by mutation testing: roughly 50 deliberate, individually targeted bugs were injected into
the implementation — one at a time, each derived from a specific definition in the paper — and the self-test was run
against each mutant to see whether it noticed, and how many grammars it needed to do so.

The outcome was that every mutation which changes a parse was detected. The median mutation was caught within a few
dozen grammars, the most stubborn one within a few thousand — well inside a routine run. Misdirected-edge mutations,
where a relation is computed over the wrong pair of states, were typically caught within one to three grammars.

The handful of mutants that survived turned out on inspection to be genuinely undetectable *by this oracle*, and for a
reason worth stating plainly: they cause **over-splitting**. The resulting parser accepts the same language and produces
the same parses, it just uses more states than it should. That is a quality regression, not a correctness one, and it is
invisible to a behavioral comparison by construction — the size invariant and the Bison state count comparison of layer
2 are the checks which cover that side.

Phase 5, conflict resolution, is not reached by the self-test at all: it operates on precedence and associativity
declarations, which generated grammars do not have. It is covered by the pinned figure 2 grammar of layer 1 and by the
well-known grammars of layer 2, which are full of real precedence declarations.

## What This Does Not Cover

Being explicit about the boundaries matters as much as the coverage itself:

- **Over-splitting is only partly covered.** The behavioral oracle cannot see it, and the size invariant only catches
  the extreme case of exceeding canonical LR(1). The 2% tolerance against Bison on the two large non-LALR(1) grammars is
  the tightest bound currently asserted.
- **Generated grammars carry no precedence or associativity declarations**, so phase 5 is exercised by the hand-picked
  and real-world grammars rather than by the random corpus.
- **The random grammars are small.** They are the right size to reach unusual *shapes* quickly, but the real grammars of
  layer 2 are what covers scale.
