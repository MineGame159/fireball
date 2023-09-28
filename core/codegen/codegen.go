package codegen

import (
	"fireball/core/architecture"
	"fireball/core/ast"
	"fireball/core/llvm"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"io"
)

type codegen struct {
	path     string
	resolver utils.Resolver

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
	fireball types.Type
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
	name  scanner.Token
	value exprValue
}

func Emit(path string, resolver utils.Resolver, decls []ast.Decl, writer io.Writer) {
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
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *ast.Struct:
			for i := range decl.StaticFields {
				c.createStaticVariable(&decl.StaticFields[i], false)
			}

		case *ast.Impl:
			for _, decl := range decl.Functions {
				if function, ok := decl.(*ast.Func); ok {
					c.defineOrDeclare(function)
				}
			}

		case *ast.Func:
			c.defineOrDeclare(decl)
		}
	}

	// Emit
	for _, decl := range decls {
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

		// Set parameter names
		if this != nil {
			f.GetParameter(0).SetName("this")
		}

		for i, param := range function.Params {
			index := i
			if this != nil {
				index++
			}

			f.GetParameter(index).SetName(param.Name.Lexeme)
		}
	} else {
		// Declare
		c.functions[function] = c.module.Declare(t)
	}
}

// IR

func (c *codegen) load(value exprValue) exprValue {
	if value.addressable {
		return exprValue{
			v:           c.block.Load(value.v),
			addressable: false,
		}
	}

	return value
}

