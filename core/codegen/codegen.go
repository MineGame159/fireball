package codegen

import (
	"fireball/core/architecture"
	"fireball/core/ast"
	"fireball/core/fuckoff"
	"fireball/core/llvm"
	"fireball/core/scanner"
	"io"
)

type codegen struct {
	path     string
	resolver fuckoff.Resolver

	types []typePair

	staticVariables map[*ast.Field]exprValue
	functions       map[*ast.Func]llvm.Value

	scopes    []scope
	variables []variable

	allocas map[ast.Node]exprValue

	function *llvm.Function
	block    *llvm.Block

	loopStart *llvm.Block
	loopEnd   *llvm.Block

	exprResult exprValue
	this       exprValue

	module *llvm.Module
}

type typePair struct {
	fireball ast.Type
	llvm     llvm.Type
}

type scope struct {
	variableI     int
	variableCount int
}

type exprValue struct {
	v           llvm.Value
	addressable bool
}

type variable struct {
	name  ast.Node
	value exprValue
}

func Emit(path string, resolver fuckoff.Resolver, file *ast.File, writer io.Writer) {
	// Init codegen
	c := &codegen{
		path:     path,
		resolver: resolver,

		staticVariables: make(map[*ast.Field]exprValue),
		functions:       make(map[*ast.Func]llvm.Value),

		module: llvm.NewModule(),
	}

	// File metadata
	c.module.Source(path)

	// Find some declarations
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
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
		}
	}

	// Emit
	for _, decl := range file.Decls {
		c.acceptDecl(decl)
	}

	// Write
	llvm.WriteText(c.module, writer)
}

func (c *codegen) defineOrDeclare(function *ast.Func) {
	t := c.getType(function)

	if function.HasBody() {
		// Define
		this := function.Method()

		f := c.module.Define(t, function.MangledName()[3:])
		c.functions[function] = f

		// Inline
		for _, attribute := range function.Attributes {
			if attribute.Name.String() == "Inline" {
				f.SetAlwaysInline()
				break
			}
		}

		// Set parameter names
		if this != nil {
			f.GetParameter(0).SetName("this")
		}

		for i, param := range function.Params {
			index := i
			if this != nil {
				index++
			}

			f.GetParameter(index).SetName(param.Name.String())
		}
	} else {
		// Declare
		c.functions[function] = c.module.Declare(t)
	}
}

// IR

func (c *codegen) load(value exprValue, type_ ast.Type) exprValue {
	if value.addressable {
		load := c.block.Load(value.v)
		load.SetAlign(type_.Align())

		return exprValue{
			v:           load,
			addressable: false,
		}
	}

	return value
}

func (c *codegen) loadExpr(expr ast.Expr) exprValue {
	return c.load(c.acceptExpr(expr), expr.Result().Type)
}

// Static variables

func (c *codegen) getStaticVariable(field *ast.Field) exprValue {
	// Get static variable already in this module
	if value, ok := c.staticVariables[field]; ok {
		return value
	}

	// Create static variable
	return c.createStaticVariable(field, true)
}

func (c *codegen) createStaticVariable(field *ast.Field, external bool) exprValue {
	ptr := ast.Pointer{Pointee: field.Type}

	llvmValue := c.module.Variable(external, c.getType(field.Type), c.getType(&ptr))
	llvmValue.SetName(field.GetMangledName())

	value := exprValue{
		v:           llvmValue,
		addressable: true,
	}

	c.staticVariables[field] = value
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

	value := c.module.Declare(c.getType(function))
	c.functions[function] = value

	return exprValue{v: value}
}

