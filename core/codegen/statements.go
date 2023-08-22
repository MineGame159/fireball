package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *codegen) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *codegen) VisitVariable(stmt *ast.Variable) {
	// Variable
	c.addVariable(stmt.Name, c.locals.named(stmt.Name.Lexeme, types.PointerType{Pointee: stmt.Type}))
	c.writeFmt("%%%s = alloca %s\n", stmt.Name, c.getType(stmt.Type))

	// Initializer
	if stmt.Initializer != nil {
		val := c.acceptExpr(stmt.Initializer)
		c.writeFmt("store %s %s, ptr %%%s\n", c.getType(val.type_), val, stmt.Name)
	}
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	if stmt.Expr == nil {
		// Void
		c.writeStr("ret void\n")
	} else {
		// Other
		val := c.load(c.acceptExpr(stmt.Expr))
		c.writeFmt("ret %s %s\n", c.getType(stmt.Expr.Type()), val)
	}
}
