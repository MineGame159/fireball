package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
	"fmt"
)

type codegen struct {
	ctx      *Context
	path     string
	resolver *ast.CombinedResolver

	types   types
	vtables vtables
	scopes  scopes

	staticVariables map[ast.Node]exprValue
	functions       map[*ast.Func]*ir.Func

	allocas map[ast.Node]exprValue

	astFunction *ast.Func
	function    *ir.Func
	block       *ir.Block

	loopStart *ir.Block
	loopEnd   *ir.Block

	exprResult exprValue
	this       exprValue

	module *ir.Module
}

type exprValue struct {
	v           ir.Value
	addressable bool
}

func Emit(ctx *Context, path string, root ast.RootResolver, file *ast.File) *ir.Module {
	// Init codegen
	c := &codegen{
		ctx:      ctx,
		path:     path,
		resolver: ast.NewCombinedResolver(root),

		staticVariables: make(map[ast.Node]exprValue),
		functions:       make(map[*ast.Func]*ir.Func),

		module: &ir.Module{},
	}

	// File
	c.module.Path = path

	c.types.module = c.module
	c.vtables.c = c
	c.scopes.c = c

	c.scopes.pushFile(path)

	// Find some declarations
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.Using:
			if resolver := root.GetResolver(decl.Name); resolver != nil {
				c.resolver.Add(resolver)
			}

		case *ast.Struct:
			for i := range decl.StaticFields {
				c.createStaticVariable(decl.StaticFields[i], false)
			}

		case *ast.Impl:
			for _, function := range decl.Methods {
				c.defineOrDeclare(function)
			}

		case *ast.Func:
			c.defineOrDeclare(decl)

		case *ast.GlobalVar:
			c.createGlobalVariable(decl, false)
		}
	}

	// Emit
	for _, decl := range file.Decls {
		c.acceptDecl(decl)
	}

	// Return
	return c.module
}

func (c *codegen) defineOrDeclare(function *ast.Func) {
	t := c.types.get(function).(*ir.FuncType)
	name := c.getMangledName(function)

	if function.HasBody() {
		// Meta
		meta := &ir.SubprogamMeta{
			Name:        function.Name.String(),
			LinkageName: name,
			Scope:       c.scopes.getMeta(),
			File:        c.scopes.file,
			Type:        c.types.getMeta(function),
			Unit:        c.scopes.unitId,
		}

		if function.Cst() != nil {
			meta.Line = uint32(function.Cst().Range.Start.Line)
		}

		// Define
		var flags ir.FuncFlags

		for _, attribute := range function.Attributes {
			if attribute.Name.String() == "Inline" {
				flags |= ir.InlineFlag
				break
			}
		}

		f := c.module.Define(name, t, flags)
		f.SetMeta(c.module.Meta(meta))

		c.functions[function] = f
	} else {
		// Declare
		c.functions[function] = c.module.Declare(name, t)
	}
}

func (c *codegen) getMangledName(function *ast.Func) string {
	intrinsicName := function.IntrinsicName()

	if intrinsicName != "" {
		name := ""

		switch intrinsicName {
		case "abs":
			name = ternary(isFloating(function.Returns), "llvm.fabs", "llvm.abs")

		case "min":
			name = ternary(isFloating(function.Returns), "llvm.minnum", ternary(isSigned(function.Returns), "llvm.smin", "llvm.umin"))

		case "max":
			name = ternary(isFloating(function.Returns), "llvm.maxnum", ternary(isSigned(function.Returns), "llvm.smax", "llvm.umax"))

		case "sqrt", "pow", "sin", "cos", "exp", "exp2", "exp10", "log", "log2", "log10", "fma", "copysign", "floor", "ceil", "round":
			name = "llvm." + intrinsicName

		case "memcpy", "memmove":
			return "llvm.memcpy.p0.p0.i32"

		case "memset":
			return "llvm.memset.p0.i32"
		}

		if name == "" {
			panic("codegen getMangledName() - Invalid intrinsic")
		}

		return name + "." + function.Returns.String()
	}

	return function.MangledName()
}

// IR

func (c *codegen) load(value exprValue, type_ ast.Type) exprValue {
	if value.addressable {
		return exprValue{
			v: c.block.Add(&ir.LoadInst{
				Typ:     value.v.Type().(*ir.PointerType).Pointee,
				Pointer: value.v,
				Align:   type_.Align() * 8,
			}),
			addressable: false,
		}
	}

	return value
}

func (c *codegen) loadExpr(expr ast.Expr) exprValue {
	return c.load(c.acceptExpr(expr), expr.Result().Type)
}

func (c *codegen) implicitCast(required ast.Type, value exprValue, valueType ast.Type) exprValue {
	if kind, ok := ast.GetImplicitCast(valueType, required); ok && kind != ast.None {
		return c.cast(value, valueType, required, nil)
	}

	return value
}

