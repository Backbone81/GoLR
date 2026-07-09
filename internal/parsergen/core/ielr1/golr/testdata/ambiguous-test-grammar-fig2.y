/*
    This grammar is the ambiguous grammar from the IELR(1) paper in Fig. 2 on page 3 (or 945).

    Create a grammar report with:

        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/ambiguous-test-grammar-fig2-lalr1.txt --verbose -Wno-other --define=lr.type=lalr "scripts/bison-grammars/ambiguous-test-grammar-fig2.y"
        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/ambiguous-test-grammar-fig2-ielr1.txt --verbose -Wno-other --define=lr.type=ielr "scripts/bison-grammars/ambiguous-test-grammar-fig2.y"

*/

%token
    TOKEN_A "a"
    TOKEN_B "b"
    TOKEN_C "c"

%%

S
    : "a" A "a"
    | "a" B "b"
    | "a" C "c"
    | "b" A "b"
    | "b" B "a"
    | "b" C "a"
    ;

A
    : "a" "a"
    ;

B
    : "a" "a"
    ;

C
    : "a" "a"
    ;
