package checker

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *checker) VisitBlock(stmt *ast.Block) {
	c.pushScope()

	for _, s := range stmt.Stmts {
		c.acceptStmt(s)
	}

	c.popScope()
}

func (c *checker) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *checker) VisitVariable(stmt *ast.Variable) {
	c.acceptExpr(stmt.Initializer)

	if stmt.Type == nil {
		if stmt.Initializer == nil {
			c.error(stmt, "Variable with no initializer needs to have an explicit type.")
		} else {
			stmt.Type = stmt.Initializer.Type()
		}
	} else {
		if stmt.Initializer != nil && !stmt.Initializer.Type().CanAssignTo(stmt.Type) {
			c.error(stmt, "Initializer with type '%s' cannot be assigned to a variable with type '%s'.", stmt.Initializer.Type(), stmt.Type)
		}
	}

	if var_ := c.getVariable(stmt.Name); var_ != nil {
		c.error(stmt, "Variable with the name '%s' already exists.", stmt.Name)
	} else {
		c.addVariable(stmt.Name, stmt.Type)
	}
}

func (c *checker) VisitIf(stmt *ast.If) {
	c.acceptExpr(stmt.Condition)
	c.acceptStmt(stmt.Then)
	c.acceptStmt(stmt.Else)

	if !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.error(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	c.acceptExpr(stmt.Expr)
}
