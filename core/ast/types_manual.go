package ast

import (
	"fmt"
)

// Helpers

func IsPrimitive(type_ Type, kind PrimitiveKind) bool {
	if primitive, ok := As[*Primitive](type_); ok {
		return primitive.Kind == kind
	}

	return false
}

func Resolved(type_ Type) Type {
	for type_ != type_.Resolved() {
		type_ = type_.Resolved()
	}

	return type_
}

func As[T Type](type_ Type) (T, bool) {
	if IsNil(type_) {
		var empty T
		return empty, false
	}

	t, ok := Resolved(type_).(T)
	return t, ok
}

// Primitive

func (p *Primitive) Equals(other Type) bool {
	return IsPrimitive(other, p.Kind)
}

// Pointer

func (p *Pointer) Equals(other Type) bool {
	if p2, ok := As[*Pointer](other); ok {
		return typesEquals(p.Pointee, p2.Pointee)
	}

	return false
}

// Array

func (a *Array) Equals(other Type) bool {
	if a2, ok := As[*Array](other); ok {
		return typesEquals(a.Base, a2.Base) && a.Count == a2.Count
	}

	return false
}

// Resolvable

func (r *Resolvable) Equals(other Type) bool {
	if r.Resolved() == nil {
		panic("ast.Resolvable.Equals() - Not resolved")
	}

	if IsNil(other) {
		return false
	}

	return r.Resolved().Equals(other.Resolved())
}

// Generic

func (g *Generic) Equals(other Type) bool {
	return other != nil && g == other.Resolved()
}

func (g *Generic) Resolved() Type {
	if g.Type != nil {
		return g.Type
	}

	return g
}

// Struct

func (s *Struct) Equals(other Type) bool {
	return other != nil && s == other.Resolved()
}

func (s *Struct) Resolved() Type {
	if s.Type != nil {
		return s.Type
	}

	return s
}

func (s *Struct) AcceptType(visitor TypeVisitor) {
	visitor.VisitStruct(s)
}

// Enum

func (e *Enum) Equals(other Type) bool {
	return e == other.Resolved()
}

func (e *Enum) Resolved() Type {
	return e
}

func (e *Enum) AcceptType(visitor TypeVisitor) {
	visitor.VisitEnum(e)
}

// Interface

func (i *Interface) Equals(other Type) bool {
	return other != nil && i == other.Resolved()
}

func (i *Interface) Resolved() Type {
	return i
}

func (i *Interface) AcceptType(visitor TypeVisitor) {
	visitor.VisitInterface(i)
}

// Func

func (f *Func) Equals(other Type) bool {
	if f2, ok := As[FuncType](other); ok {
		if f.Name != nil && f2.Underlying().Name != nil {
			return f == f2
		}

		return typesEquals(f.Returns(), f2.Returns()) && paramsEquals(f, f2)
	}

	return false
}

func (f *Func) NameAndSignatureEquals(other FuncType) bool {
	return tokensEquals(f.Name, other.Underlying().Name) && typesEquals(f.Returns(), other.Returns()) && paramsEquals(f, other)
}

func paramsEquals(f1, f2 FuncType) bool {
	if f1.ParameterCount() != f2.ParameterCount() {
		return false
	}

	for i := 0; i < f1.ParameterCount(); i++ {
		if !f1.ParameterIndex(i).Type.Equals(f2.ParameterIndex(i).Type) {
			return false
		}
	}

	return true
}

func (f *Func) Resolved() Type {
	return f
}

func (f *Func) AcceptType(visitor TypeVisitor) {
	visitor.VisitFunc(f)
}

// Printer

type TypePrintOptions struct {
	FuncNames  bool
	ParamNames bool
}

type typePrinter struct {
	options TypePrintOptions
	str     string
}

func PrintTypeOptions(type_ Type, options TypePrintOptions) string {
	t := typePrinter{options: options}
	t.VisitNode(type_)

	return t.str
}

func PrintType(type_ Type) string {
	return PrintTypeOptions(type_, TypePrintOptions{})
}

func (t *typePrinter) VisitPrimitive(type_ *Primitive) {
	t.str += type_.Kind.String()
}

func (t *typePrinter) VisitPointer(type_ *Pointer) {
	t.str += "*"
	type_.AcceptChildren(t)
}

func (t *typePrinter) VisitArray(type_ *Array) {
	t.str += fmt.Sprintf("[%d]", type_.Count)
	type_.AcceptChildren(t)
}

func (t *typePrinter) VisitResolvable(type_ *Resolvable) {
	for i, part := range type_.Parts {
		if i > 0 {
			t.str += "."
		}

		t.str += part.String()
	}
}

func (t *typePrinter) VisitGeneric(type_ *Generic) {
	t.str += type_.String()
}

func (t *typePrinter) VisitStruct(_ *Struct) {
	panic("ast.typePrinter.VisitStruct() - Not implemented")
}

func (t *typePrinter) VisitEnum(type_ *Enum) {
	if type_.Name != nil {
		t.str += type_.Name.String()
	}
}

func (t *typePrinter) VisitInterface(type_ *Interface) {
	if type_.Name != nil {
		t.str += type_.Name.String()
	}
}

func (t *typePrinter) VisitFunc(_ *Func) {
	panic("ast.typePrinter.VisitFunc() - Not implemented")
}

func (t *typePrinter) VisitNode(node Node) {
	switch type_ := node.(type) {
	case StructType:
		if type_.Underlying().Name != nil {
			t.str += type_.Underlying().Name.String()
		}

		if spec, ok := type_.(*SpecializedStruct); ok {
			t.str += "!["

			for i, type_ := range spec.Types {
				if i > 0 {
					t.str += ", "
				}

				t.VisitNode(Resolved(type_))
			}

			t.str += "]"
		} else if s, ok := type_.(*Struct); ok && len(s.GenericParams) != 0 {
			t.str += "!["

			for i, param := range s.GenericParams {
				if i > 0 {
					t.str += ", "
				}

				t.str += param.Name.String()
			}

			t.str += "]"
		}

	case FuncType:
		if t.options.FuncNames && type_.Underlying().Name != nil {
			t.str += type_.Underlying().Name.String()
		}

		t.str += Signature(type_, t.options.ParamNames)

	case Type:
		type_.AcceptType(t)
	}
}

// Utils

func tokensEquals(t1, t2 *Token) bool {
	if t1 == nil || t2 == nil {
		return t1 == nil && t2 == nil
	}

	return t1.String() == t2.String()
}

func typesEquals(t1, t2 Type) bool {
	return (t1 == nil && t2 == nil) || (t1 != nil && t2 != nil && t1.Equals(t2))
}
