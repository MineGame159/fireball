package ast

import (
	"fireball/core/types"
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
