package ast

// Func flags

type FuncFlags uint8

const (
	Static FuncFlags = 1 << iota
	Variadic
)
