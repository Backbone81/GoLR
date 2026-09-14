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

Any not supported functionality is ignored. This includes semantic actions: the generated parser builds a parse tree
instead, so the logic of the actions has to be moved into code which walks that tree.

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
BenchmarkToGrammar/GNU_Bison_3.8.2-32          878    1,541,391 ns/op    1,423,597 B/op    2,872 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32             313    4,021,261 ns/op    1,967,210 B/op    9,325 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32   234    4,866,708 ns/op    2,219,996 B/op   12,083 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32            182    6,496,936 ns/op    4,092,436 B/op   21,614 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32           228    4,916,663 ns/op    3,509,716 B/op   11,442 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                 348    3,115,382 ns/op    1,780,297 B/op    8,271 allocs/op
BenchmarkToGrammar/PHP_8.6.7-32                222    4,761,388 ns/op    2,455,905 B/op   15,639 allocs/op
BenchmarkToGrammar/PostgreSQL_18.4-32           75   14,859,096 ns/op   12,509,091 B/op   73,037 allocs/op
BenchmarkToGrammar/Ruby_3.2.11-32              144    8,375,123 ns/op    4,659,083 B/op   16,458 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/bison    10.530s
```
