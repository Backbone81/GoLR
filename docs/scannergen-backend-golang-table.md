# Scanner Generator Backend: Golang Table

This backend outputs a scanner as Go source code. The generated Go code is a table driven scanner which holds the
automaton in lookup tables.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.
