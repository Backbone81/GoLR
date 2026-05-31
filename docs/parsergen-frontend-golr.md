# Parser Generator Frontend: GoLR

This frontend describes the context free grammar of a parser as a GoLR grammar document. This is a custom format
specifically designed for this library.

The goal for designing this format was to make things as explicit as possible and to not have any automatic or implicit
mechanics. We want users which are unfamiliar with the format to quickly understand the grammar. All tokens need to be
declared and given a technical name. This provides reliable and good names when generating code in the backend.

See the `ide` folder for IDE extensions providing syntax highlighting for GoLR files.

## Basic Structure

A GoLR grammar file consists of exactly two top-level sections: a scanner section and a parser section.

```
@scanner {
    // token declarations
}

@parser {
    // production rules
}
```

Comments can be line comments with `//` until the end of the line, or block comments which contain everything between
`/*` and `*/`.

Here is a small but complete example for a grammar that parses simple arithmetic expressions:

```
@scanner {
    NUMBER: /[0-9]+/;
    PLUS:   "+";
    MINUS:  "-";
    STAR:   "*";
    SLASH:  "/";
    LPAREN: "(";
    RPAREN: ")";
}

@parser {
    @start: expr;

    @precedence {
        @left:  "+" "-";
        @left:  "*" "/";
    }

    expr: expr "+"  expr;
    expr: expr "-" expr;
    expr: expr "*"  expr;
    expr: expr "/" expr;
    expr: "(" expr ")";
    expr: NUMBER;
}
```

## `@scanner` Section

The scanner section declares the terminal symbols (tokens) of the grammar. Every token used in the parser section
must be declared here.

### Token Declaration

```
NAME: pattern;
```

`NAME` can be any name starting with a-z, A-Z or an underscore, followed by any number of characters a-z, A-Z, 0-9 and
underscore. The name is case-sensitive and uniquely identifies the token. The pattern is one of the three forms
described below.

### Regular Expression Pattern

```
NUMBER: /[0-9]+/;
```

The pattern is a regular expression delimited by `/`. The regex supports the following constructs:

| Construct       | Description                                                                |
|-----------------|----------------------------------------------------------------------------|
| `abc`           | Literal characters                                                         |
| `.`             | Any single character                                                       |
| `[abc]`         | Character class                                                            |
| `[a-z]`         | Character range inside a class                                             |
| `[^abc]`        | Negated character class                                                    |
| `\d`            | Digit shorthand — equivalent to `[0-9]`                                    |
| `\w`            | Word character shorthand — equivalent to `[a-zA-Z0-9_]`                    |
| `\s`            | Whitespace shorthand — equivalent to `[ \t\n\r\f\v]`                       |
| `\n \t \r` etc. | Escape sequences for special characters                                    |
| `a\|b`          | Alternation — matches `a` or `b`                                           |
| `a*`            | Zero or more                                                               |
| `a+`            | One or more                                                                |
| `a?`            | Optional                                                                   |
| `a{n}`          | Exactly `n` repetitions                                                    |
| `a{n,m}`        | Between `n` and `m` repetitions                                            |
| `(a)`           | Grouping / subexpression                                                   |
| `{NAME}`        | Fragment reference — expands to the pattern of the named `@fragment` token |
| `[[:name:]]`    | POSIX character class (see table below)                                    |

A `{NAME}` construct is recognized as a fragment reference when the first character after `{` is a letter or
underscore. A `{` followed by a digit or comma is always treated as a repetition quantifier.

POSIX character classes are Unicode-aware. The following named classes are supported:

