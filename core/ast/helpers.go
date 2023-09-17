package ast

import (
	"fireball/core/types"
	"fmt"
	"strings"
)

func (s *Struct) GetField(name string) (int, *Field) {
	for i := range s.Fields {
		field := &s.Fields[i]

		if field.Name.Lexeme == name {
			return i, field
		}
	}

	return 0, nil
}

func (i *Impl) GetMethod(name string) *Func {
	for _, decl := range i.Functions {
		function := decl.(*Func)

		if function.Name.Lexeme == name {
			return function
		}
	}

	return nil
}

func (e *Enum) GetCase(name string) *EnumCase {
	for i := range e.Cases {
		case_ := &e.Cases[i]

		if case_.Name.Lexeme == name {
			return case_
		}
	}

	return nil
}

func (f *Func) Signature(paramNames bool) string {
	signature := strings.Builder{}
	signature.WriteRune('(')

	for i, param := range f.Params {
		if i > 0 {
			signature.WriteString(", ")
		}

		if paramNames {
			signature.WriteString(param.Name.Lexeme)
			signature.WriteRune(' ')
		}

		signature.WriteString(param.Type.String())
	}

	signature.WriteRune(')')

	if !types.IsPrimitive(f.Returns, types.Void) {
		signature.WriteRune(' ')
		signature.WriteString(f.Returns.String())
	}

	return signature.String()
}

func (f *Func) Method() *Struct {
	if impl, ok := f.Parent().(*Impl); ok {
		return impl.Type_
	}

	return nil
}

func (f *Func) MangledName() string {
	name := f.Name.Lexeme

	if f.Extern {
		return name
	}

	if struct_ := f.Method(); struct_ != nil {
		name = fmt.Sprintf("%s.%s", struct_, name)
	}

	return "fb$" + name
}
