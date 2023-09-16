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

	functions []llvm.Value

	scopes    []scope
	variables []variable

	function *llvm.Function
	block    *llvm.Block

	loopStart *llvm.Block
	loopEnd   *llvm.Block

	exprResult exprValue

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

		module: llvm.NewModule(),
	}

	// File metadata
	c.module.Source(path)

	// Define and declare functions
	for _, decl := range decls {
		if function, ok := decl.(*ast.Func); ok {
			t := c.createFunctionType(function)

			if function.Extern {
				// Declare
				c.functions = append(c.functions, c.module.Declare(t))
			} else {
				// Define
				f := c.module.Define(t)
				c.functions = append(c.functions, f)

				// Set parameter names
				for i, param := range function.Params {
					f.GetParameter(i).SetName(param.Name.Lexeme)
				}
			}
		}
	}

	// Emit
	for _, decl := range decls {
		c.acceptDecl(decl)
	}

	// Write
	llvm.WriteText(c.module, writer)
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

func mangleName(name string) string {
	return "fb$" + name
}

func (c *codegen) createFunctionType(function *ast.Func) llvm.Type {
	parameters := make([]llvm.Type, len(function.Params))

	for i, param := range function.Params {
		parameters[i] = c.getType(param.Type)
	}

	name := function.Name.Lexeme
	if !function.Extern {
		name = mangleName(name)
	}

	return c.module.Function(name, parameters, function.Variadic, c.getType(function.Returns))
}

func (c *codegen) getFunction(name scanner.Token) exprValue {
	// Get function in this file
	mangledName := mangleName(name.Lexeme)

	for _, function := range c.functions {
		if function.Name() == mangledName || function.Name() == name.Lexeme {
			return exprValue{
				v: function,
			}
		}
	}

	// Resolve function from project
	function, filePath := c.resolver.GetFunction(name.Lexeme)

	if filePath == c.path {
		panic("codegen.getFunction() - Local function not found in functions array")
	}

	t := c.createFunctionType(function)

	value := c.module.Declare(t)
	c.functions = append(c.functions, value)

	return exprValue{
		v: value,
	}
}

func (c *codegen) beginBlock(block *llvm.Block) {
	c.block = block
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
			llvmType = c.module.Primitive("bool", 1, llvm.BooleanEncoding)

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
