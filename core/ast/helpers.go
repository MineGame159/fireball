package ast

import (
	"fireball/core/scanner"
	"strings"
)

// NamespaceName

func (n *NamespaceName) WriteTo(sb *strings.Builder) {
	for i, part := range n.Parts {
		if i > 0 {
			sb.WriteRune('.')
		}

		sb.WriteString(part.String())
	}
}

// Struct

func (s *Struct) GetStaticField(name string) (int, *Field) {
	for i := range s.StaticFields {
		field := s.StaticFields[i]

		if field.Name.String() == name {
			return i, field
		}
	}

	return 0, nil
}

func (s *Struct) GetField(name string) (int, *Field) {
	for i := range s.Fields {
		field := s.Fields[i]

		if field.Name.String() == name {
			return i, field
		}
	}

	return 0, nil
}

// Impl

func (i *Impl) GetMethod(name string, static bool) *Func {
	staticValue := FuncFlags(0)
	if static {
		staticValue = 1
	}

	for _, function := range i.Methods {
		if function.Name != nil && function.Name.String() == name && function.Flags&Static == staticValue {
			return function
		}
	}

	return nil
}

// Enum

func (e *Enum) GetCase(name string) *EnumCase {
	for i := range e.Cases {
		case_ := e.Cases[i]

		if case_.Name.String() == name {
			return case_
		}
	}

	return nil
}

// Field

func (f *Field) IsStatic() bool {
	if f.Cst() != nil {
		return f.Cst().Contains(scanner.Static)
	}

	return false
}

func (f *Field) Index() int {
	if f.IsStatic() {
		for i, field := range f.Parent().(*Struct).StaticFields {
			if field == f {
				return i
			}
		}

		panic("ast.Field.Index() - Static field not found")
	}

	for i, field := range f.Parent().(*Struct).Fields {
		if field == f {
			return i
		}
	}

	panic("ast.Field.Index() - Field not found")
}

func (f *Field) MangledName() string {
	sb := strings.Builder{}
	sb.WriteString("fb$")

	file := GetParent[*File](f)
	file.Namespace.Name.WriteTo(&sb)

	if f.IsStatic() {
		sb.WriteString(":s:")
	} else {
		sb.WriteString(":i:")
	}

	sb.WriteString(f.Parent().(*Struct).Name.String())
	sb.WriteRune('.')

	if f.Name != nil {
		sb.WriteString(f.Name.String())
	}

	return sb.String()
}

// Func

func (f *Func) IsStatic() bool {
	return f.Flags&Static != 0
}

func (f *Func) IsVariadic() bool {
	return f.Flags&Variadic != 0
}

func (f *Func) ExternName() string {
	for _, attribute := range f.Attributes {
		if attribute.Name.String() == "Extern" {
			if len(attribute.Args) > 0 {
				return attribute.Args[0].String()[1 : len(attribute.Args[0].String())-1]
			}

			return f.Name.String()
		}
	}

	return ""
}

func (f *Func) IntrinsicName() string {
	for _, attribute := range f.Attributes {
		if attribute.Name.String() == "Intrinsic" {
			if len(attribute.Args) > 0 {
				return attribute.Args[0].String()[1 : len(attribute.Args[0].String())-1]
			}

			return f.Name.String()
		}
	}

	return ""
}

func (f *Func) HasBody() bool {
	for _, attribute := range f.Attributes {
		if attribute.Name.String() == "Extern" || attribute.Name.String() == "Intrinsic" {
			return false
		}
	}

	return true
}

func (f *Func) Signature(paramNames bool) string {
	signature := strings.Builder{}
	signature.WriteRune('(')

	for i, param := range f.Params {
		if i > 0 {
			signature.WriteString(", ")
		}

		if paramNames {
			signature.WriteString(param.Name.String())
			signature.WriteRune(' ')
		}

		signature.WriteString(PrintType(param.Type))
	}

	if f.IsVariadic() {
		if len(f.Params) > 0 {
			signature.WriteString(", ...")
		} else {
			signature.WriteString("...")
		}
	}

	signature.WriteRune(')')

	if !IsPrimitive(f.Returns, Void) {
		signature.WriteRune(' ')
		signature.WriteString(PrintType(f.Returns))
	}

	return signature.String()
}

func (f *Func) Receiver() *Struct {
	if impl, ok := f.Parent().(*Impl); ok {
		if s, ok := As[*Struct](impl.Type); ok {
			return s
		}
	}

	return nil
}

func (f *Func) Method() *Struct {
	if impl, ok := f.Parent().(*Impl); ok && !f.IsStatic() {
		if s, ok := As[*Struct](impl.Type); ok {
			return s
		}
	}

	return nil
}

func (f *Func) MangledName() string {
	// Extern
	externName := f.ExternName()

	if externName != "" {
		return externName
	}

	// Normal
	sb := strings.Builder{}
	sb.WriteString("fb$")

	file := GetParent[*File](f)
	file.Namespace.Name.WriteTo(&sb)

	if _, ok := f.Parent().(*Impl); ok && !f.IsStatic() {
		sb.WriteString(":m:")
	} else {
		sb.WriteString(":f:")
	}

	if struct_, ok := f.Parent().(*Impl); ok {
		sb.WriteString(struct_.Struct.String())
		sb.WriteRune('.')
	}

	if f.Name != nil {
		sb.WriteString(f.Name.String())
	}

	return sb.String()
}

// GlobalVar

func (g *GlobalVar) MangledName() string {
	sb := strings.Builder{}
	sb.WriteString("fb$")

	file := GetParent[*File](g)
	file.Namespace.Name.WriteTo(&sb)

	sb.WriteString(":g:")

	if g.Name != nil {
		sb.WriteString(g.Name.String())
	}

	return sb.String()
}
