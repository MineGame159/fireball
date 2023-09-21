package codegen

import (
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

	functions map[*ast.Func]llvm.Value

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

		functions: make(map[*ast.Func]llvm.Value),

		module: llvm.NewModule(),
	}

	// File metadata
	c.module.Source(path)

	// Define and declare functions
	for _, decl := range decls {
		if impl, ok := decl.(*ast.Impl); ok {
			for _, decl := range impl.Functions {
				if function, ok := decl.(*ast.Func); ok {
					c.defineOrDeclare(function)
				}
			}
		} else if function, ok := decl.(*ast.Func); ok {
			c.defineOrDeclare(function)
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

	if function.IsExtern() {
		// Declare
		c.functions[function] = c.module.Declare(t)
	} else {
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
	if call, ok := expr.(*ast.Call); ok && callNeedsTempVariable(call) {
		a.c.allocas[call] = exprValue{
			v:           a.c.block.Alloca(a.c.getType(call.Callee.Result().Function.Returns)),
			addressable: true,
		}
	} else if member, ok := expr.(*ast.Member); ok {
		if member.Result().Kind == ast.FunctionResultKind && !member.Value.Result().IsAddressable() {
			a.c.allocas[member] = exprValue{
				v:           a.c.block.Alloca(a.c.getType(member.Value.Result().Type)),
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
		this := v.Method()

		parameterCount := len(v.Params)
		if this != nil {
			parameterCount++
		}

		parameters := make([]llvm.Type, parameterCount)

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

		llvmType = c.module.Function(v.MangledName(), parameters, v.IsVariadic(), c.getType(v.Returns))
	} else if v, ok := type_.(*ast.Struct); ok {
		// Struct
		fields := make([]llvm.Field, len(v.Fields))

		for i, field := range v.Fields {
			fields[i] = llvm.Field{
				Name: field.Name.Lexeme,
				Type: c.getType(field.Type),
			}
		}

		llvmType = c.module.Struct(v.Name.Lexeme, fields)
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
