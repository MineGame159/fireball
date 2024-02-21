package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
	"strings"
)

type vtableInfo struct {
	type_ ast.Type
	inter ast.Type
	value ir.Value
}

type vtables struct {
	c *codegen

	cache []vtableInfo
}

func (v *vtables) get(type_, inter ast.Type) ir.Value {
	// Check cache
	for _, vtable := range v.cache {
		if vtable.type_.Equals(type_) && vtable.inter.Equals(inter) {
			return vtable.value
		}
	}

	// Create
	impl := ast.GetParent[*ast.File](type_).Resolver.GetImpl(type_, inter.Resolved().(*ast.Interface))
	methods := make([]ir.Value, len(impl.Methods))

	for i, method := range impl.Methods {
		methods[i] = v.c.getFunction(method).v
	}

	typ := v.getType(inter.Resolved().(*ast.Interface))

	value := v.c.module.Constant(
		getVtableName(type_, inter),
		&ir.StructConst{
			Typ: typ,
			Fields: []ir.Value{
				&ir.IntConst{
					Typ:   ir.I32,
					Value: ir.Unsigned(uint64(v.c.ctx.GetTypeID(type_))),
				},
				&ir.ArrayConst{
					Typ:    typ.Fields[1],
					Values: methods,
				},
			},
		},
	)

	v.cache = append(v.cache, vtableInfo{
		type_: type_,
		inter: inter,
		value: value,
	})

	return value
}

func (v *vtables) getType(inter *ast.Interface) *ir.StructType {
	funcPtrArrayType := &ir.ArrayType{
		Count: uint32(len(inter.Methods)),
		Base:  &ir.PointerType{},
	}

	return &ir.StructType{
		Name: "",
		Fields: []ir.Type{
			ir.I32,
			funcPtrArrayType,
		},
	}
}

func getVtableName(type_, inter ast.Type) string {
	sb := strings.Builder{}
	sb.WriteString("__fb_vtable__")

	writeFullTypeName(&sb, type_)
	sb.WriteString("__")
	writeFullTypeName(&sb, inter)

	return sb.String()
}

func writeFullTypeName(sb *strings.Builder, type_ ast.Type) {
	file := ast.GetParent[*ast.File](type_)

	for i, part := range file.Namespace.Name.Parts {
		if i > 0 {
			sb.WriteRune('_')
		}

		sb.WriteString(part.String())
	}

	sb.WriteRune('_')

	switch type_ := type_.Resolved().(type) {
	case ast.StructType:
		sb.WriteString(type_.Underlying().Name.String())
	case *ast.Interface:
		sb.WriteString(type_.Name.String())

	default:
		panic("codegen.vtables.getVtableName() - Not implemented")
	}
}
