package main

import (
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       int
		wantErr    bool
	}{
		// Basic literals
		{name: "integer", expression: "42", want: 42},
		{name: "zero", expression: "0", want: 0},

		// Arithmetic operators
		{name: "addition", expression: "1 + 2", want: 3},
		{name: "subtraction", expression: "5 - 3", want: 2},
		{name: "multiplication", expression: "3 * 4", want: 12},
		{name: "division", expression: "10 / 2", want: 5},

		// Operator precedence: * and / bind tighter than + and -
		{name: "precedence mul over add", expression: "2 + 3 * 4", want: 14},
		{name: "precedence div over sub", expression: "10 - 6 / 2", want: 7},

		// Left associativity
		{name: "left assoc subtraction", expression: "10 - 3 - 2", want: 5},
		{name: "left assoc division", expression: "24 / 4 / 2", want: 3},

		// Unary minus: higher precedence than * and /, so -2 * 3 = (-2) * 3
		{name: "unary minus", expression: "-5", want: -5},
		{name: "unary minus with add", expression: "-2 + 3", want: 1},
		{name: "unary minus with mul", expression: "-2 * 3", want: -6},
		{name: "double unary minus", expression: "--5", want: 5},

		// Parentheses override precedence
		{name: "parens override precedence", expression: "(2 + 3) * 4", want: 20},
		{name: "nested parens", expression: "((3 + 4))", want: 7},
		{name: "parens both sides", expression: "(1 + 2) * (3 + 4)", want: 21},

		// Negative results
		{name: "negative result", expression: "3 - 5", want: -2},

		// Whitespace variations
		{name: "no spaces", expression: "1+2", want: 3},
		{name: "extra spaces", expression: "  1  +  2  ", want: 3},

		// Error cases
		{name: "division by zero", expression: "5 / 0", wantErr: true},
		{name: "invalid input", expression: "abc", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Evaluate(test.expression)
			if test.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", test.expression)
				}
				return
			}
			if err != nil {
				t.Errorf("Evaluate(%q) unexpected error: %v", test.expression, err)
				return
			}
			if got != test.want {
				t.Errorf("Evaluate(%q) = %d, want %d", test.expression, got, test.want)
			}
		})
	}
}
