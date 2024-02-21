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

	if f.Name() != nil {
		sb.WriteString(f.Name().String())
	}

	return sb.String()
}

// Interface

func (i *Interface) GetMethod(name string) (*Func, int) {
	for i, function := range i.Methods {
		if function.Name != nil && function.Name.String() == name {
			return function, i
		}
	}

	return nil, 0
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

func (f *Func) TestName() string {
	for _, attribute := range f.Attributes {
		if attribute.Name.String() == "Test" {
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

	if _, ok := f.Parent().(*Interface); ok {
		return false
	}

	return true
}

func (f *Func) Struct() StructType {
	if impl, ok := f.Parent().(*Impl); ok {
		if s, ok := impl.Type.(StructType); ok {
			return s
		}
	}

	return nil
}

func Signature(f FuncType, paramNames bool) string {
	signature := strings.Builder{}
	signature.WriteRune('(')

	for i := 0; i < f.ParameterCount(); i++ {
		param := f.ParameterIndex(i)

		if i > 0 {
			signature.WriteString(", ")
		}

		if paramNames {
			signature.WriteString(param.Param.Name.String())
			signature.WriteRune(' ')
		}

		signature.WriteString(PrintType(param.Type))
	}

	if f.Underlying().IsVariadic() {
		if f.ParameterCount() > 0 {
			signature.WriteString(", ...")
		} else {
			signature.WriteString("...")
		}
	}

	signature.WriteRune(')')

	if !IsPrimitive(f.Returns(), Void) {
		signature.WriteRune(' ')
		signature.WriteString(PrintType(f.Returns()))
	}

	return signature.String()
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
