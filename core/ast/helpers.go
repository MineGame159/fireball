package ast

import (
	"fmt"
	"strings"
)

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

func (i *Impl) GetMethod(name string, static bool) *Func {
	staticValue := FuncFlags(0)
	if static {
		staticValue = 1
	}

	for _, function := range i.Methods {
		if function.Name.String() == name && function.Flags&Static == staticValue {
			return function
		}
	}

	return nil
}

func (e *Enum) GetCase(name string) *EnumCase {
	for i := range e.Cases {
		case_ := e.Cases[i]

		if case_.Name.String() == name {
			return case_
		}
	}

	return nil
}

func (f *Field) GetMangledName() string {
	return fmt.Sprintf("fb$%s::%s", f.Parent, f.Name)
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

	signature.WriteRune(')')

	if !IsPrimitive(f.Returns, Void) {
		signature.WriteRune(' ')
		signature.WriteString(PrintType(f.Returns))
	}

	return signature.String()
}

func (f *Func) Method() *Struct {
	if impl, ok := f.Parent().(*Impl); ok && !f.IsStatic() {
		return impl.Type.(*Struct)
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
	name := ""

	if f.Name != nil {
		name = f.Name.String()
	}

	if struct_, ok := f.Parent().(*Impl); ok {
		name = fmt.Sprintf("%s.%s", struct_.Struct, name)
	}

	return "fb$" + name
}