func (c *codegen) loadExpr(expr ast.Expr) exprValue {
	return c.load(c.acceptExpr(expr))
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
	ptr := types.PointerType{Pointee: field.Type}

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
	_, filePath := c.resolver.GetFunction(function.Name.Lexeme)

	if filePath == c.path {
		panic("codegen.getFunction() - Local function not found in functions map")
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
	a.AcceptDecl(function)
}

type allocaFinder struct {
	c *codegen
}

func (a *allocaFinder) AcceptDecl(decl ast.Decl) {
	decl.AcceptChildren(a)
}

func (a *allocaFinder) AcceptStmt(stmt ast.Stmt) {
	if variable, ok := stmt.(*ast.Variable); ok {
		pointer := a.c.block.Alloca(a.c.getType(variable.Type))
		pointer.SetName(variable.Name.Lexeme + ".var")

		a.c.allocas[variable] = exprValue{
			v:           pointer,
			addressable: true,
		}
	}

	stmt.AcceptChildren(a)
}

func (a *allocaFinder) AcceptExpr(expr ast.Expr) {
	switch expr := expr.(type) {
	case *ast.Call:
		if callNeedsTempVariable(expr) {
			a.c.allocas[expr] = exprValue{
				v:           a.c.block.Alloca(a.c.getType(expr.Callee.Result().Function.Returns)),
				addressable: true,
			}
		}

	case *ast.Member:
		if expr.Value.Result().Kind != ast.TypeResultKind && expr.Result().Kind == ast.FunctionResultKind && !expr.Value.Result().IsAddressable() {
			a.c.allocas[expr] = exprValue{
				v:           a.c.block.Alloca(a.c.getType(expr.Value.Result().Type)),
				addressable: true,
			}
		}
	}

	expr.AcceptChildren(a)
}

func callNeedsTempVariable(expr *ast.Call) bool {
	function := expr.Callee.Result().Function

	if f, ok := expr.Callee.Result().Type.(*ast.Func); ok && function == nil {
		function = f
	}

	if _, ok := expr.Parent().(*ast.Expression); !ok && !types.IsPrimitive(function.Returns, types.Void) {
		if _, ok := function.Returns.(*types.ArrayType); ok {
			if _, ok := expr.Parent().(*ast.Index); ok {
				return true
			}
		}
	}

	return false
}

// Types

func (c *codegen) getType(type_ types.Type) llvm.Type {
	// Check cache
	for _, pair := range c.types {
		if pair.fireball.Equals(type_) {
			return pair.llvm
		}
	}

	// Create type
	var llvmType llvm.Type

	if v, ok := type_.(*types.PrimitiveType); ok {
		// Primitive
		switch v.Kind {
		case types.Void:
			llvmType = c.module.Void()
		case types.Bool:
			llvmType = c.module.Primitive("bool", 8, llvm.BooleanEncoding)

		case types.U8:
			llvmType = c.module.Primitive("u8", 8, llvm.UnsignedEncoding)
		case types.U16:
			llvmType = c.module.Primitive("u16", 16, llvm.UnsignedEncoding)
		case types.U32:
			llvmType = c.module.Primitive("u32", 32, llvm.UnsignedEncoding)
		case types.U64:
			llvmType = c.module.Primitive("u64", 64, llvm.UnsignedEncoding)

		case types.I8:
			llvmType = c.module.Primitive("i8", 8, llvm.SignedEncoding)
		case types.I16:
			llvmType = c.module.Primitive("i16", 16, llvm.SignedEncoding)
		case types.I32:
			llvmType = c.module.Primitive("i32", 32, llvm.SignedEncoding)
		case types.I64:
			llvmType = c.module.Primitive("i64", 64, llvm.SignedEncoding)

		case types.F32:
			llvmType = c.module.Primitive("f32", 32, llvm.FloatEncoding)
		case types.F64:
			llvmType = c.module.Primitive("f64", 64, llvm.FloatEncoding)
		}
	} else if v, ok := type_.(*types.ArrayType); ok {
		// Array
		llvmType = c.module.Array(v.String(), int(v.Count), c.getType(v.Base))
	} else if v, ok := type_.(*types.PointerType); ok {
		// Pointer
		llvmType = c.module.Pointer(v.String(), c.getType(v.Pointee))
	} else if v, ok := type_.(*ast.Func); ok {
		// Function
		var parameters []llvm.Type
		var returns llvm.Type

		if v.IsIntrinsic() {
			intrinsic := c.getIntrinsic(v)

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
				type_ := types.PointerType{Pointee: this}
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
	} else if v, ok := type_.(*ast.Struct); ok {
		// Struct
		layout := architecture.CLayout{}
		fields := make([]llvm.Field, len(v.Fields))

		for i, field := range v.Fields {
			offset := layout.Add(field.Type)

			fields[i] = llvm.Field{
				Name:   field.Name.Lexeme,
				Type:   c.getType(field.Type),
				Offset: offset * 8,
			}
		}

		llvmType = c.module.Struct(v.Name.Lexeme, layout.Size()*8, fields)
	} else if v, ok := type_.(*ast.Enum); ok {
		// Enum
		llvmType = c.module.Alias(v.Name.Lexeme, c.getType(v.Type))
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

func (c *codegen) getIntrinsic(function *ast.Func) []llvm.Type {
	param := c.getType(function.Params[0].Type)

	switch function.Name.Lexeme {
	case "abs":
		if isFloating(function.Params[0].Type) {
			return []llvm.Type{
				param,
				param,
			}
		} else {
			i1 := c.getPrimitiveType(types.Bool)

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
		void := c.getPrimitiveType(types.Void)
		i1 := c.getPrimitiveType(types.Bool)
		i32 := c.getPrimitiveType(types.I32)

		return []llvm.Type{
			void,
			param,
			param,
			i32,
			i1,
		}

	case "memset":
		void := c.getPrimitiveType(types.Void)
		i1 := c.getPrimitiveType(types.Bool)
		i8 := c.getPrimitiveType(types.I8)
		i32 := c.getPrimitiveType(types.I32)

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

func (c *codegen) modifyIntrinsicArgs(function *ast.Func, args []llvm.Value) []llvm.Value {
	switch function.Name.Lexeme {
	case "abs":
		if !isFloating(function.Returns) {
			i1 := c.getPrimitiveType(types.Bool)
			args = append(args, c.function.Literal(i1, llvm.Literal{}))
		}

	case "memcpy", "memmove", "memset":
		i1 := c.getPrimitiveType(types.Bool)
		args = append(args, c.function.Literal(i1, llvm.Literal{}))
	}

	return args
}

func getMangledName(function *ast.Func) string {
	if function.IsIntrinsic() {
		name := ""

		switch function.Name.Lexeme {
		case "abs":
			name = ternary(isFloating(function.Returns), "llvm.fabs", "llvm.abs")

		case "min":
			name = ternary(isFloating(function.Returns), "llvm.minnum", ternary(isSigned(function.Returns), "llvm.smin", "llvm.umin"))

		case "max":
			name = ternary(isFloating(function.Returns), "llvm.maxnum", ternary(isSigned(function.Returns), "llvm.smax", "llvm.umax"))

		case "sqrt", "pow", "sin", "cos", "exp", "exp2", "exp10", "log", "log2", "log10", "fma", "copysign", "floor", "ceil", "round":
			name = "llvm." + function.Name.Lexeme

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

func isSigned(type_ types.Type) bool {
	if v, ok := type_.(*types.PrimitiveType); ok {
		return types.IsSigned(v.Kind)
	}

	return false
}

func isFloating(type_ types.Type) bool {
	if v, ok := type_.(*types.PrimitiveType); ok {
		return types.IsFloating(v.Kind)
	}

	return false
}

func ternary[T any](condition bool, true T, false T) T {
	if condition {
		return true
	}

	return false
}

func (c *codegen) getPrimitiveType(kind types.PrimitiveKind) llvm.Type {
	type_ := types.PrimitiveType{Kind: kind}
	return c.getType(&type_)
}

// Scope / Variables

func (c *codegen) getVariable(name scanner.Token) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.Lexeme == name.Lexeme {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *codegen) addVariable(name scanner.Token, value exprValue) *variable {
	c.block.Variable(name.Lexeme, value.v).SetLocation(name)

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
		decl.Accept(c)
	}
}

func (c *codegen) acceptStmt(stmt ast.Stmt) {
	if stmt != nil {
		stmt.Accept(c)
	}
}

func (c *codegen) acceptExpr(expr ast.Expr) exprValue {
	if expr != nil {
		expr.Accept(c)
		return c.exprResult
	}

	return exprValue{}
}
