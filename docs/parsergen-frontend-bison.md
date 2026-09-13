# Parser Generator Frontend: Bison

This frontend describes the context free grammar of a parser as a [GNU Bison](https://www.gnu.org/software/bison/) grammar document.

The following functionality is currently supported:

- %token
- %left
- %right
- %nonassoc
- %precedence
- rules
- %prec
- %start

Any not supported functionality is ignored.

The GNU Bison grammar parser is tested against a set of well known GNU Bison grammar files for several programming
languages, to make sure that it works correctly. The well known grammar files include GNU Bison, GCC C, GCC
Objective C, GCC C++, GCC Java, Go and more. See thee `testdata` directory in the repository root for the full set
of grammars.

## Example

The GNU Bison input looks like this:

```text
%token NUMBER
%token LPAREN "("
%token RPAREN ")"

%right UMINUS
%left  PLUS MINUS
%left  STAR SLASH
%nonassoc LT GT
%precedence NOT

%start expr

%%

expr
  : expr PLUS  expr
  | expr MINUS expr
  | expr STAR  expr
  | expr SLASH expr
  | expr LT    expr
  | expr GT    expr
  | NOT expr   %prec NOT
  | MINUS expr %prec UMINUS
  | LPAREN expr RPAREN
  | NUMBER
  ;

stmts
  : %empty
  | stmts stmt
  ;

stmt
  : expr
  ;
```

See the official [GNU Bison documentation](https://www.gnu.org/software/bison/manual/) for details on the grammar
syntax.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/frontend/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                878           1541391 ns/op         1423597 B/op       2872 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                   313           4021261 ns/op         1967210 B/op       9325 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32         234           4866708 ns/op         2219996 B/op      12083 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                  182           6496936 ns/op         4092436 B/op      21614 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                 228           4916663 ns/op         3509716 B/op      11442 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                       348           3115382 ns/op         1780297 B/op       8271 allocs/op
BenchmarkToGrammar/PHP_8.6.7-32                      222           4761388 ns/op         2455905 B/op      15639 allocs/op
BenchmarkToGrammar/PostgreSQL_18.4-32                 75          14859096 ns/op        12509091 B/op      73037 allocs/op
BenchmarkToGrammar/Ruby_3.2.11-32                    144           8375123 ns/op         4659083 B/op      16458 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/bison    10.530s
```
