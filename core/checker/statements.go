package checker

import (
	"fireball/core/ast"
)

func (c *checker) VisitBlock(stmt *ast.Block) {
	c.pushScope()
	stmt.AcceptChildren(c)
	c.popScope()
}

func (c *checker) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(c)
}

func (c *checker) VisitVar(stmt *ast.Var) {
	stmt.AcceptChildren(c)

	// Check initializer value
	valueOk := true

	// TODO: Somehow make this code nicer lmao
	if stmt.Value != nil && stmt.Value.Result().Kind == ast.InvalidResultKind {
		valueOk = false
	} else {
		if stmt.Value != nil && stmt.Value.Result().Kind != ast.ValueResultKind {
			c.error(stmt.Value, "Invalid value")
			valueOk = false
		} else {
			if stmt.ActualType == nil {
				if stmt.Value == nil {
					c.error(stmt.Name, "Variable with no initializer needs to have an explicit type")
					valueOk = false
				} else {
					stmt.ActualType = stmt.Value.Result().Type
				}
			} else {
				if stmt.Value != nil && !stmt.Value.Result().Type.CanAssignTo(stmt.Type) {
					c.error(stmt.Value, "Initializer with type '%s' cannot be assigned to a variable with type '%s'", stmt.Value.Result().Type, stmt.Type)
				}
			}
		}
	}

	if !valueOk {
		stmt.ActualType = &ast.Primitive{Kind: ast.Void}
	}

	// Check name collision
	if c.hasVariableInScope(stmt.Name) {
		c.error(stmt.Name, "Variable with the name '%s' already exists in the current scope", stmt.Name)
	} else {
		c.addVariable(stmt.Name, stmt.ActualType)
	}

	// Check void type
	if valueOk && ast.IsPrimitive(stmt.ActualType, ast.Void) {
		c.error(stmt.Name, "Variable cannot be of type 'void'")
	}
}

func (c *checker) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(c)

	// Check condition value
	if stmt.Condition.Result().Kind != ast.ValueResultKind {
		c.error(stmt.Condition, "Invalid value")
	} else {
		if !ast.IsPrimitive(stmt.Condition.Result().Type, ast.Bool) {
			c.error(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'", stmt.Condition.Result().Type)
		}
	}
}

func (c *checker) VisitFor(stmt *ast.For) {
	// Visit children
	c.pushScope()
	c.loopDepth++

	stmt.AcceptChildren(c)

	c.loopDepth--
	c.popScope()

	// Check condition value
	if stmt.Condition != nil && stmt.Condition.Result().Kind != ast.InvalidResultKind {
		if stmt.Condition.Result().Kind != ast.ValueResultKind {
			c.error(stmt.Condition, "Invalid value")
		} else {
			if stmt.Condition != nil && !ast.IsPrimitive(stmt.Condition.Result().Type, ast.Bool) {
				c.error(stmt.Condition, "Condition needs to be of type 'bool' but got '%s'", stmt.Condition.Result().Type)
			}
		}
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	// Check return value
	var type_ ast.Type
	var errorNode ast.Node

	if stmt.Value != nil {
		if stmt.Value.Result().Kind != ast.ValueResultKind {
			c.error(stmt.Value, "Invalid value")
			return
		}

		type_ = stmt.Value.Result().Type
		errorNode = stmt.Value
	} else {
		type_ = &ast.Primitive{Kind: ast.Void}
		errorNode = stmt
	}

	if !type_.CanAssignTo(c.function.Returns) {
		c.error(errorNode, "Cannot return type '%s' from a function with return type '%s'", type_, c.function.Returns)
	}
}

func (c *checker) VisitBreak(stmt *ast.Break) {
	stmt.AcceptChildren(c)

	// Check if break is inside a loop
	if c.loopDepth == 0 {
		c.error(stmt, "A 'break' statement needs to be inside a loop")
	}
}

func (c *checker) VisitContinue(stmt *ast.Continue) {
	stmt.AcceptChildren(c)

	// Check if continue is inside a loop
	if c.loopDepth == 0 {
		c.error(stmt, "A 'continue' statement needs to be inside a loop")
	}
}
