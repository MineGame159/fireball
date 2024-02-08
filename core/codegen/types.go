package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/ir"
	"fmt"
	"strconv"
)

type cachedType struct {
	type_ ast.Type
	typ   ir.Type
}

type cachedTypeMeta struct {
	type_ ast.Type
	id    ir.MetaID
}

type types struct {
	c *codegen

	interfaceType ir.Type
	types         []cachedType

	interfaceMetadata ir.MetaID
	metadata          []cachedTypeMeta
}

// Types

func (t *types) get(type_ ast.Type) ir.Type {
	switch type_ := type_.Resolved().(type) {
	case *ast.Primitive:
		switch type_.Kind {
		case ast.Void:
			return ir.Void
		case ast.Bool:
			return ir.I1

		case ast.U8, ast.I8:
			return ir.I8
		case ast.U16, ast.I16:
			return ir.I16
		case ast.U32, ast.I32:
			return ir.I32
		case ast.U64, ast.I64:
			return ir.I64

		case ast.F32:
			return ir.F32
		case ast.F64:
			return ir.F64

		default:
			panic("codegen.types.get() - ast.Primitive - Not implemented")
		}

	case *ast.Pointer:
		if typ := t.getCachedType(type_); typ != nil {
			return typ
		}

		return t.cacheType(type_, &ir.PointerType{Pointee: t.get(type_.Pointee)})

	case *ast.Array:
		if typ := t.getCachedType(type_); typ != nil {
			return typ
		}

		return t.cacheType(type_, &ir.ArrayType{Count: type_.Count, Base: t.get(type_.Base)})

	case *ast.Struct:
		if typ := t.getCachedType(type_); typ != nil {
			return typ
		}

		astFields, _ := abi.GetStructLayout(type_).Fields(abi.GetTargetAbi(), type_)
		fields := make([]ir.Type, len(astFields))

		for i, field := range astFields {
			fields[i] = t.get(field.Type)
		}

		typ := &ir.StructType{
			Name:   type_.Name.String(),
			Fields: fields,
		}

		t.c.module.Struct(typ)
		return t.cacheType(type_, typ)

	case *ast.Enum:
		return t.get(type_.ActualType)

	case *ast.Interface:
		if t.interfaceType == nil {
			void := ast.Primitive{Kind: ast.Void}
			ptr := ast.Pointer{Pointee: &void}

			typ := &ir.StructType{
				Name: "__interface",
				Fields: []ir.Type{
					t.get(&ptr),
					t.get(&ptr),
				},
			}

			t.c.module.Struct(typ)
			t.interfaceType = typ
		}

		return t.interfaceType

	case *ast.Func:
		if typ := t.getCachedType(type_); typ != nil {
			return typ
		}

		typ := t.createFuncType(type_)
		return t.cacheType(type_, typ)

	default:
		panic("codegen.types.get() - Not implemented")
	}
}

func (t *types) createFuncType(f *ast.Func) *ir.FuncType {
	// Intrinsic
	if f.IntrinsicName() != "" {
		return t.createIntrinsicFuncType(f)
	}

	// Normal
	this := f.Method()

	parameterCount := len(f.Params)
	if this != nil {
		parameterCount++
	}

	params := make([]*ir.Param, 0, parameterCount)

	funcAbi := abi.GetFuncAbi(f)
	returnArgs := funcAbi.Classify(f.Returns, nil)

	if len(returnArgs) == 1 && returnArgs[0].Class == abi.Memory {
		params = append(params, &ir.Param{
			Typ: &ir.PointerType{
				Pointee: ir.Void,
				SRet:    t.get(f.Returns),
			},
			Name_: "__return",
		})

		returnArgs = nil
	}

	if this != nil {
		type_ := ast.Pointer{Pointee: this}

		params = append(params, &ir.Param{
			Typ:   t.get(&type_),
			Name_: "this",
		})
	}

	for _, param := range f.Params {
		args := funcAbi.Classify(param.Type, nil)
		if len(args) == 0 {
			panic("codegen.types.createFuncType() - Failed to classify parameter type")
		}

		for i, arg := range args {
			typ := t.getAbiArgType(arg, param.Type)

			if typ != nil {
				name := param.Name.String()

				if len(args) > 1 {
					name += "." + strconv.Itoa(i)
				}

				params = append(params, &ir.Param{
					Typ:   typ,
					Name_: name,
				})
			}
		}
	}

	return &ir.FuncType{
		Returns:  t.createReturnType(returnArgs, f.Returns),
		Params:   params,
		Variadic: f.IsVariadic(),
	}
}

