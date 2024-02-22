package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
	"slices"
	"strings"
)

// Struct

func (s *Struct) Specialize(types []Type) StructType {
	if len(types) != len(s.GenericParams) {
		return s
	}

	// Check cache
	for _, spec := range s.Specializations {
		if slices.EqualFunc(spec.Types, types, typesEquals) {
			return spec
		}
	}

	// Create
	staticFields := specializeFields(s, types, s.StaticFields)
	fields := specializeFields(s, types, s.Fields)

	spec := &SpecializedStruct{
		wrapper:      wrapper[*Struct]{wrapped: s},
		Types:        types,
		staticFields: staticFields,
		fields:       fields,
	}

	for i := 0; i < len(staticFields); i++ {
		staticFields[i].struct_ = spec
	}

	for i := 0; i < len(fields); i++ {
		fields[i].struct_ = spec
	}

	s.Specializations = append(s.Specializations, spec)
	return spec
}

func specializeFields(s *Struct, types []Type, fields []*Field) []SpecializedField {
	specFields := make([]SpecializedField, 0, len(fields))

	for _, field := range fields {
		if IsNil(field.Type()) {
			continue
		}

		type_ := specialize(s.GenericParams, types, field.Type())

		if type_ == nil {
			type_ = field.Type()
		}

		specFields = append(specFields, SpecializedField{
			wrapper: wrapper[*Field]{wrapped: field},
			type_:   type_,
		})
	}

	return specFields
}

// Func

func (f *Func) Generics() []*Generic {
	return f.GenericParams
}

func (f *Func) Specialize(types []Type) FuncType {
	if len(types) != len(f.GenericParams) || len(types) == 0 {
		return f
	}

	spec, _ := specializeFunc(f.Receiver(), &f.Specializations, f, types)
	return spec
}

func specializeFunc(receiver Type, specializations *[]*SpecializedFunc, f SpecializableFunc, types []Type) (FuncType, bool) {
	// Check cache
	for _, spec := range *specializations {
		if slices.EqualFunc(spec.Types, types, typesEquals) {
			return spec, false
		}
	}

	// Create
	params := make([]SpecializedParam, f.ParameterCount())
	returns := f.Returns()

	for i := 0; i < f.ParameterCount(); i++ {
		param := f.ParameterIndex(i)
		type_ := specialize(f.Generics(), types, param.Type)

		if type_ == nil {
			type_ = param.Type
		}

		params[i] = SpecializedParam{
			Param: param.Param,
			Type:  type_,
		}
	}

	if spec := specialize(f.Generics(), types, returns); spec != nil {
		returns = spec
	}

	spec := &SpecializedFunc{
		wrapper:  wrapper[*Func]{wrapped: f.Underlying()},
		receiver: receiver,
		Types:    types,
		params:   params,
		returns:  returns,
	}

	*specializations = append(*specializations, spec)
	return spec, true
}

func specialize(generics []*Generic, types []Type, type_ Type) Type {
	switch type_ := type_.Resolved().(type) {
	case *Pointer:
		pointee := specialize(generics, types, type_.Pointee)

		if pointee != nil {
			return &Pointer{
				cst:     type_.cst,
				Pointee: pointee,
			}
		}

		return nil

	case *Resolvable:
		return specialize(generics, types, type_.Type)

	case *Generic:
		i := slices.Index(generics, type_)
		if i != -1 {
			return types[i]
		}

		return nil

	case StructType:
		var copied *SpecializedStruct

		for i := 0; i < type_.StaticFieldCount(); i++ {
			field := type_.StaticFieldIndex(i)
			spec := specialize(generics, types, field.Type())

			if spec != nil {
				if copied == nil {
					copied = shallowCopyStruct(type_, types)
				}

				copied.staticFields[i].type_ = spec
			}
		}

		for i := 0; i < type_.FieldCount(); i++ {
			field := type_.FieldIndex(i)
			spec := specialize(generics, types, field.Type())

			if spec != nil {
				if copied == nil {
					copied = shallowCopyStruct(type_, types)
				}

				copied.fields[i].type_ = spec
			}
		}

		if copied == nil {
			return nil
		}

		return copied

	case *Func:
		var copied *SpecializedFunc

		for i, param := range type_.Params {
			spec := specialize(generics, types, param.Type)

			if spec != nil {
				if copied == nil {
					copied = shallowCopyFuncType(type_, types)
				}

				copied.params[i].Type = spec
			}
		}

		spec := specialize(generics, types, type_.Returns())

		if spec != nil {
			if copied == nil {
				copied = shallowCopyFuncType(type_, types)
			}

			copied.returns = spec
		}

		return copied

	default:
		return nil
	}
}

