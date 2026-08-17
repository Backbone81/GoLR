# Scanner Generator Backend: Rust

This backend outputs a scanner as Rust source code. The generated Rust code is a table driven scanner which holds the
automaton in lookup tables.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.
