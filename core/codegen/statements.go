package codegen

import (
	"fireball/core/ast"
)

func (c *codegen) VisitBlock(stmt *ast.Block) {
	c.pushScope()

	for _, s := range stmt.Stmts {
		c.acceptStmt(s)
	}

	c.popScope()
}

func (c *codegen) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *codegen) VisitVariable(stmt *ast.Variable) {
	// Variable
	c.addVariable(stmt.Name, c.locals.named(stmt.Name.Lexeme, stmt.Type))
	c.writeFmt("%%_%s = alloca %s\n", stmt.Name, c.getType(stmt.Type))

	// Initializer
	if stmt.Initializer != nil {
		val := c.load(c.acceptExpr(stmt.Initializer), stmt.Initializer.Type())
		c.writeFmt("store %s %s, ptr %%_%s\n", c.getType(stmt.Initializer.Type()), val, stmt.Name)
	}
}

func (c *codegen) VisitIf(stmt *ast.If) {
	// GetLeaf basic block names
	then := c.blocks.unnamedRaw()
	else_ := ""
	end := ""

	if stmt.Else != nil {
		else_ = c.blocks.unnamedRaw()
		end = c.blocks.unnamedRaw()
	} else {
		else_ = c.blocks.unnamedRaw()
		end = else_
	}

	// Condition
	condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Type())
	c.writeFmt("br i1 %s, label %%%s, label %%%s\n", condition, then, else_)

	// Then
	c.writeBlock(then)
	c.acceptStmt(stmt.Then)
	c.writeFmt("br label %%%s\n", end)

	// Else
	if stmt.Else != nil {
		c.writeBlock(else_)
		c.acceptStmt(stmt.Else)
		c.writeFmt("br label %%%s\n", end)
	}

	// End
	c.writeBlock(end)
}

func (c *codegen) VisitFor(stmt *ast.For) {
	// GetLeaf basic block names
	c.loopStart = c.blocks.unnamedRaw()
	body := c.loopStart
	c.loopEnd = ""

	c.writeFmt("br label %%%s\n", c.loopStart)

	// Condition
	c.writeBlock(c.loopStart)

	if stmt.Condition != nil {
		body = c.blocks.unnamedRaw()
		c.loopEnd = c.blocks.unnamedRaw()

		condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Type())
		c.writeFmt("br i1 %s, label %%%s, label %%%s\n", condition, body, c.loopEnd)
	} else {
		c.loopEnd = c.blocks.unnamedRaw()
	}

	// Body
	if c.loopStart != body {
		c.writeBlock(body)
	}

	c.acceptStmt(stmt.Body)
	c.writeFmt("br label %%%s\n", c.loopStart)

	// End
	c.writeBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = ""
	c.loopEnd = ""
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	if stmt.Expr == nil {
		// Void
		c.writeStr("ret void\n")
	} else {
		// Other
		val := c.load(c.acceptExpr(stmt.Expr), stmt.Expr.Type())
		c.writeFmt("ret %s %s\n", c.getType(stmt.Expr.Type()), val)
	}
}

func (c *codegen) VisitBreak(stmt *ast.Break) {
	c.writeFmt("br label %%%s\n", c.loopEnd)
}

func (c *codegen) VisitContinue(stmt *ast.Continue) {
	c.writeFmt("br label %%%s\n", c.loopStart)
}