func (c *codegen) implicitCastLoadExpr(required ast.Type, expr ast.Expr) exprValue {
	return c.implicitCast(required, c.loadExpr(expr), expr.Result().Type)
}

func (c *codegen) cast(value exprValue, from, to ast.Type, location ast.Node) exprValue {
	kind, ok := ast.GetCast(from, to)
	if !ok {
		panic("codegen.convertAstCastKind() - ast.GetCast() returned false")
	}

	return c.convertCast(value, kind, from, to, location)
}

func (c *codegen) convertCast(value exprValue, kind ast.CastKind, from, to ast.Type, location ast.Node) exprValue {
	if kind == ast.None {
		return value
	}

	value = c.load(value, from)
	toIr := c.types.get(to)

	switch kind {
	case ast.Truncate:
		result := c.block.Add(&ir.TruncInst{
			Value: value.v,
			Typ:   toIr,
		})

		c.setLocationMeta(result, location)
		return exprValue{v: result}

	case ast.Extend:
		var result ir.MetaValue

		if ast.IsFloating(to.Resolved().(*ast.Primitive).Kind) {
			result = c.block.Add(&ir.FExtInst{
				Value: value.v,
				Typ:   toIr,
			})
		} else {
			signed := ast.IsSigned(to.Resolved().(*ast.Primitive).Kind)

			if from, ok := ast.As[*ast.Primitive](from); !ok || !ast.IsSigned(from.Kind) {
				signed = false
			}

			result = c.block.Add(&ir.ExtInst{
				SignExtend: signed,
				Value:      value.v,
				Typ:        toIr,
			})
		}

		c.setLocationMeta(result, location)
		return exprValue{v: result}

	case ast.Int2Float:
		result := c.block.Add(&ir.I2FInst{
			Signed: ast.IsSigned(from.Resolved().(*ast.Primitive).Kind),
			Value:  value.v,
			Typ:    toIr,
		})

		c.setLocationMeta(result, location)
		return exprValue{v: result}

	case ast.Float2Int:
		result := c.block.Add(&ir.F2IInst{
			Signed: ast.IsSigned(to.Resolved().(*ast.Primitive).Kind),
			Value:  value.v,
			Typ:    toIr,
		})

		c.setLocationMeta(result, location)
		return exprValue{v: result}

	case ast.Pointer2Interface:
		type_ := from.Resolved().(*ast.Pointer).Pointee

		result := c.block.Add(&ir.InsertValueInst{
			Value:   &ir.ZeroInitConst{Typ: c.types.get(to)},
			Element: c.vtables.get(type_, to),
			Indices: []uint32{0},
		})
		c.setLocationMeta(result, location)

		result = c.block.Add(&ir.InsertValueInst{
			Value:   result,
			Element: value.v,
			Indices: []uint32{1},
		})
		c.setLocationMeta(result, location)

		return exprValue{v: result}

	default:
		panic("codegen.convertAstCastKind() - Not implemented")
	}
}

// Static / Global variables

func (c *codegen) getStaticVariable(field *ast.Field) exprValue {
	// Get static variable already in this module
	if value, ok := c.staticVariables[field]; ok {
		return value
	}

	// Create static variable
	return c.createStaticVariable(field, true)
}

func (c *codegen) getGlobalVariable(variable *ast.GlobalVar) exprValue {
	// Get static variable already in this module
	if value, ok := c.staticVariables[variable]; ok {
		return value
	}

	// Create static variable
	return c.createGlobalVariable(variable, true)
}

func (c *codegen) createStaticVariable(field *ast.Field, external bool) exprValue {
	type_ := c.types.get(field.Type)
	var initializer ir.Value

	if !external {
		initializer = &ir.ZeroInitConst{Typ: type_}
	}

	global := c.module.Global(field.MangledName(), type_, initializer)

	if !external {
		meta := &ir.GlobalVarMeta{
			Name:        fmt.Sprintf("%s.%s", field.Parent().(*ast.Struct).Name, field.Name),
			LinkageName: field.MangledName(),
			Scope:       c.scopes.getMeta(),
			File:        c.scopes.file,
			Type:        c.types.getMeta(field.Type),
			Local:       false,
			Definition:  true,
		}

		if field.Cst() != nil {
			meta.Line = uint32(field.Cst().Range.Start.Line)
		}

		id := c.module.Meta(&ir.GlobalVarExpr{Var: c.module.Meta(meta)})

		global.SetMeta(id)
		c.scopes.unitMeta.Globals = append(c.scopes.unitMeta.Globals, id)
	}

	value := exprValue{
		v:           global,
		addressable: true,
	}

	c.staticVariables[field] = value
	return value
}

