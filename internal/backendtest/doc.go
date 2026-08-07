// Package backendtest contains the test harness which proves that the code the language backends emit behaves
// identically to a reference implementation in Go. It tests the emitted code and the driver around it, not the
// construction of the tables, which is covered where those tables are built.
//
// Every runner in every language prints the same canonical text for the same input, so comparing two backends is a
// plain text diff instead of a language specific assertion. That text is defined by this package.
package backendtest
