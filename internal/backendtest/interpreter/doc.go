// Package interpreter contains the reference scanner and the reference parser of the backend test harness. They define
// the behavior every generated scanner and every generated parser has to reproduce, in every language.
//
// The interpreters read the tables a table driven backend emits, and never the code a backend generated. Judging the Go
// backend by an interpreter which called into it would prove nothing about the Go backend, and an interpreter which
// shared its scan loop or its parse loop with the code under test would hide a bug in that loop from both.
package interpreter
