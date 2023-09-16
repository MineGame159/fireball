package codegen

import (
	"fireball/core/ast"
	"fireball/core/llvm"
	"fireball/core/types"
)

func (c *codegen) VisitStruct(_ *ast.Struct) {
}

func (c *codegen) VisitEnum(_ *ast.Enum) {
}

func (c *codegen) VisitFunc(decl *ast.Func) {
	// Get function
	mangledName := mangleName(decl.Name.Lexeme)
	var function *llvm.Function

	for _, f := range c.functions {
		if f.Name() == mangledName {
			if fu, ok := f.(*llvm.Function); ok {
				function = fu
				break
			}
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

	// Copy parameters
	for i, param := range decl.Params {
		type_ := c.getType(param.Type)

		pointer := c.block.Alloca(type_)
		pointer.SetName(param.Name.Lexeme + ".var")

		c.block.Store(pointer, function.GetParameter(i))
		c.addVariable(param.Name, exprValue{v: pointer})
	}

	// Body
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
