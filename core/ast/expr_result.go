package ast

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

	Type     Type
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

func (e *ExprResult) SetType(type_ Type) {
	e.Kind = TypeResultKind
	e.Flags = 0
	e.Type = type_.Resolved()
}

func (e *ExprResult) SetFunction(function *Func) {
	e.Kind = FunctionResultKind
	e.Flags = 0
	e.Type = function
	e.Function = function
}

func (e *ExprResult) SetValue(type_ Type, flags ExprResultFlags) {
	e.Kind = ValueResultKind
	e.Flags = flags
	e.Type = type_.Resolved()
}
