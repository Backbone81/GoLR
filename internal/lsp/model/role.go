package model

// Role is the part an occurrence plays.
type Role int

const (
	// RoleDefinition is the name of a scanner or parser rule.
	RoleDefinition Role = iota

	// RoleAlias is the string of a terminal declared with a string.
	RoleAlias

	// RoleReference is a use of a symbol by name or by string.
	RoleReference
)
