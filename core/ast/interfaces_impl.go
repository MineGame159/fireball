package ast

import "strings"

// Field

func (f *Field) Underlying() *Field {
	return f
}

func (f *Field) Struct() StructType {
	return f.Parent().(StructType)
}

func (f *Field) Name() *Token {
	return f.Name_
}

func (f *Field) Type() Type {
	return f.Type_
}

// Struct

func (s *Struct) Underlying() *Struct {
	return s
}

func (s *Struct) StaticFieldCount() int {
	return len(s.StaticFields)
}

func (s *Struct) StaticFieldIndex(index int) FieldLike {
	return s.StaticFields[index]
}

func (s *Struct) StaticFieldName(name string) FieldLike {
	for _, field := range s.StaticFields {
		if field.Name() != nil && field.Name().String() == name {
			return field
		}
	}

	return nil
}

func (s *Struct) FieldCount() int {
	return len(s.Fields)
}

func (s *Struct) FieldIndex(index int) FieldLike {
	return s.Fields[index]
}

func (s *Struct) FieldName(name string) FieldLike {
	for _, field := range s.Fields {
		if field.Name() != nil && field.Name().String() == name {
			return field
		}
	}

	return nil
}

func (s *Struct) MangledName(name *strings.Builder) {
	file := GetParent[*File](s)
	file.Namespace.Name.WriteTo(name)

	name.WriteRune('.')
	name.WriteString(s.Name.String())
}

// Func

func (f *Func) Underlying() *Func {
	return f
}

func (f *Func) Receiver() Type {
	if impl, ok := f.Parent().(*Impl); ok && !f.IsStatic() {
		return impl.Type.(*Struct)
	}

	if inter, ok := f.Parent().(*Interface); ok {
		return inter
	}

	return nil
}

func (f *Func) ParameterCount() int {
	return len(f.Params)
}

func (f *Func) ParameterIndex(index int) SpecializedParam {
	return SpecializedParam{
		Param: f.Params[index],
		Type:  f.Params[index].Type,
	}
}

func (f *Func) Returns() Type {
	return f.Returns_
}

func (f *Func) MangledName(name *strings.Builder) {
	funcMangledName(f, f.Struct(), name)
}

func funcMangledName(f *Func, receiver StructType, name *strings.Builder) {
	// Extern
	if extern := f.ExternName(); extern != "" {
		name.WriteString(extern)
		return
	}

	if receiver != nil {
		// Receiver
		receiver.MangledName(name)

		if f.IsStatic() {
			name.WriteString(":f:")
		} else {
			name.WriteString(":m:")
		}
	} else {
		// Namespace
		file := GetParent[*File](f)

		file.Namespace.Name.WriteTo(name)
		name.WriteRune('.')
	}

	// Name
	name.WriteString(f.Name.String())
}
