# Grammar Checks

The parser generator checks every grammar before it builds the parser, no matter which frontend or core is used.

## Unproductive nonterminals

A nonterminal is unproductive when it cannot derive any finite sequence of terminals. This is an error, because no
input can ever match it. The usual cause is a recursion without a base case:

```
@parser {
    document
        : list
        ;

    list
        : list ITEM
        ;
}
```

```text
Error: start nonterminal "document" does not derive any finite string
nonterminal "list" does not derive any finite string
```

Add an alternative which ends the recursion, e.g. `| ITEM` or `| @empty`. When the start nonterminal is unproductive,
the grammar accepts no input at all.

A nonterminal which is used but never defined has no productions and is reported as
`nonterminal "name" has no productions`.

## Unreachable nonterminals

A nonterminal is unreachable when the start nonterminal never derives it. This is a warning:

```text
warning: nonterminal "unused" is unreachable from the start nonterminal "list"
```

The nonterminal stays in the grammar and in the generated parser. Remove it, or use it from a production of the
grammar.

## Warnings

`golr parser` writes warnings to stderr and still generates the parser. With `--fail-on-warnings`, every warning
fails the run instead.

Besides unreachable nonterminals, the parser generator warns about productions which are never reduced after conflict
resolution, see [conflicts](parsergen-conflicts.md#productions-which-are-never-reduced).
