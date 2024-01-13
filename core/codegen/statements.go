package codegen

import (
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
	pointer := c.allocas[stmt]
	c.scopes.addVariable(stmt.Name, stmt.ActualType, pointer, 0)

	// Initializer
	if stmt.Value != nil {
		initializer := c.loadExpr(stmt.Value)

		store := c.block.Add(&ir.StoreInst{
			Pointer: pointer.v,
			Value:   initializer.v,
			Align:   stmt.ActualType.Align() * 8,
		})

		c.setLocationMeta(store, stmt)
	}
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
	condition := c.loadExpr(stmt.Condition)
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
	prevLoopStart := c.loopStart
	prevLoopEnd := c.loopEnd

	c.loopStart = c.function.Block("while.start")
	body := c.function.Block("while.body")
	c.loopEnd = c.function.Block("while.end")

	c.block.Add(&ir.BrInst{True: c.loopStart})

	// Condition
	c.beginBlock(c.loopStart)
	condition := c.acceptExpr(stmt.Condition)
	c.block.Add(&ir.BrInst{Condition: condition.v, True: body, False: c.loopEnd})

	// Body
	c.beginBlock(body)
	if c.acceptStmt(stmt.Body) {
		c.block.Add(&ir.BrInst{True: c.loopStart})
	}

	// End
	c.beginBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = prevLoopStart
	c.loopEnd = prevLoopEnd
}

func (c *codegen) VisitFor(stmt *ast.For) {
	// Get blocks
	prevLoopStart := c.loopStart
	prevLoopEnd := c.loopEnd

	c.loopStart = c.function.Block("for.start")
	c.loopEnd = c.function.Block("for.end")
	body := c.loopStart

	if stmt.Condition != nil {
		body = c.function.Block("for.body")
	}

	// Initializer
	c.scopes.pushBlock(stmt)

	if c.acceptStmt(stmt.Initializer) {
		c.block.Add(&ir.BrInst{True: c.loopStart})
	}

	// Condition
	c.beginBlock(c.loopStart)

	if stmt.Condition != nil {
		condition := c.loadExpr(stmt.Condition)
		c.block.Add(&ir.BrInst{Condition: condition.v, True: body, False: c.loopEnd})
	}

	// Body and increment
	if c.loopStart != body {
		c.beginBlock(body)
	}

	c.acceptStmt(stmt.Body)
	c.acceptExpr(stmt.Increment)

	c.block.Add(&ir.BrInst{True: c.loopStart})

	// End
	c.scopes.pop()
	c.beginBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = prevLoopStart
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
		value := c.loadExpr(stmt.Value)

		c.setLocationMeta(
			c.block.Add(&ir.RetInst{Value: value.v}),
			stmt,
		)
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
		c.block.Add(&ir.BrInst{True: c.loopStart}),
		stmt,
	)

	c.block = nil
}
