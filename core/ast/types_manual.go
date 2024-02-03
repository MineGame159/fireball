package ast

import (
	"fmt"
	"slices"
)

// Helpers

func IsPrimitive(type_ Type, kind PrimitiveKind) bool {
	if primitive, ok := As[*Primitive](type_); ok {
		return primitive.Kind == kind
	}

	return false
}

func As[T Type](type_ Type) (T, bool) {
	if type_ == nil {
		var empty T
		return empty, false
	}

	t, ok := type_.Resolved().(T)
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

	return r.Resolved().Equals(other.Resolved())
}

// Struct

func (s *Struct) Equals(other Type) bool {
	return other != nil && s == other.Resolved()
}

func (s *Struct) Resolved() Type {
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
	if f2, ok := As[*Func](other); ok {
		if f.Name != nil && f2.Name != nil {
			return f == f2
		}

		return typesEquals(f.Returns, f2.Returns) && slices.EqualFunc(f.Params, f2.Params, paramEquals)
	}

	return false
}

func (f *Func) NameAndSignatureEquals(other *Func) bool {
	return tokensEquals(f.Name, other.Name) && typesEquals(f.Returns, other.Returns) && slices.EqualFunc(f.Params, other.Params, paramEquals)
}

func paramEquals(v1, v2 *Param) bool {
	return typesEquals(v1.Type, v2.Type)
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

func (t *typePrinter) VisitStruct(type_ *Struct) {
	if type_.Name != nil {
		t.str += type_.Name.String()
	}
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

func (t *typePrinter) VisitFunc(type_ *Func) {
	if t.options.FuncNames && type_.Name != nil {
		t.str += type_.Name.String()
	}

	t.str += type_.Signature(t.options.ParamNames)
}

func (t *typePrinter) VisitNode(node Node) {
	if type_, ok := node.(Type); ok {
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
