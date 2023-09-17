package codegen

import (
	"fireball/core/ast"
	"fireball/core/llvm"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (c *codegen) VisitStruct(_ *ast.Struct) {
}

func (c *codegen) VisitImpl(impl *ast.Impl) {
	for _, function := range impl.Functions {
		c.acceptDecl(function)
	}
}

func (c *codegen) VisitEnum(_ *ast.Enum) {
}

func (c *codegen) VisitFunc(decl *ast.Func) {
	// Get function
	var function *llvm.Function

	if f, ok := c.functions[decl]; ok {
		if fu, ok := f.(*llvm.Function); ok {
			function = fu
		}
	}

	if function == nil {
		return
	}

	// Setup state
	c.function = function
	c.beginBlock(function.Block("entry"))

	c.pushScope()
	function.PushScope()

	// Add this variable
	if struct_ := decl.Method(); struct_ != nil {
		name := scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}
		c.addVariable(name, exprValue{v: function.GetParameter(0)})
	}

	// Copy parameters
	for i, param := range decl.Params {
		index := i
		if decl.Method() != nil {
			index++
		}

		pointer := c.block.Alloca(c.getType(param.Type))
		pointer.SetName(param.Name.Lexeme + ".var")

		c.block.Store(pointer, function.GetParameter(index))
		c.addVariable(param.Name, exprValue{v: pointer})
	}

	// Body
	c.findAllocas(decl)

	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	// Add return if needed
	if types.IsPrimitive(decl.Returns, types.Void) {
		c.block.Ret(nil)
	}

	// Reset state
	function.PopScope()
	c.popScope()

	c.block = nil
	c.function = nil
}
