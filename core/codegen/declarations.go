package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/cst"
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
	c.allocas.reset()

	// Add this variable
	if struct_ := decl.Method(); struct_ != nil {
		name := scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}
		node := cst.Node{Kind: cst.IdentifierNode, Token: name, Range: decl.Name.Cst().Range}

		c.scopes.addVariable(ast.NewToken(node, name), struct_, function.Typ.Params[0], 1)
	}

	// Copy parameters
	funcAbi := abi.GetFuncAbi(decl)
	returnArgs := funcAbi.Classify(decl.Returns, nil)

	index := 0
	if len(returnArgs) == 1 && returnArgs[0].Class == abi.Memory {
		index++
	}
	if decl.Method() != nil {
		index++
	}

	paramI := index

	for _, param := range decl.Params {
		pointer := c.allocas.get(param.Type, param.Name.String()+".var")
		c.setLocationMeta(pointer, param)

		args := funcAbi.Classify(param.Type, nil)
		params := function.Typ.Params[index : index+len(args)]
		index += len(args)

		c.paramsToVariable(args, params, pointer, param.Type)

		paramI++
		c.scopes.addVariable(param.Name, param.Type, pointer, uint32(paramI))
	}

	// Body
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