func shallowCopyStruct(s StructType, types []Type) *SpecializedStruct {
	// Cache
	for _, spec := range s.Underlying().Specializations {
		if slices.EqualFunc(spec.Types, types, typesEquals) {
			return spec
		}
	}

	// Create
	staticFields := make([]SpecializedField, s.StaticFieldCount())
	for i := 0; i < s.StaticFieldCount(); i++ {
		field := s.StaticFieldIndex(i)

		staticFields[i] = SpecializedField{
			wrapper: wrapper[*Field]{wrapped: field.Underlying()},
			type_:   field.Type(),
		}
	}

	fields := make([]SpecializedField, s.FieldCount())
	for i := 0; i < s.FieldCount(); i++ {
		field := s.FieldIndex(i)

		fields[i] = SpecializedField{
			wrapper: wrapper[*Field]{wrapped: field.Underlying()},
			type_:   field.Type(),
		}
	}

	spec := &SpecializedStruct{
		wrapper:      wrapper[*Struct]{wrapped: s.Underlying()},
		Types:        types,
		Methods:      nil,
		staticFields: staticFields,
		fields:       fields,
	}

	for i := 0; i < len(staticFields); i++ {
		staticFields[i].struct_ = spec
	}

	for i := 0; i < len(fields); i++ {
		fields[i].struct_ = spec
	}

	s.Underlying().Specializations = append(s.Underlying().Specializations, spec)
	return spec
}

func shallowCopyFuncType(f *Func, types []Type) *SpecializedFunc {
	// Cache
	for _, spec := range f.Specializations {
		if slices.EqualFunc(spec.Types, types, typesEquals) {
			return spec
		}
	}

	// Create
	params := make([]SpecializedParam, len(f.Params))
	for i := 0; i < f.ParameterCount(); i++ {
		param := f.ParameterIndex(i)

		params[i] = SpecializedParam{
			Param: param.Param,
			Type:  param.Type,
		}
	}

	spec := &SpecializedFunc{
		wrapper:  wrapper[*Func]{wrapped: f},
		receiver: f.Receiver(),
		Types:    types,
		params:   params,
		returns:  f.Returns_,
	}

	f.Specializations = append(f.Specializations, spec)
	return spec
}

// Specialized Field

type SpecializedField struct {
	wrapper[*Field]

	struct_ StructType
	type_   Type
}

func (s *SpecializedField) Clone() Node {
	return &SpecializedField{
		wrapper: s.wrapper,
		struct_: s.struct_,
		type_:   s.type_,
	}
}

func (s *SpecializedField) Underlying() *Field {
	return s.wrapped
}

func (s *SpecializedField) Struct() StructType {
	return s.struct_
}

func (s *SpecializedField) Name() *Token {
	return s.Underlying().Name_
}

func (s *SpecializedField) Type() Type {
	return s.type_
}

// Specialized Struct

type SpecializedStruct struct {
	wrapper[*Struct]

	Types   []Type
	Methods []*SpecializedFunc

	staticFields []SpecializedField
	fields       []SpecializedField

	Type Type
}

func (s *SpecializedStruct) Clone() Node {
	return &SpecializedStruct{
		wrapper:      s.wrapper,
		Types:        s.Types,
		staticFields: s.staticFields,
		fields:       s.fields,
	}
}