func (c *codegen) beginBlock(block *llvm.Block) {
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
		pointer := a.c.block.Alloca(a.c.getType(node.ActualType))
		pointer.SetName(node.Name.String() + ".var")
		pointer.SetAlign(node.ActualType.Align())

		a.c.allocas[node] = exprValue{
			v:           pointer,
			addressable: true,
		}

	case *ast.Call:
		if callNeedsTempVariable(node) {
			returns := node.Callee.Result().Function.Returns

			pointer := a.c.block.Alloca(a.c.getType(returns))
			pointer.SetAlign(returns.Align())

			a.c.allocas[node] = exprValue{
				v:           pointer,
				addressable: true,
			}
		}

	case *ast.Member:
		if node.Value.Result().Kind != ast.TypeResultKind && node.Result().Kind == ast.FunctionResultKind && !node.Value.Result().IsAddressable() {
			type_ := node.Value.Result().Type

			pointer := a.c.block.Alloca(a.c.getType(type_))
			pointer.SetAlign(type_.Align())

			a.c.allocas[node] = exprValue{
				v:           pointer,
				addressable: true,
			}
		}
	}

	node.AcceptChildren(a)
}

func callNeedsTempVariable(expr *ast.Call) bool {
	function := expr.Callee.Result().Function

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

// Types

func (c *codegen) getType(type_ ast.Type) llvm.Type {
	type_ = type_.Resolved()

	// Check cache
	for _, pair := range c.types {
		if pair.fireball.Equals(type_) {
			return pair.llvm
		}
	}

	// Create type
	var llvmType llvm.Type

	if v, ok := ast.As[*ast.Primitive](type_); ok {
		// Primitive
		switch v.Kind {
		case ast.Void:
			llvmType = c.module.Void()
		case ast.Bool:
			llvmType = c.module.Primitive("bool", 8, llvm.BooleanEncoding)

		case ast.U8:
			llvmType = c.module.Primitive("u8", 8, llvm.UnsignedEncoding)
		case ast.U16:
			llvmType = c.module.Primitive("u16", 16, llvm.UnsignedEncoding)
		case ast.U32:
			llvmType = c.module.Primitive("u32", 32, llvm.UnsignedEncoding)
		case ast.U64:
			llvmType = c.module.Primitive("u64", 64, llvm.UnsignedEncoding)

		case ast.I8:
			llvmType = c.module.Primitive("i8", 8, llvm.SignedEncoding)
		case ast.I16:
			llvmType = c.module.Primitive("i16", 16, llvm.SignedEncoding)
		case ast.I32:
			llvmType = c.module.Primitive("i32", 32, llvm.SignedEncoding)
		case ast.I64:
			llvmType = c.module.Primitive("i64", 64, llvm.SignedEncoding)

		case ast.F32:
			llvmType = c.module.Primitive("f32", 32, llvm.FloatEncoding)
		case ast.F64:
			llvmType = c.module.Primitive("f64", 64, llvm.FloatEncoding)

		default:
			panic("codegen.getType() - Not implemented")
		}
	} else if v, ok := ast.As[*ast.Array](type_); ok {
		// Array
		llvmType = c.module.Array(v.String(), v.Count, c.getType(v.Base))
	} else if v, ok := ast.As[*ast.Pointer](type_); ok {
		// Pointer
		llvmType = c.module.Pointer(v.String(), c.getType(v.Pointee))
	} else if v, ok := ast.As[*ast.Func](type_); ok {
		// Function
		var parameters []llvm.Type
		var returns llvm.Type

		intrinsicName := v.IntrinsicName()

		if intrinsicName != "" {
			intrinsic := c.getIntrinsic(v, intrinsicName)

			parameters = intrinsic[1:]
			returns = intrinsic[0]
		} else {
			this := v.Method()

			parameterCount := len(v.Params)
			if this != nil {
				parameterCount++
			}

			parameters = make([]llvm.Type, parameterCount)

			if this != nil {
				type_ := ast.Pointer{Pointee: this}
				parameters[0] = c.getType(&type_)
			}

			for i, param := range v.Params {
				index := i
				if this != nil {
					index++
				}

				parameters[index] = c.getType(param.Type)
			}

			returns = c.getType(v.Returns)
		}

		llvmType = c.module.Function(getMangledName(v), parameters, v.IsVariadic(), returns)
	} else if v, ok := ast.As[*ast.Struct](type_); ok {
		// Struct
		layout := architecture.CLayout{}
		fields := make([]llvm.Field, len(v.Fields))

		for i, field := range v.Fields {
			offset := layout.Add(field.Type.Size(), field.Type.Align())

			fields[i] = llvm.Field{
				Name:   field.Name.String(),
				Type:   c.getType(field.Type),
				Offset: offset * 8,
			}
		}

		llvmType = c.module.Struct(v.Name.String(), layout.Size()*8, fields)
	} else if v, ok := ast.As[*ast.Enum](type_); ok {
		// Enum
		llvmType = c.module.Alias(v.Name.String(), c.getType(v.ActualType))
	}

	if llvmType != nil {
		c.types = append(c.types, typePair{
			fireball: type_,
			llvm:     llvmType,
		})

		return llvmType
	}

	panic("codegen.getType() - Invalid type")
}

func (c *codegen) getIntrinsic(function *ast.Func, intrinsicName string) []llvm.Type {
	param := c.getType(function.Params[0].Type)

	switch intrinsicName {
	case "abs":
		if isFloating(function.Params[0].Type) {
			return []llvm.Type{
				param,
				param,
			}
		} else {
			i1 := c.getPrimitiveType(ast.Bool)

			return []llvm.Type{
				param,
				param,
				i1,
			}
		}

	case "pow", "min", "max", "copysign":
		return []llvm.Type{param, param, param}

	case "sqrt", "sin", "cos", "exp", "exp2", "exp10", "log", "log2", "log10", "floor", "ceil", "round":
		return []llvm.Type{param, param}

	case "fma":
		return []llvm.Type{param, param, param, param}

	case "memcpy", "memmove":
		void := c.getPrimitiveType(ast.Void)
		i1 := c.getPrimitiveType(ast.Bool)
		i32 := c.getPrimitiveType(ast.I32)

		return []llvm.Type{
			void,
			param,
			param,
			i32,
			i1,
		}

	case "memset":
		void := c.getPrimitiveType(ast.Void)
		i1 := c.getPrimitiveType(ast.Bool)
		i8 := c.getPrimitiveType(ast.I8)
		i32 := c.getPrimitiveType(ast.I32)

		return []llvm.Type{
			void,
			param,
			i8,
			i32,
			i1,
		}

	default:
		panic("codegen.getIntrinsic() - Invalid intrinsic")
	}
}

func (c *codegen) modifyIntrinsicArgs(function *ast.Func, intrinsicName string, args []llvm.Value) []llvm.Value {
	switch intrinsicName {
	case "abs":
		if !isFloating(function.Returns) {
			i1 := c.getPrimitiveType(ast.Bool)
			args = append(args, c.function.Literal(i1, llvm.Literal{}))
		}

	case "memcpy", "memmove", "memset":
		i1 := c.getPrimitiveType(ast.Bool)
		args = append(args, c.function.Literal(i1, llvm.Literal{}))
	}

	return args
}

func getMangledName(function *ast.Func) string {
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

func (c *codegen) getPrimitiveType(kind ast.PrimitiveKind) llvm.Type {
	type_ := ast.Primitive{Kind: kind}
	return c.getType(&type_)
}

// Scope / Variables

func (c *codegen) getVariable(name scanner.Token) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.String() == name.Lexeme {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *codegen) addVariable(name ast.Node, value exprValue) *variable {
	c.block.Variable(name.String(), value.v).SetLocation(name.Cst())

	value.addressable = true

	c.variables = append(c.variables, variable{
		name:  name,
		value: value,
	})

	c.peekScope().variableCount++
	return &c.variables[len(c.variables)-1]
}

func (c *codegen) pushScope() {
	c.scopes = append(c.scopes, scope{
		variableI:     len(c.variables),
		variableCount: 0,
	})
}

func (c *codegen) popScope() {
	c.variables = c.variables[:c.peekScope().variableI]
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *codegen) peekScope() *scope {
	return &c.scopes[len(c.scopes)-1]
}

// Accept

func (c *codegen) acceptDecl(decl ast.Decl) {
	if decl != nil {
		decl.AcceptDecl(c)
	}
}

func (c *codegen) acceptStmt(stmt ast.Stmt) {
	if stmt != nil {
		stmt.AcceptStmt(c)
	}
}

func (c *codegen) acceptExpr(expr ast.Expr) exprValue {
	if expr != nil {
		expr.AcceptExpr(c)
		return c.exprResult
	}

	return exprValue{}
}
