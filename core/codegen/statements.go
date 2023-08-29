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
	c.writeFmt("%%%s = alloca %s\n", stmt.Name, c.getType(stmt.Type))

	// Initializer
	if stmt.Initializer != nil {
		val := c.acceptExpr(stmt.Initializer)
		c.writeFmt("store %s %s, ptr %%%s\n", c.getType(stmt.Initializer.Type()), val, stmt.Name)
	}
}

func (c *codegen) VisitIf(stmt *ast.If) {
	// Get basic block names
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
	c.writeRaw(then + ":\n")
	c.acceptStmt(stmt.Then)
	c.writeFmt("br label %%%s\n", end)

	// Else
	if stmt.Else != nil {
		c.writeRaw(else_ + ":\n")
		c.acceptStmt(stmt.Else)
		c.writeFmt("br label %%%s\n", end)
	}

	// End
	c.writeRaw(end + ":\n")
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
