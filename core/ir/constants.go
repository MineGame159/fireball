package ir

type Const interface {
	Value

	isConst()
}

// Null

var Null = &NullConst{}
var nullType = &PointerType{Pointee: Void}

type NullConst struct{}

func (n *NullConst) Type() Type {
	return nullType
}

func (n *NullConst) Name() string {
	return ""
}

func (n *NullConst) isConst() {}

// Int

var True = &IntConst{Typ: I1, Value: Unsigned(1)}
var False = &IntConst{Typ: I1, Value: Unsigned(0)}

type IntConst struct {
	Typ   Type
	Value Int
}

func (i *IntConst) Type() Type {
	return i.Typ
}

func (i *IntConst) Name() string {
	return ""
}

func (i *IntConst) isConst() {}

// Float

type FloatConst struct {
	Typ   Type
	Value float64
}

func (f *FloatConst) Type() Type {
	return f.Typ
}

func (f *FloatConst) Name() string {
	return ""
}

func (f *FloatConst) isConst() {}

// String

type StringConst struct {
	typ ArrayType

	Length uint32
	Value  []byte
}

func (s *StringConst) Type() Type {
	if s.typ.Count == 0 {
		s.typ = ArrayType{
			Count: s.Length,
			Base:  I8,
		}
	}

	return &s.typ
}

func (s *StringConst) Name() string {
	return ""
}

func (s *StringConst) isConst() {}

// Zero Initializer

type ZeroInitConst struct {
	Typ Type
}

func (z *ZeroInitConst) Type() Type {
	return z.Typ
}

func (z *ZeroInitConst) Name() string {
	return ""
}

func (z *ZeroInitConst) isConst() {}

// Array

type ArrayConst struct {
	Typ    Type
	Values []Value
}

func (a *ArrayConst) Type() Type {
	return a.Typ
}

func (a *ArrayConst) Name() string {
	return ""
}

func (a *ArrayConst) isConst() {}

// Struct

type StructConst struct {
	Typ    Type
	Fields []Value
}

func (s *StructConst) Type() Type {
	return s.Typ
}

func (s *StructConst) Name() string {
	return ""
}

func (s *StructConst) isConst() {}