| Class          | Matches                                                                                             |
|----------------|-----------------------------------------------------------------------------------------------------|
| `[[:alnum:]]`  | Unicode letters, letter numbers, and decimal digits (`L`, `Nl`, `Nd`)                               |
| `[[:alpha:]]`  | Unicode letters and letter numbers (`L`, `Nl`)                                                      |
| `[[:ascii:]]`  | ASCII characters (`U+0000`–`U+007F`)                                                                |
| `[[:blank:]]`  | Unicode space separators (`Zs`) plus horizontal tab (`\t`)                                          |
| `[[:cntrl:]]`  | Unicode control characters (`Cc`)                                                                   |
| `[[:digit:]]`  | Unicode decimal digit numbers (`Nd`)                                                                |
| `[[:graph:]]`  | All characters except Unicode separators (`Z`) and other (`C`)                                      |
| `[[:lower:]]`  | Unicode lowercase letters (`Ll`)                                                                    |
| `[[:print:]]`  | All characters except Unicode other (`C`)                                                           |
| `[[:punct:]]`  | Unicode punctuation (`P`) plus `$`, `+`, `<`, `=`, `>`, `^`, `` ` ``, `\|`, `~`                     |
| `[[:space:]]`  | Unicode separators (`Z`) plus ASCII whitespace controls (`\t`, `\n`, `\v`, `\f`, `\r`)              |
| `[[:upper:]]`  | Unicode uppercase letters (`Lu`)                                                                    |
| `[[:word:]]`   | Unicode letters, letter numbers, decimal digits, and connector punctuation (`L`, `Nl`, `Nd`, `Pc`)  |
| `[[:xdigit:]]` | Hexadecimal digits (`0–9`, `A–F`, `a–f`)                                                            |

POSIX classes may be combined with other character ranges inside a single class: `[a-z[:digit:]]`. Negation applies to 
the entire class: `[^[:alpha:]]`.

### String Literal Pattern

```
PLUS: "+";
```

The pattern is a double-quoted string literal. The quoted string also acts as an **alias** for the token: production
rules may reference it by the string literal directly instead of by the token name, which is useful for punctuation
tokens where the operator character is more readable than a technical name.

Each alias must be unique — no two tokens may share the same string literal.

### Empty Pattern

```
INDENT: @empty;
```

Declares a token with no regular expression. This is intended for tokens that are injected by a custom scanning layer 
on top of the base scanner, such as indentation-sensitive tokens.

### Token Annotations

A pattern declaration may be followed by one or more annotations before the semicolon:

```
WHITESPACE: /[ \t\n\r]+/ @skip;
COMMENT: /\/\/.*/ @skip;
DIGIT: /[0-9]/ @fragment;
```

| Annotation  | Meaning                                                                                                                                                  |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `@skip`     | The token is recognized by the scanner but not passed to the parser. Used for whitespace and comments.                                                   |
| `@fragment` | The token is a named helper pattern and is excluded from the grammar terminals. Its pattern can be referenced in other regular expressions using {NAME}. |

### Fragment Tokens

A token annotated with `@fragment` is a named helper pattern. It is not included in the list of scanner
tokens and is invisible to the parser. Its sole purpose is to be referenced inside other regular expressions.

```
DIGIT: /[0-9]/ @fragment;
HEX_DIGIT: /[a-f]{DIGIT}/ @fragment;
HEX_NUMBER: /0x{HEX_DIGIT}+/;
```

Fragment references use the `{NAME}` syntax inside a regular expression pattern.

Fragments may reference other fragments (the references are resolved recursively).

String-literal tokens may also be used as fragments, in which case the quoted text is used literally when
the fragment is expanded.

## `@parser` Section

The parser section describes the context free grammar itself. It optionally contains a start declaration and a
precedence section, followed by one or more production rules.

```
@parser {
    @start: name;         // optional

    @precedence { ... }   // optional

    name: alternative_list;
    ...
}
```

## `@start` Declaration

```
@start: expr;
```

Declares which nonterminal is the start symbol of the grammar — that is, the root of every valid parse tree.

When `@start` is omitted, the first nonterminal encountered in the production rules is used as the start symbol.

The name given to `@start` must be a nonterminal defined by at least one production rule in the same file.

## `@precedence` Section

Operator precedence and associativity are declared in the `@precedence` block. This section is optional and may
be omitted entirely if the grammar has no shift/reduce conflicts that require resolution.

```
@precedence {
    @left:       "+" "-";
    @left:       "*" "/";
    @right:      POWER;
    @none:       "==" "!=" "<" ">";
    @precedence: UMINUS;
}
```

### Precedence Levels

Each line inside `@precedence` declares one precedence level. **Levels declared first have higher precedence
than levels declared later.** All tokens listed on the same line share the same precedence level and associativity.

### Associativity

| Keyword       | Meaning                                                                                                                                                                                                                   |
|---------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `@left`       | Left-associative. `a + b + c` parses as `(a + b) + c`.                                                                                                                                                                    |
| `@right`      | Right-associative. `a = b = c` parses as `a = (b = c)`.                                                                                                                                                                   |
| `@none`       | Non-associative. `a < b < c` is a parse error.                                                                                                                                                                            |
| `@precedence` | Precedence-only. The token is assigned a precedence level but no associativity. Useful for unary operators that need a precedence level for `@precedence(...)` annotations but should not resolve conflicts on their own. |

All tokens referenced in the precedence section must be declared in the `@scanner` section. Tokens referenced
by their string alias (e.g. `"+"`) are also allowed, provided the alias has been declared.

## Production Rules

```
name: symbol symbol symbol;
name: alternative_one | alternative_two | alternative_three;
```

A production rule defines how a nonterminal `name` can be expanded. Multiple alternatives may be listed on
a single rule separated by `|`, or split across multiple rule declarations for the same name — both forms are
equivalent.

### Symbols

Each alternative is a sequence of one or more symbols. A symbol is either:

- A **terminal** — referenced by its token name (e.g. `NUMBER`) or by its string alias (e.g. `"+"`)
- A **nonterminal** — referenced by its name (e.g. `expr`)

### Empty Alternative

```
opt_semicolon: ";" | @empty;
```

`@empty` denotes an empty right-hand side, meaning the nonterminal can derive the empty string.

### Precedence Override

When a production rule is ambiguous due to conflicting precedence, and none of the symbols in the rule carry
the right precedence, the production's effective precedence can be set explicitly:

```
expr: MINUS expr @precedence(UMINUS);
```

The `@precedence(symbol)` annotation overrides the precedence of the production with the precedence of the
given terminal. The terminal must have been assigned a precedence level in the `@precedence` section.

## Constraints

- Every token used in a production rule must be declared in the `@scanner` section.
- Every nonterminal referenced on the right-hand side of a production must have at least one production rule
  defining it as a left-hand side.
- A name cannot be used as both a terminal (declared in `@scanner`) and a nonterminal (used as a production
  left-hand side).
- String aliases must be unique across all token declarations.
- Token names must be unique across all token declarations.
- Fragment names share the same namespace as token names — a name may not be declared as both a token and a fragment.
- Fragment references inside a regular expression (`{NAME}`) must refer to a declared `@fragment` token.
- Cyclic fragment references (where a fragment directly or indirectly references itself) are not allowed.
- The grammar must contain at least one production rule.
