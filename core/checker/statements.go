package checker

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *checker) VisitBlock(stmt *ast.Block) {
	c.pushScope()
	stmt.AcceptChildren(c)
	c.popScope()
}

func (c *checker) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(c)
}

func (c *checker) VisitVariable(stmt *ast.Variable) {
	stmt.AcceptChildren(c)

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
	stmt.AcceptChildren(c)

	if !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.error(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitFor(stmt *ast.For) {
	stmt.AcceptChildren(c)

	if stmt.Condition != nil && !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.error(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	var type_ types.Type = types.Primitive(types.Void)

	if stmt.Expr != nil {
		type_ = stmt.Expr.Type()
	}

	if !type_.CanAssignTo(c.function.Returns) {
		c.error(stmt, "Cannot return type '%s' from a function with return type '%s'.", type_, c.function.Returns)
	}
}
