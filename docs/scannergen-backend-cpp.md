# Scanner Generator Backend: C++

This backend outputs a scanner as C++ source code. The generated C++ code is a table driven scanner which holds the
automaton in lookup tables. It targets C++17, is a self contained header and includes nothing beyond the standard
library.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.

Use `--backend-cpp-namespace` to set the namespace, which defaults to `parser`.
