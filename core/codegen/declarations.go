package codegen

import (
	"fireball/core/ast"
	"fireball/core/llvm"
	"fireball/core/scanner"
)

func (c *codegen) VisitNamespace(_ *ast.Namespace) {
}

func (c *codegen) VisitUsing(_ *ast.Using) {
}

func (c *codegen) VisitStruct(_ *ast.Struct) {
}

func (c *codegen) VisitImpl(impl *ast.Impl) {
	for _, function := range impl.Methods {
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
		c.addVariable(&ast.Token{Token_: name}, exprValue{v: function.GetParameter(0)})
	}

	// Copy parameters
	for i, param := range decl.Params {
		index := i
		if decl.Method() != nil {
			index++
		}

		pointer := c.block.Alloca(c.getType(param.Type))
		pointer.SetName(param.Name.String() + ".var")
		pointer.SetAlign(param.Type.Align())

		store := c.block.Store(pointer, function.GetParameter(index))
		store.SetAlign(param.Type.Align())

		c.addVariable(param.Name, exprValue{v: pointer})
	}

	// Body
	c.findAllocas(decl)

	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	// Add return if needed
	if ast.IsPrimitive(decl.Returns, ast.Void) {
		c.block.Ret(nil)
	}

	// Reset state
	function.PopScope()
	c.popScope()

	c.block = nil
	c.function = nil
}
