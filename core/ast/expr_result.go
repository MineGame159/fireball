package ast

import (
	"fireball/core"
	"fireball/core/types"
)

type ExprResultKind = uint8

const (
	InvalidResultKind ExprResultKind = iota
	TypeResultKind
	FunctionResultKind
	ValueResultKind
)

type ExprResultFlags = uint8

const (
	AssignableFlag ExprResultFlags = 1 << iota
	AddressableFlag
)

type ExprResult struct {
	Kind  ExprResultKind
	Flags ExprResultFlags

	Type     types.Type
	Function *Func
}

func (e *ExprResult) IsAssignable() bool {
	return e.Flags&AssignableFlag != 0
}

func (e *ExprResult) IsAddressable() bool {
	return e.Flags&AddressableFlag != 0
}

// Set

func (e *ExprResult) SetInvalid() {
	e.Kind = InvalidResultKind
	e.Flags = 0
}

func (e *ExprResult) SetType(type_ types.Type) {
	e.Kind = TypeResultKind
	e.Flags = 0
	e.Type = type_
}

func (e *ExprResult) SetFunction(function *Func) {
	e.Kind = FunctionResultKind
	e.Flags = 0
	e.Type = function.WithRange(core.Range{})
	e.Function = function
}

func (e *ExprResult) SetValue(type_ types.Type, flags ExprResultFlags) {
	e.Kind = ValueResultKind
	e.Flags = flags
	e.Type = type_.WithRange(core.Range{})
}

func (e *ExprResult) SetValueRaw(type_ types.Type, flags ExprResultFlags) {
	e.Kind = ValueResultKind
	e.Flags = flags
	e.Type = type_
}
