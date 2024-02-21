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

func (c *codegen) VisitImpl(decl *ast.Impl) {
	struct_, _ := ast.As[*ast.Struct](decl.Type)

	if len(struct_.GenericParams) > 0 {
		prevResolver := c.resolver
		c.resolver = ast.NewGenericResolver(c.resolver, struct_.GenericParams)

		for _, spec := range struct_.Specializations {
			for _, method := range spec.Methods {
				var s specializer
				s.prepare(method.Underlying(), struct_.GenericParams)

				s.specialize(spec.Types)
				c.visitFunc(method)

				s.finish()
			}
		}

		for _, method := range decl.Methods {
			if method.IsStatic() {
				c.VisitFunc(method)
			}
		}

		c.resolver = prevResolver
	} else {
		for _, function := range decl.Methods {
			c.VisitFunc(function)
		}
	}
}

func (c *codegen) VisitEnum(_ *ast.Enum) {}

func (c *codegen) VisitInterface(_ *ast.Interface) {}

func (c *codegen) VisitFunc(decl *ast.Func) {
	if !decl.HasBody() {
		return
	}

	c.visitFunc(decl)
}

func (c *codegen) visitFunc(f ast.FuncType) {
	decl := f.Underlying()

	if len(decl.Generics()) == 0 {
		c.genFunc(f)
	} else {
		var s specializer
		s.prepare(decl, decl.Generics())

		for _, spec := range decl.Specializations {
			s.specialize(spec.Types)
			c.genFunc(spec)
		}

		s.finish()
	}
}

func (c *codegen) genFunc(f ast.FuncType) {
	decl := f.Underlying()

	// Get function
	function := c.functions[f]

	// Setup state
	c.astFunction = f
	c.function = function
	c.beginBlock(function.Block("entry"))

	c.scopes.push(function.Meta())
	c.allocas.reset()

	prevResolver := c.resolver
	if len(f.Underlying().GenericParams) != 0 {
		c.resolver = ast.NewGenericResolver(c.resolver, f.Underlying().GenericParams)
	}

	// Add this variable
	if receiver := f.Receiver(); receiver != nil {
		name := scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}
		node := cst.Node{Kind: cst.TokenNode, Token: name, Range: decl.Name.Cst().Range}

		c.scopes.addVariable(ast.NewToken(node, name), receiver, function.Typ.Params[0], 1)
	}

	// Copy parameters
	funcAbi := abi.GetFuncAbi(decl)
	returnArgs := funcAbi.Classify(f.Returns(), nil)

	index := 0
	if len(returnArgs) == 1 && returnArgs[0].Class == abi.Memory {
		index++
	}
	if decl.Receiver() != nil {
		index++
	}

	paramI := index

	for i := 0; i < f.ParameterCount(); i++ {
		param := f.ParameterIndex(i)

		pointer := c.allocas.get(param.Type, param.Param.Name.String()+".var")
		c.setLocationMeta(pointer, param.Param)

		args := funcAbi.Classify(param.Type, nil)
		params := function.Typ.Params[index : index+len(args)]
		index += len(args)

		c.paramsToVariable(args, params, pointer, param.Type)

		paramI++
		c.scopes.addVariable(param.Param.Name, param.Type, pointer, uint32(paramI))
	}

	// Body
	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	// Add return if needed
	if ast.IsPrimitive(f.Returns(), ast.Void) {
		c.block.Add(&ir.RetInst{})
	}

	// Reset state
	c.resolver = prevResolver

	c.scopes.pop()

	c.block = nil
	c.function = nil
	c.astFunction = nil
}

func (c *codegen) VisitGlobalVar(_ *ast.GlobalVar) {}
