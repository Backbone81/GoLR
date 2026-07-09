/*
    This grammar is the goto follows caveats grammar from the IELR(1) paper in Fig. 6 on page 17 (or 959).

    Create a grammar report with:

        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/goto-follows-caveats-test-grammar-fig6-lalr1.txt --verbose -Wno-other --define=lr.type=lalr "scripts/bison-grammars/goto-follows-caveats-test-grammar-fig6.y"
        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/goto-follows-caveats-test-grammar-fig6-ielr1.txt --verbose -Wno-other --define=lr.type=ielr "scripts/bison-grammars/goto-follows-caveats-test-grammar-fig6.y"

*/

%token
    TOKEN_A "a"
    TOKEN_B "b"

%%

S
    : "a" A "a"
    | "a" "a" "b"
    | "b" A "b"
    ;

A
    : B C
    ;

B
    : "a"
    ;

C
    : D
    ;

D
    : %empty
    ;
