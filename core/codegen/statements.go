package codegen

import (
	"fireball/core/ast"
)

func (c *codegen) VisitBlock(stmt *ast.Block) {
	c.pushScope()
	c.module.PushScope(stmt.Cst())

	for _, s := range stmt.Stmts {
		c.acceptStmt(s)
	}

	c.module.PopScope()
	c.popScope()
}

func (c *codegen) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *codegen) VisitVar(stmt *ast.Var) {
	// Variable
	pointer := c.allocas[stmt]
	c.addVariable(stmt.Name, pointer)

	// Initializer
	if stmt.Value != nil {
		initializer := c.loadExpr(stmt.Value)

		store := c.block.Store(pointer.v, initializer.v)
		store.SetAlign(stmt.ActualType.Align())
		store.SetLocation(stmt.Name.Cst())
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
	c.block.Br(condition.v, then, else_)

	// Then
	c.beginBlock(then)
	c.acceptStmt(stmt.Then)
	c.block.Br(nil, end, nil)

	// Else
	if stmt.Else != nil {
		c.beginBlock(else_)
		c.acceptStmt(stmt.Else)
		c.block.Br(nil, end, nil)
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

	c.block.Br(nil, c.loopStart, nil)

	// Condition
	c.beginBlock(c.loopStart)
	condition := c.acceptExpr(stmt.Condition)
	c.block.Br(condition.v, body, c.loopEnd)

	// Body
	c.beginBlock(body)
	c.acceptStmt(stmt.Body)
	c.block.Br(nil, c.loopStart, nil)

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
	c.pushScope()
	c.module.PushScope(stmt.Cst())

	c.acceptStmt(stmt.Initializer)
	c.block.Br(nil, c.loopStart, nil)

	// Condition
	c.beginBlock(c.loopStart)

	if stmt.Condition != nil {
		condition := c.loadExpr(stmt.Condition)
		c.block.Br(condition.v, body, c.loopEnd)
	}

	// Body and increment
	if c.loopStart != body {
		c.beginBlock(body)
	}

	c.acceptStmt(stmt.Body)
	c.acceptExpr(stmt.Increment)

	c.block.Br(nil, c.loopStart, nil)

	// End
	c.module.PopScope()
	c.popScope()
	c.beginBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = prevLoopStart
	c.loopEnd = prevLoopEnd
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	if stmt.Value == nil {
		// Void
		c.block.Ret(nil).SetLocation(stmt.Cst())
	} else {
		// Other
		value := c.loadExpr(stmt.Value)
		c.block.Ret(value.v).SetLocation(stmt.Cst())
	}
}

func (c *codegen) VisitBreak(stmt *ast.Break) {
	c.block.Br(nil, c.loopEnd, nil).SetLocation(stmt.Cst())
}

func (c *codegen) VisitContinue(stmt *ast.Continue) {
	c.block.Br(nil, c.loopStart, nil).SetLocation(stmt.Cst())
}