func (t *types) createReturnType(args []abi.Arg, type_ ast.Type) ir.Type {
	// Void
	if len(args) == 0 {
		return ir.Void
	}

	// Struct
	if typeIsAbiStruct(type_) {
		return getAbiStructType(args, true)
	}

	// Single
	return t.getAbiArgType(args[0], type_)
}

func (t *types) getAbiArgType(arg abi.Arg, type_ ast.Type) ir.Type {
	switch arg.Class {
	case abi.Integer:
		if _, ok := ast.As[*ast.Pointer](type_); ok {
			void := ast.Primitive{Kind: ast.Void}
			ptr := ast.Pointer{Pointee: &void}

			return t.get(&ptr)
		} else {
			return getAbiIntIrType(arg)
		}

	case abi.SSE:
		return getAbiSseIrType(arg)

	case abi.Memory:
		return &ir.PointerType{
			Pointee: ir.Void,
			ByVal:   t.get(type_),
		}

	default:
		panic("codegen.types.getAbiArgType() - ABI class not implemented")
	}
}

func (t *types) createIntrinsicFuncType(f *ast.Func) *ir.FuncType {
	intrinsicName := f.IntrinsicName()

	intrinsic := t.getIntrinsicTypes(f, intrinsicName)

	params := make([]*ir.Param, len(intrinsic[1:]))

	for i, param := range intrinsic[1:] {
		params[i] = &ir.Param{
			Typ:   param,
			Name_: fmt.Sprintf("_%d", i),
		}
	}

	return &ir.FuncType{
		Returns: intrinsic[0],
		Params:  params,
	}
}

func (t *types) getIntrinsicTypes(function *ast.Func, intrinsicName string) []ir.Type {
	param := t.get(function.Params[0].Type)

	switch intrinsicName {
	case "abs":
		if isFloating(function.Params[0].Type) {
			return []ir.Type{
				param,
				param,
			}
		} else {
			return []ir.Type{
				param,
				param,
				ir.I1,
			}
		}

	case "pow", "min", "max", "copysign":
		return []ir.Type{param, param, param}

	case "sqrt", "sin", "cos", "exp", "exp2", "exp10", "log", "log2", "log10", "floor", "ceil", "round":
		return []ir.Type{param, param}

	case "fma":
		return []ir.Type{param, param, param, param}

	case "memcpy", "memmove":
		return []ir.Type{
			ir.Void,
			param,
			param,
			ir.I32,
			ir.I1,
		}

	case "memset":
		return []ir.Type{
			ir.Void,
			param,
			ir.I8,
			ir.I32,
			ir.I1,
		}

	default:
		panic("codegen.types.getIntrinsic() - Not implemented")
	}
}

func (t *types) getCachedType(type_ ast.Type) ir.Type {
	for _, cached := range t.types {
		if cached.type_.Equals(type_) {
			return cached.typ
		}
	}

	return nil
}

func (t *types) cacheType(type_ ast.Type, typ ir.Type) ir.Type {
	t.types = append(t.types, cachedType{
		type_: type_,
		typ:   typ,
	})

	return typ
}

// Meta