func (c *codegen) createGlobalVariable(variable *ast.GlobalVar, external bool) exprValue {
	type_ := c.types.get(variable.Type)
	var initializer ir.Value

	if !external {
		initializer = &ir.ZeroInitConst{Typ: type_}
	}

	global := c.module.Global(variable.MangledName(), type_, initializer)

	if !external {
		meta := &ir.GlobalVarMeta{
			Name:        variable.Name.String(),
			LinkageName: variable.MangledName(),
			Scope:       c.scopes.getMeta(),
			File:        c.scopes.file,
			Type:        c.types.getMeta(variable.Type),
			Local:       false,
			Definition:  true,
		}

		if variable.Cst() != nil {
			meta.Line = uint32(variable.Cst().Range.Start.Line)
		}

		id := c.module.Meta(&ir.GlobalVarExpr{Var: c.module.Meta(meta)})

		global.SetMeta(id)
		c.scopes.unitMeta.Globals = append(c.scopes.unitMeta.Globals, id)
	}

	value := exprValue{
		v:           global,
		addressable: true,
	}

	c.staticVariables[variable] = value
	return value
}

// Functions

func (c *codegen) getFunction(function *ast.Func) exprValue {
	// Get function already in this module
	for f, value := range c.functions {
		if f.Equals(function) {
			return exprValue{v: value}
		}
	}

	// Resolve function from project
	if f := c.resolver.GetFunction(function.Name.String()); f != nil {
		filePath := ast.GetParent[*ast.File](f).Path

		if filePath == c.path {
			panic("codegen.getFunction() - Local function not found in functions map")
		}
	}

	value := c.module.Declare(c.getMangledName(function), c.types.get(function).(*ir.FuncType))
	c.functions[function] = value

	return exprValue{v: value}
}

func (c *codegen) beginBlock(block *ir.Block) {
	c.block = block
}

func (c *codegen) findAllocas(function *ast.Func) {
	c.allocas = make(map[ast.Node]exprValue)

	a := &allocaFinder{c: c}
	a.VisitNode(function)
}

type allocaFinder struct {
	c *codegen
}

func (a *allocaFinder) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Var:
		a.c.allocas[node] = exprValue{
			v:           a.c.alloca(node.ActualType, node.Name.String()+".var", node),
			addressable: true,
		}

	case *ast.Call:
		if callNeedsTempVariable(node) {
			returns := node.Callee.Result().Type.(*ast.Func).Returns

			a.c.allocas[node] = exprValue{
				v:           a.c.alloca(returns, "", node),
				addressable: true,
			}
		}

	case *ast.Member:
		if node.Value.Result().Kind != ast.TypeResultKind && node.Result().Kind == ast.CallableResultKind && !node.Value.Result().IsAddressable() {
			type_ := node.Value.Result().Type

			if type_ != nil {
				a.c.allocas[node] = exprValue{
					v:           a.c.alloca(type_, "", node),
					addressable: true,
				}
			}
		}
	}

	node.AcceptChildren(a)
}

func callNeedsTempVariable(expr *ast.Call) bool {
	function := expr.Callee.Result().Type.(*ast.Func)

	if f, ok := ast.As[*ast.Func](expr.Callee.Result().Type); ok && function == nil {
		function = f
	}

	if _, ok := expr.Parent().(*ast.Expression); !ok && !ast.IsPrimitive(function.Returns, ast.Void) {
		if _, ok := ast.As[*ast.Array](function.Returns); ok {
			if _, ok := expr.Parent().(*ast.Index); ok {
				return true
			}
		}
	}

	return false
}

func (c *codegen) modifyIntrinsicArgs(function *ast.Func, intrinsicName string, args []ir.Value) []ir.Value {
	switch intrinsicName {
	case "abs":
		if !isFloating(function.Returns) {
			args = append(args, ir.False)
		}

	case "memcpy", "memmove", "memset":
		args = append(args, ir.False)
	}

	return args
}

func isSigned(type_ ast.Type) bool {
	if v, ok := ast.As[*ast.Primitive](type_); ok {
		return ast.IsSigned(v.Kind)
	}

	return false
}

func isFloating(type_ ast.Type) bool {
	if v, ok := ast.As[*ast.Primitive](type_); ok {
		return ast.IsFloating(v.Kind)
	}

	return false
}

func ternary[T any](condition bool, true T, false T) T {
	if condition {
		return true
	}

	return false
}

// Accept

func (c *codegen) acceptDecl(decl ast.Decl) {
	if decl != nil {
		decl.AcceptDecl(c)
	}
}

func (c *codegen) acceptStmt(stmt ast.Stmt) bool {
	if stmt != nil && c.block != nil {
		stmt.AcceptStmt(c)
	}

	return c.block != nil
}

func (c *codegen) acceptExpr(expr ast.Expr) exprValue {
	if expr != nil && c.block != nil {
		expr.AcceptExpr(c)
		return c.exprResult
	}

	return exprValue{}
}