func (s *SpecializedStruct) Equals(other Type) bool {
	return other != nil && s == other.Resolved()
}

func (s *SpecializedStruct) Resolved() Type {
	if s.Type != nil {
		return s.Type
	}

	return s
}

func (s *SpecializedStruct) AcceptType(visitor TypeVisitor) {
	s.Underlying().AcceptType(visitor)
}

func (s *SpecializedStruct) Underlying() *Struct {
	return s.wrapped
}

func (s *SpecializedStruct) StaticFieldCount() int {
	return len(s.staticFields)
}

func (s *SpecializedStruct) StaticFieldIndex(index int) FieldLike {
	return &s.staticFields[index]
}

func (s *SpecializedStruct) StaticFieldName(name string) FieldLike {
	for i, field := range s.staticFields {
		if field.Name() != nil && field.Name().String() == name {
			return &s.staticFields[i]
		}
	}

	return nil
}

func (s *SpecializedStruct) FieldCount() int {
	return len(s.fields)
}

func (s *SpecializedStruct) FieldIndex(index int) FieldLike {
	return &s.fields[index]
}

func (s *SpecializedStruct) FieldName(name string) FieldLike {
	for i, field := range s.fields {
		if field.Name() != nil && field.Name().String() == name {
			return &s.fields[i]
		}
	}

	return nil
}

func (s *SpecializedStruct) SpecializeMethod(f *Func) FuncType {
	// Cache
	if len(f.Generics()) == 0 {
		for _, method := range s.Methods {
			if method.Underlying().Equals(f) {
				return method
			}
		}
	}

	// Create
	params := make([]SpecializedParam, f.ParameterCount())
	returns := f.Returns()

	for i := 0; i < f.ParameterCount(); i++ {
		param := f.ParameterIndex(i)
		type_ := specialize(s.Underlying().GenericParams, s.Types, param.Type)

		if type_ == nil {
			type_ = param.Type
		}

		params[i] = SpecializedParam{
			Param: param.Param,
			Type:  type_,
		}
	}

	if spec := specialize(s.Underlying().GenericParams, s.Types, returns); spec != nil {
		returns = spec
	}

	if len(f.Generics()) > 0 {
		return &PartiallySpecializedFunc{
			wrapper:  wrapper[*Func]{wrapped: f.Underlying()},
			receiver: s,
			params:   params,
			returns:  returns,
		}
	}

	spec := &SpecializedFunc{
		wrapper:  wrapper[*Func]{wrapped: f.Underlying()},
		receiver: s,
		Types:    nil,
		params:   params,
		returns:  returns,
	}

	s.Methods = append(s.Methods, spec)
	return spec
}

func (s *SpecializedStruct) MangledName(name *strings.Builder) {
	s.Underlying().MangledName(name)
	name.WriteString("![")

	for i, type_ := range s.Types {
		if i > 0 {
			name.WriteRune(',')
		}

		name.WriteString(PrintType(type_.Resolved()))
	}

	name.WriteRune(']')
}

// Partially Specialized Func

type PartiallySpecializedFunc struct {
	wrapper[*Func]

	receiver *SpecializedStruct

	params  []SpecializedParam
	returns Type
}

func (p *PartiallySpecializedFunc) Clone() Node {
	return &PartiallySpecializedFunc{
		wrapper:  p.wrapper,
		receiver: p.receiver,
		params:   p.params,
		returns:  p.returns,
	}
}

func (p *PartiallySpecializedFunc) Equals(other Type) bool {
	if f2, ok := As[FuncType](other); ok {
		if p.Underlying().Name != nil && f2.Underlying().Name != nil {
			return p == f2
		}

		return typesEquals(p.Returns(), f2.Returns()) && paramsEquals(p, f2)
	}

	return false
}

func (p *PartiallySpecializedFunc) Resolved() Type {
	return p
}

func (p *PartiallySpecializedFunc) AcceptType(visitor TypeVisitor) {
	p.Underlying().AcceptType(visitor)
}

