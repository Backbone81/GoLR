/*
    This grammar is the unambiguous grammar from the IELR(1) paper in Fig. 1 on page 3 (or 945).

    Create a grammar report with:

        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/unambiguous-test-grammar-fig1-lalr1.txt --verbose -Wno-other --define=lr.type=lalr "scripts/bison-grammars/unambiguous-test-grammar-fig1.y"
        bin/bison --header=/dev/null --output=/dev/null --report-file=scripts/bison-grammars/unambiguous-test-grammar-fig1-ielr1.txt --verbose -Wno-other --define=lr.type=ielr "scripts/bison-grammars/unambiguous-test-grammar-fig1.y"

*/

%token
    TOKEN_A "a"
    TOKEN_B "b"

%%

S
    : "a" A "a"
    | "b" A "b"
    ;

A
    : "a"
    | "a" "a"
    ;
