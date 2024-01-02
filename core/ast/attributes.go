package ast

import "reflect"

// Attributes

type ExternAttribute struct {
	Name string
}

type IntrinsicAttribute struct {
	Name string
}

type InlineAttribute struct {
}

func GetAttribute(attributes []any, attribute any) bool {
	value := reflect.ValueOf(attribute)
	elem := value.Elem()
	type_ := elem.Type()

	for _, attr := range attributes {
		if reflect.TypeOf(attr) == type_ {
			elem.Set(reflect.ValueOf(attr))
			return true
		}
	}

	return false
}

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
