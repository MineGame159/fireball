package ast

import (
	"fireball/core/architecture"
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

func (p *Primitive) Size() uint32 {
	switch p.Kind {
	case Void:
		return 0
	case Bool, U8, I8:
		return 1
	case U16, I16:
		return 2
	case U32, I32, F32:
		return 4
	case U64, I64, F64:
		return 8

	default:
		panic("ast.Primitive.Size() - Not implemented")
	}
}

func (p *Primitive) Align() uint32 {
	return p.Size()
}

func (p *Primitive) Equals(other Type) bool {
	return IsPrimitive(other, p.Kind)
}

func (p *Primitive) CanAssignTo(other Type) bool {
	return IsPrimitive(other, p.Kind)
}

// Pointer

func (p *Pointer) Size() uint32 {
	return 8
}

func (p *Pointer) Align() uint32 {
	return 8
}

func (p *Pointer) Equals(other Type) bool {
	if p2, ok := As[*Pointer](other); ok {
		return typesEquals(p.Pointee, p2.Pointee)
	}

	return false
}

func (p *Pointer) CanAssignTo(other Type) bool {
	if p2, ok := As[*Pointer](other); ok {
		return IsPrimitive(p2.Pointee, Void) || typesEquals(p.Pointee, p2.Pointee)
	}

	return false
}

// Array

func (a *Array) Size() uint32 {
	return a.Base.Size() * a.Count
}

func (a *Array) Align() uint32 {
	return a.Base.Align()
}

func (a *Array) Equals(other Type) bool {
	if a2, ok := As[*Array](other); ok {
		return typesEquals(a.Base, a2.Base) && a.Count == a2.Count
	}

	return false
}

func (a *Array) CanAssignTo(other Type) bool {
	return a.Equals(other)
}

// Resolvable

func (r *Resolvable) Size() uint32 {
	if r.Resolved() == nil {
		panic("ast.Resolvable.Size() - Not resolved")
	}

	return r.Resolved().Size()
}

func (r *Resolvable) Align() uint32 {
	if r.Resolved() == nil {
		panic("ast.Resolvable.Align() - Not resolved")
	}

	return r.Resolved().Align()
}

func (r *Resolvable) Equals(other Type) bool {
	if r.Resolved() == nil {
		panic("ast.Resolvable.Equals() - Not resolved")
	}

	return r.Resolved().Equals(other.Resolved())
}

func (r *Resolvable) CanAssignTo(other Type) bool {
	if r.Resolved() == nil {
		panic("ast.Resolvable.Equals() - Not resolved")
	}

	return r.Resolved().CanAssignTo(other.Resolved())
}

// Struct

func (s *Struct) Size() uint32 {
	layout := architecture.CLayout{}

	for _, field := range s.Fields {
		layout.Add(field.Type.Size(), field.Type.Align())
	}

	return layout.Size()
}

func (s *Struct) Align() uint32 {
	align := uint32(0)

	for _, field := range s.Fields {
		align = max(align, field.Type.Align())
	}

	return align
}

func (s *Struct) Equals(other Type) bool {
	if s2, ok := As[*Struct](other); ok {
		return tokensEquals(s.Name, s2.Name) && slices.EqualFunc(s.Fields, s2.Fields, fieldEquals)
	}

	return false
}

func fieldEquals(v1, v2 *Field) bool {
	return tokensEquals(v1.Name, v2.Name) && typesEquals(v1.Type, v2.Type)
}

func (s *Struct) CanAssignTo(other Type) bool {
	if s2, ok := As[*Struct](other); ok {
		return slices.EqualFunc(s.Fields, s2.Fields, fieldEquals)
	}

	return false
}

func (s *Struct) Resolved() Type {
	return s
}

func (s *Struct) AcceptType(visitor TypeVisitor) {
	visitor.VisitStruct(s)
}

// Enum

func (e *Enum) Size() uint32 {
	return e.ActualType.Size()
}

func (e *Enum) Align() uint32 {
	return e.ActualType.Align()
}

func (e *Enum) Equals(other Type) bool {
	if e2, ok := As[*Enum](other); ok {
		return typesEquals(e.ActualType, e2.ActualType) && slices.EqualFunc(e.Cases, e2.Cases, enumCaseEquals)
	}

	return false
}

func enumCaseEquals(v1, v2 *EnumCase) bool {
	return tokensEquals(v1.Name, v2.Name) && v1.ActualValue == v2.ActualValue
}

func (e *Enum) CanAssignTo(other Type) bool {
	return e.Equals(other)
}

func (e *Enum) Resolved() Type {
	return e
}

func (e *Enum) AcceptType(visitor TypeVisitor) {
	visitor.VisitEnum(e)
}

// Func

func (f *Func) Size() uint32 {
	return 8
}

func (f *Func) Align() uint32 {
	return 8
}

func (f *Func) Equals(other Type) bool {
	if f2, ok := As[*Func](other); ok {
		return tokensEquals(f.Name, f2.Name) && typesEquals(f.Returns, f2.Returns) && slices.EqualFunc(f.Params, f2.Params, paramEquals)
	}

	return false
}

func paramEquals(v1, v2 *Param) bool {
	return typesEquals(v1.Type, v2.Type)
}

func (f *Func) CanAssignTo(other Type) bool {
	if f2, ok := As[*Func](other); ok {
		return typesEquals(f.Returns, f2.Returns) && slices.EqualFunc(f.Params, f2.Params, paramEquals)
	}

	return false
}

func (f *Func) Resolved() Type {
	return f
}

func (f *Func) AcceptType(visitor TypeVisitor) {
	visitor.VisitFunc(f)
}

// Printer

type TypePrintOptions struct {
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

func (t *typePrinter) VisitFunc(type_ *Func) {
	t.str += type_.Signature(t.options.ParamNames)
}

func (t *typePrinter) VisitNode(node Node) {
	if type_, ok := node.(Type); ok {
		type_.AcceptType(t)
	}
}

// Utils

func tokensEquals(t1, t2 *Token) bool {
	return (t1 == nil && t2 == nil) || (t1 != nil && t2 != nil && t1.String() == t2.String())
}

func typesEquals(t1, t2 Type) bool {
	return (t1 == nil && t2 == nil) || (t1 != nil && t2 != nil && t1.Equals(t2))
}
