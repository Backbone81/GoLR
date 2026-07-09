/*
    This grammar is the classic grammar which is LR(1) but not LALR(1). LALR(1) merges the two "c" states because they
    share the same core, which merges their lookahead sets and produces a reduce/reduce conflict on both "d" and "e".
    It is used to verify that the LALR(1) builder faithfully produces overlapping reduction lookahead sets for grammars
    with a genuine LALR(1) conflict.

    Create a grammar report with:

        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/reduce-reduce-conflict-test-grammar-lalr1.txt --verbose -Wno-other --define=lr.type=lalr "scripts/bison-grammars/reduce-reduce-conflict-test-grammar.y"
        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/reduce-reduce-conflict-test-grammar-ielr1.txt --verbose -Wno-other --define=lr.type=ielr "scripts/bison-grammars/reduce-reduce-conflict-test-grammar.y"

*/

%token
    TOKEN_A "a"
    TOKEN_B "b"
    TOKEN_C "c"
    TOKEN_D "d"
    TOKEN_E "e"

%%

S
    : "a" A "d"
    | "b" B "d"
    | "a" B "e"
    | "b" A "e"
    ;

A
    : "c"
    ;

B
    : "c"
    ;
