# Parser Generator Backend: C++

This backend outputs a parser as C++ source code. The generated C++ code is a table driven parser which holds the
parsing decisions in lookup tables. It targets C++17 and is a self contained header.

The parser uses the token type and the scanner interface of the generated scanner. Use
`--backend-cpp-scanner-include` to set the header it includes them from, which defaults to `scanner.hpp`, and
`--backend-cpp-namespace` to set the namespace, which has to be the one the scanner was generated into.
