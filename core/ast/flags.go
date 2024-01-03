package ast

// Func flags

type FuncFlags uint8

const (
	Static FuncFlags = 1 << iota
	Variadic
)

// Identifier kind

type IdentifierKind uint8

const (
	FunctionKind IdentifierKind = iota
	StructKind
	EnumKind
	VariableKind
	ParameterKind
)
