# Parser Generator Backend: Python

This backend outputs a parser as Python source code. The generated Python code is a table driven parser which holds the
parsing decisions in lookup tables. It targets Python 3.10 and imports nothing beyond the standard library.

The parser imports the token constants and the scanner protocol from the generated scanner. Use
`--backend-python-scanner-module` to set the module, which defaults to `scanner`. A leading dot makes the import
relative, which is what a scanner and a parser sitting in the same package need.
