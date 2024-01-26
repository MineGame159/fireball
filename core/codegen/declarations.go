package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
	"fireball/core/scanner"
)

func (c *codegen) VisitNamespace(_ *ast.Namespace) {}

func (c *codegen) VisitUsing(_ *ast.Using) {}

func (c *codegen) VisitStruct(_ *ast.Struct) {}

func (c *codegen) VisitImpl(impl *ast.Impl) {
	for _, function := range impl.Methods {
		c.acceptDecl(function)
	}
}

func (c *codegen) VisitEnum(_ *ast.Enum) {}

func (c *codegen) VisitInterface(_ *ast.Interface) {}

func (c *codegen) VisitFunc(decl *ast.Func) {
	if !decl.HasBody() {
		return
	}

	// Get function
	function := c.functions[decl]

	// Setup state
	c.astFunction = decl
	c.function = function
	c.beginBlock(function.Block("entry"))

	c.scopes.push(function.Meta())

	// Add this variable
	if struct_ := decl.Method(); struct_ != nil {
		name := scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}
		type_ := ast.Pointer{Pointee: struct_}

		c.scopes.addVariable(&ast.Token{Token_: name}, &type_, exprValue{v: function.Typ.Params[0]}, 1)
	}

	// Copy parameters
	for i, param := range decl.Params {
		index := i
		if decl.Method() != nil {
			index++
		}

		pointer := c.alloca(param.Type, param.Name.String()+".var", param)

		c.block.Add(&ir.StoreInst{
			Pointer: pointer,
			Value:   function.Typ.Params[index],
			Align:   param.Type.Align() * 8,
		})

		c.scopes.addVariable(param.Name, param.Type, exprValue{v: pointer}, uint32(index+1))
	}

	// Body
	c.findAllocas(decl)

	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	// Add return if needed
	if ast.IsPrimitive(decl.Returns, ast.Void) {
		c.block.Add(&ir.RetInst{})
	}

	// Reset state
	c.scopes.pop()

	c.block = nil
	c.function = nil
	c.astFunction = nil
}

func (c *codegen) VisitGlobalVar(_ *ast.GlobalVar) {}
