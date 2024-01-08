package ast

type ExprResultKind = uint8

const (
	InvalidResultKind ExprResultKind = iota
	TypeResultKind
	ValueResultKind
	CallableResultKind
)

type ExprResultFlags = uint8

const (
	AssignableFlag ExprResultFlags = 1 << iota
	AddressableFlag
)

type ExprResult struct {
	Kind  ExprResultKind
	Flags ExprResultFlags

	Type Type
	data any
}

// Flags

func (e *ExprResult) IsAssignable() bool {
	return e.Flags&AssignableFlag != 0
}

func (e *ExprResult) IsAddressable() bool {
	return e.Flags&AddressableFlag != 0
}

// Invalid

func (e *ExprResult) SetInvalid() {
	*e = ExprResult{}
}

// Type

func (e *ExprResult) SetType(type_ Type) {
	*e = ExprResult{
		Kind: TypeResultKind,
		Type: type_.Resolved(),
	}
}

// Value

func (e *ExprResult) SetValue(type_ Type, flags ExprResultFlags, node Node) {
	*e = ExprResult{
		Kind:  ValueResultKind,
		Flags: flags,
		Type:  type_.Resolved(),
		data:  node,
	}
}

func (e *ExprResult) Value() Node {
	if e.Kind != ValueResultKind {
		panic("expr.ExprResult.Value() - Result is not a value")
	}

	if e.data == nil {
		return nil
	}

	return e.data.(Node)
}

// Callable

func (e *ExprResult) SetCallable(type_ Type, node Node) {
	*e = ExprResult{
		Kind: CallableResultKind,
		Type: type_,
		data: node,
	}
}

func (e *ExprResult) Callable() Node {
	if e.Kind != CallableResultKind {
		panic("expr.ExprResult().Callable() - Result is not a callable")
	}

	if e.data == nil {
		return nil
	}

	return e.data.(Node)
}
