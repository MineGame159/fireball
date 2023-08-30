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

	// Check initializer type
	if stmt.Type == nil {
		if stmt.Initializer == nil {
			c.errorNode(stmt, "Variable with no initializer needs to have an explicit type.")
		} else {
			stmt.Type = stmt.Initializer.Type()
		}
	} else {
		if stmt.Initializer != nil && !stmt.Initializer.Type().CanAssignTo(stmt.Type) {
			c.errorNode(stmt, "Initializer with type '%s' cannot be assigned to a variable with type '%s'.", stmt.Initializer.Type(), stmt.Type)
		}
	}

	// Check name collision
	if var_ := c.getVariable(stmt.Name); var_ != nil {
		c.errorNode(stmt, "Variable with the name '%s' already exists.", stmt.Name)
	} else {
		c.addVariable(stmt.Name, stmt.Type)
	}

	// Check void type
	if types.IsPrimitive(stmt.Type, types.Void) {
		c.errorToken(stmt.Name, "Variable cannot be of type 'void'.")
	}
}

func (c *checker) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(c)

	// Check condition type
	if !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.errorNode(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitFor(stmt *ast.For) {
	// Visit children
	c.loopDepth++
	stmt.AcceptChildren(c)
	c.loopDepth--

	// Check condition type
	if stmt.Condition != nil && !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.errorNode(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	// Check return type
	var type_ types.Type = types.Primitive(types.Void)

	if stmt.Expr != nil {
		type_ = stmt.Expr.Type()
	}

	if !type_.CanAssignTo(c.function.Returns) {
		c.errorNode(stmt, "Cannot return type '%s' from a function with return type '%s'.", type_, c.function.Returns)
	}
}

func (c *checker) VisitBreak(stmt *ast.Break) {
	stmt.AcceptChildren(c)

	// Check if break is inside a loop
	if c.loopDepth == 0 {
		c.errorNode(stmt, "A 'break' statement needs to be inside a loop.")
	}
}

func (c *checker) VisitContinue(stmt *ast.Continue) {
	stmt.AcceptChildren(c)

	// Check if continue is inside a loop
	if c.loopDepth == 0 {
		c.errorNode(stmt, "A 'continue' statement needs to be inside a loop.")
	}
}
