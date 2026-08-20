# Scanner Generator Backend: Python

This backend outputs a scanner as Python source code. The generated Python code is a table driven scanner which holds
the automaton in lookup tables. It targets Python 3.10 and imports nothing beyond the standard library.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.