func (p *PartiallySpecializedFunc) Underlying() *Func {
	return p.wrapped
}

func (p *PartiallySpecializedFunc) Receiver() Type {
	return p.receiver
}

func (p *PartiallySpecializedFunc) ParameterCount() int {
	return len(p.params)
}

func (p *PartiallySpecializedFunc) ParameterIndex(index int) SpecializedParam {
	return p.params[index]
}

func (p *PartiallySpecializedFunc) Returns() Type {
	return p.returns
}

func (p *PartiallySpecializedFunc) MangledName(name *strings.Builder) {
	panic("ast.PartiallySpecializedFunc.MangledName() - Not fully specialized")
}

func (p *PartiallySpecializedFunc) Generics() []*Generic {
	return p.Underlying().Generics()
}

func (p *PartiallySpecializedFunc) Specialize(types []Type) FuncType {
	if len(types) != len(p.Generics()) || len(types) == 0 {
		return p
	}

	spec, new_ := specializeFunc(p.receiver, &p.receiver.Methods, p, types)

	if new_ {
		p.Underlying().Specializations = append(p.Underlying().Specializations, spec.(*SpecializedFunc))
	}

	return spec
}

// Specialized Func

type SpecializedFunc struct {
	wrapper[*Func]

	receiver Type
	Types    []Type

	params  []SpecializedParam
	returns Type
}

func (s *SpecializedFunc) Clone() Node {
	return &SpecializedFunc{
		wrapper:  s.wrapper,
		receiver: s.receiver,
		Types:    s.Types,
		params:   s.params,
		returns:  s.returns,
	}
}

func (s *SpecializedFunc) Equals(other Type) bool {
	if f2, ok := As[FuncType](other); ok {
		if s.Underlying().Name != nil && f2.Underlying().Name != nil {
			return s == f2
		}

		return typesEquals(s.Returns(), f2.Returns()) && paramsEquals(s, f2)
	}

	return false
}

func (s *SpecializedFunc) Resolved() Type {
	return s
}

func (s *SpecializedFunc) AcceptType(visitor TypeVisitor) {
	s.Underlying().AcceptType(visitor)
}

func (s *SpecializedFunc) Underlying() *Func {
	return s.wrapped
}

func (s *SpecializedFunc) Receiver() Type {
	return s.receiver
}

func (s *SpecializedFunc) ParameterCount() int {
	return len(s.params)
}

func (s *SpecializedFunc) ParameterIndex(index int) SpecializedParam {
	return s.params[index]
}

func (s *SpecializedFunc) Returns() Type {
	return s.returns
}

func (s *SpecializedFunc) MangledName(name *strings.Builder) {
	// Base
	var receiver StructType

	if s.Receiver() != nil {
		if s, ok := s.Receiver().(StructType); ok {
			receiver = s
		} else {
			panic("ast.SpecializedFunc.MangledName() - Receiver is not a StructType, this shouldn't happen")
		}
	}

	if receiver == nil {
		receiver = s.Underlying().Struct()
	}

	funcMangledName(s.Underlying(), receiver, name)

	// Generic args
	if len(s.Types) > 0 {
		name.WriteString("![")

		for i, type_ := range s.Types {
			if i > 0 {
				name.WriteRune(',')
			}

			name.WriteString(PrintType(type_))
		}

		name.WriteRune(']')
	}
}

// wrapper

type wrapper[T Node] struct {
	wrapped T
}

func (w *wrapper[T]) Cst() *cst.Node {
	return w.wrapped.Cst()
}

func (w *wrapper[T]) Token() scanner.Token {
	return w.wrapped.Token()
}

func (w *wrapper[T]) Parent() Node {
	return w.wrapped.Parent()
}

func (w *wrapper[T]) SetParent(parent Node) {
	panic("ast.wrapper.SetParent() - Not supported")
}

func (w *wrapper[T]) AcceptChildren(visitor Visitor) {
	w.wrapped.AcceptChildren(visitor)
}

func (w *wrapper[T]) String() string {
	return w.wrapped.String()
}
