/*
    This grammar is the goto follows grammar from the IELR(1) paper in Fig. 5 on page 13 (or 955).

    Create a grammar report with:

        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/goto-follows-test-grammar-fig5-lalr1.txt --verbose -Wno-other --define=lr.type=lalr "scripts/bison-grammars/goto-follows-test-grammar-fig5.y"
        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/goto-follows-test-grammar-fig5-ielr1.txt --verbose -Wno-other --define=lr.type=ielr "scripts/bison-grammars/goto-follows-test-grammar-fig5.y"

*/

%token
    TOKEN_A "a"
    TOKEN_B "b"
    TOKEN_C "c"

%%

S
    : "a" A B "a"
    | "b" A B "b"
    ;

A
    : "a" C D E
    ;

B
    : "c"
    | %empty
    ;

C
    : D
    ;

D
    : "a"
    ;

E
    : "a"
    | %empty
    ;
