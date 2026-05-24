# Motivation

## Problem Statement

While LR parsing is the most general deterministic shift-reduce parsing method known, its usage for real world
applications was significantly restricted in the past. As canonical LR parsers were seen as too resource intense
with regard to time and space, inferior algorithms like LALR were used which made it difficult for language designers
to design and iterate on their grammars in a productive manner.

With the introduction of new algorithms like the Inadequacy Elimination LR(1) algorithm as described in the paper
"[The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution](https://doi.org/10.1016/j.scico.2009.08.001)"
by Joel E. Denny and Brian A. Malloy in 2009, the full power of LR parsers are now available with time and space
requirements similar to the likes of LALR. This makes LR parsers interesting for real world applications and for
interactive iteration on grammars.

Unfortunately, these new algorithms have not seen widespread adoption and implementations in parser generator tools.
Reasons are probably the complexity and effort involved for implementing them as well as their use case focusing on a
quite small community of people working on compilers, parsers and domain specific languages.

In addition, those parser generators which exist and are maintained with an implementation of those modern algorithms
are usually difficult to adapt and extend to custom use cases. The input format for a context free grammar is tightly
coupled with the algorithm itself as well as the output format and the code generation. This means that when you have
a different format for describing a context free grammar or if you require a parser for a programming language which
is not natively supported by the parser generator, you are usually out of luck and are better off rolling your own
parser implementation.

This is not a good use of time and effort of everybody involved - be it for those involved in building parser generators
as well as for those in need of parsers for their specific use case.

## Solution

GoLR tries to solve this issue by providing a parser generator which provides a modular and extensible design right from
the start. It strictly separates the frontend from the core and the backend. The frontend is responsible for loading
a context free grammar from any arbitrary format. The core is responsible for producing a parser for a given context
free grammar. The backend is responsible for outputting the parser to any arbitrary format.

The frontends could support different formats like JSON, YAML and GNU Bison grammar files with a
varying degree of compatibility. Making it easy for users to reuse grammar descriptions they already have.

The core could support different algorithms but should start with a IELR(1) implementation.

The backends could support different formats like JSON, YAML, GNU Bison report files, DOT files, C code, C++ code,
Java code, Go code or Rust code. Making it easy to target whatever use case you have in mind.

By supporting formats like JSON and YAML for frontends and backends, this opens up the possibility for others to develop
their own frontends and backends with their own programming language in their own applications by writing or reading
those formats and piping them to or from GoLR. Making modularity and extensibility independent of the maintainers
of GoLR.

This makes advanced algorithms like IELR(1) reusable by a broad range of use cases without the need to extend the parser
generator itself or even to use the programming language the parser generator was developed in. Ultimately, this
approach promotes a broader adoption and application of LR parsers across various domains.
