package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/ir"
)

func (c *codegen) VisitBlock(stmt *ast.Block) {
	c.scopes.pushBlock(stmt)

	for _, s := range stmt.Stmts {
		c.acceptStmt(s)
	}

	c.scopes.pop()
}

func (c *codegen) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *codegen) VisitVar(stmt *ast.Var) {
	// Variable
	pointer := c.allocas.get(stmt.ActualType, stmt.Name.String()+".var")
	c.setLocationMeta(pointer, stmt)

	c.scopes.addVariable(stmt.Name, stmt.ActualType, pointer, 0)

	// Initializer
	var initializer ir.Value

	if stmt.Value != nil {
		initializer = c.implicitCastLoadExpr(stmt.ActualType, stmt.Value).v
	} else {
		initializer = &ir.ZeroInitConst{Typ: c.types.get(stmt.ActualType)}
	}

	store := c.block.Add(&ir.StoreInst{
		Pointer: pointer,
		Value:   initializer,
		Align:   abi.GetTargetAbi().Align(stmt.ActualType),
	})

	c.setLocationMeta(store, stmt)
}

func (c *codegen) VisitIf(stmt *ast.If) {
	// Get blocks
	then := c.function.Block("if.then")
	end := c.function.Block("if.end")
	else_ := end

	if stmt.Else != nil {
		else_ = c.function.Block("if.else")
	}

	// Condition
	required := ast.Primitive{Kind: ast.Bool}
	condition := c.implicitCastLoadExpr(&required, stmt.Condition)

	c.block.Add(&ir.BrInst{Condition: condition.v, True: then, False: else_})

	// Then
	c.beginBlock(then)
	if c.acceptStmt(stmt.Then) {
		c.block.Add(&ir.BrInst{True: end})
	}

	// Else
	if stmt.Else != nil {
		c.beginBlock(else_)
		if c.acceptStmt(stmt.Else) {
			c.block.Add(&ir.BrInst{True: end})
		}
	}

	// End
	c.beginBlock(end)
}

func (c *codegen) VisitWhile(stmt *ast.While) {
	// Get blocks
	prevLoopSkip := c.loopSkip
	prevLoopEnd := c.loopEnd

	c.loopSkip = c.function.Block("while.condition")
	body := c.function.Block("while.body")
	c.loopEnd = c.function.Block("while.end")

	c.block.Add(&ir.BrInst{True: c.loopSkip})

	// Condition
	c.beginBlock(c.loopSkip)

	required := ast.Primitive{Kind: ast.Bool}
	condition := c.implicitCastLoadExpr(&required, stmt.Condition)

	c.block.Add(&ir.BrInst{Condition: condition.v, True: body, False: c.loopEnd})

	// Body
	c.beginBlock(body)

	if c.acceptStmt(stmt.Body) {
		c.block.Add(&ir.BrInst{True: c.loopSkip})
	}

	// End
	c.beginBlock(c.loopEnd)

	// Reset basic block names
	c.loopSkip = prevLoopSkip
	c.loopEnd = prevLoopEnd
}

func (c *codegen) VisitFor(stmt *ast.For) {
	// Get blocks
	prevLoopSkip := c.loopSkip
	prevLoopEnd := c.loopEnd

	condition := c.function.Block("for.condition")
	body := c.function.Block("for.body")
	c.loopSkip = c.function.Block("for.increment")
	c.loopEnd = c.function.Block("for.end")

	// Initializer
	c.scopes.pushBlock(stmt)

	if c.acceptStmt(stmt.Initializer) {
		c.block.Add(&ir.BrInst{True: condition})
	}

	// Condition
	c.beginBlock(condition)

	if stmt.Condition != nil {
		required := ast.Primitive{Kind: ast.Bool}
		cond := c.implicitCastLoadExpr(&required, stmt.Condition)

		c.block.Add(&ir.BrInst{Condition: cond.v, True: body, False: c.loopEnd})
	} else {
		c.block.Add(&ir.BrInst{True: body})
	}

	// Body
	c.beginBlock(body)

	if c.acceptStmt(stmt.Body) {
		c.block.Add(&ir.BrInst{True: c.loopSkip})
	}

	// Increment
	c.beginBlock(c.loopSkip)

	c.acceptExpr(stmt.Increment)
	c.block.Add(&ir.BrInst{True: condition})

	// End
	c.scopes.pop()
	c.beginBlock(c.loopEnd)

	// Reset basic block names
	c.loopSkip = prevLoopSkip
	c.loopEnd = prevLoopEnd
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	if stmt.Value == nil {
		// Void
		c.setLocationMeta(
			c.block.Add(&ir.RetInst{}),
			stmt,
		)
	} else {
		// Other
		funcAbi := abi.GetFuncAbi(c.astFunction.Underlying())

		value := c.implicitCastLoadExpr(c.astFunction.Returns(), stmt.Value)
		value = c.valueToReturnValue(funcAbi, value, c.astFunction.Returns(), c.function.Typ.Params)

		if value.v == nil {
			c.setLocationMeta(
				c.block.Add(&ir.RetInst{}),
				stmt,
			)
		} else {
			c.setLocationMeta(
				c.block.Add(&ir.RetInst{Value: value.v}),
				stmt,
			)
		}
	}

	c.block = nil
}

func (c *codegen) VisitBreak(stmt *ast.Break) {
	c.setLocationMeta(
		c.block.Add(&ir.BrInst{True: c.loopEnd}),
		stmt,
	)

	c.block = nil
}

func (c *codegen) VisitContinue(stmt *ast.Continue) {
	c.setLocationMeta(
		c.block.Add(&ir.BrInst{True: c.loopSkip}),
		stmt,
	)

	c.block = nil
}