func (t *types) getMeta(type_ ast.Type) ir.MetaID {
	// Interface
	if _, ok := ast.As[*ast.Interface](type_); ok {
		if !t.interfaceMetadata.Valid() {
			void := ast.Primitive{Kind: ast.Void}
			ptr := ast.Pointer{Pointee: &void}

			t.interfaceMetadata = t.c.module.Meta(&ir.CompositeTypeMeta{
				Tag:   ir.StructureTypeTag,
				Name:  "__interface",
				Size:  abi.GetTargetAbi().Size(type_) * 8,
				Align: abi.GetTargetAbi().Align(type_) * 8,
				Elements: []ir.MetaID{
					t.c.module.Meta(&ir.DerivedTypeMeta{
						Tag:      ir.MemberTag,
						Name:     "vtable",
						BaseType: t.getMeta(&ptr),
						Offset:   0,
					}),
					t.c.module.Meta(&ir.DerivedTypeMeta{
						Tag:      ir.MemberTag,
						Name:     "data",
						BaseType: t.getMeta(&ptr),
						Offset:   abi.GetTargetAbi().Size(&ptr) * 8,
					}),
				},
			})
		}

		return t.interfaceMetadata
	}

	// Check cache
	for _, cached := range t.metadata {
		if cached.type_.Equals(type_) {
			return cached.id
		}
	}

	// Create
	switch type_ := type_.Resolved().(type) {
	case *ast.Primitive:
		if type_.Kind == ast.Void {
			return 0
		}

		typ := &ir.BasicTypeMeta{
			Size:  abi.GetTargetAbi().Size(type_) * 8,
			Align: abi.GetTargetAbi().Align(type_) * 8,
		}

		switch type_.Kind {
		case ast.Bool:
			typ.Name = "bool"
			typ.Encoding = ir.BooleanEncoding

		case ast.U8:
			typ.Name = "u8"
			typ.Encoding = ir.UnsignedEncoding
		case ast.U16:
			typ.Name = "u16"
			typ.Encoding = ir.UnsignedEncoding
		case ast.U32:
			typ.Name = "u32"
			typ.Encoding = ir.UnsignedEncoding
		case ast.U64:
			typ.Name = "u64"
			typ.Encoding = ir.UnsignedEncoding

		case ast.I8:
			typ.Name = "i8"
			typ.Encoding = ir.SignedEncoding
		case ast.I16:
			typ.Name = "i16"
			typ.Encoding = ir.SignedEncoding
		case ast.I32:
			typ.Name = "i32"
			typ.Encoding = ir.SignedEncoding
		case ast.I64:
			typ.Name = "i64"
			typ.Encoding = ir.SignedEncoding

		case ast.F32:
			typ.Name = "f32"
			typ.Encoding = ir.FloatEncoding
		case ast.F64:
			typ.Name = "f64"
			typ.Encoding = ir.FloatEncoding

		default:
			panic("codegen.types.getMeta() - ast.Primitive - Not implemented")
		}

		return t.cacheMeta(type_, typ)

	case *ast.Pointer:
		typ := &ir.DerivedTypeMeta{
			Tag:      ir.PointerTypeTag,
			BaseType: t.getMeta(type_.Pointee),
			Size:     abi.GetTargetAbi().Size(type_) * 8,
			Align:    abi.GetTargetAbi().Align(type_) * 8,
		}

		return t.cacheMeta(type_, typ)

	case *ast.Array:
		typ := &ir.CompositeTypeMeta{
			Tag:      ir.ArrayTypeTag,
			Size:     abi.GetTargetAbi().Size(type_) * 8,
			Align:    abi.GetTargetAbi().Align(type_) * 8,
			BaseType: t.getMeta(type_.Base),
			Elements: []ir.MetaID{t.c.module.Meta(&ir.SubrangeMeta{
				LowerBound: 0,
				Count:      type_.Count,
			})},
		}

		return t.cacheMeta(type_, typ)

	case *ast.Struct:
		fields, offsets := abi.GetStructLayout(type_).Fields(abi.GetTargetAbi(), type_)
		fieldsMeta := make([]ir.MetaID, len(fields))

		for i, field := range fields {
			fieldsMeta[i] = t.c.module.Meta(&ir.DerivedTypeMeta{
				Tag:      ir.MemberTag,
				Name:     field.Name.String(),
				BaseType: t.getMeta(field.Type),
				Offset:   offsets[i] * 8,
			})
		}

		typ := &ir.CompositeTypeMeta{
			Tag:      ir.StructureTypeTag,
			Name:     type_.Name.String(),
			Size:     abi.GetTargetAbi().Size(type_) * 8,
			Align:    abi.GetTargetAbi().Align(type_) * 8,
			Elements: fieldsMeta,
		}

		return t.cacheMeta(type_, typ)

	case *ast.Enum:
		cases := make([]ir.MetaID, len(type_.Cases))

		for i, case_ := range type_.Cases {
			cases[i] = t.c.module.Meta(&ir.EnumeratorMeta{
				Name:  case_.Name.String(),
				Value: ir.Signed(case_.ActualValue),
			})
		}

		typ := &ir.CompositeTypeMeta{
			Tag:      ir.EnumerationTypeTag,
			Name:     type_.Name.String(),
			Size:     abi.GetTargetAbi().Size(type_) * 8,
			Align:    abi.GetTargetAbi().Align(type_) * 8,
			BaseType: t.getMeta(type_.ActualType),
			Elements: cases,
		}

		return t.cacheMeta(type_, typ)

	case *ast.Func:
		this := type_.Method()

		parameterCount := len(type_.Params)
		if this != nil {
			parameterCount++
		}

		params := make([]ir.MetaID, parameterCount)

		if this != nil {
			type_ := ast.Pointer{Pointee: this}
			params[0] = t.getMeta(&type_)
		}

		for index, param := range type_.Params {
			i := index
			if this != nil {
				i++
			}

			params[i] = t.getMeta(param.Type)
		}

		typ := &ir.SubroutineTypeMeta{
			Returns: t.getMeta(type_.Returns),
			Params:  params,
		}

		return t.cacheMeta(type_, typ)

	default:
		panic("codegen.types.getMeta() - Not implemented")
	}
}

func (t *types) cacheMeta(type_ ast.Type, meta ir.Meta) ir.MetaID {
	id := t.c.module.Meta(meta)

	t.metadata = append(t.metadata, cachedTypeMeta{
		type_: type_,
		id:    id,
	})

	return id
}
