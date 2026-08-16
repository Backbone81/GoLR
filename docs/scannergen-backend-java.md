# Scanner Generator Backend: Java

This backend outputs a scanner as Java source code. The generated Java code is a table driven scanner which holds the
automaton in lookup tables.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.

The generated code targets Java 17. Use `--backend-java-package-name` to set the package, which defaults to `parser`.
