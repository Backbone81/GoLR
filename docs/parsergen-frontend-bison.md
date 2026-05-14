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
Objective C, GCC C++, GCC Java and Go.
