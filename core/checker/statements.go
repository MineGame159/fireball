package checker

import (
	"fireball/core"
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

	// Check initializer value
	valueOk := true

	// TODO: Somehow make this code nicer lmao
	if stmt.Initializer != nil && stmt.Initializer.Result().Kind == ast.InvalidResultKind {
		valueOk = false
	} else {
		if stmt.Initializer != nil && stmt.Initializer.Result().Kind != ast.ValueResultKind {
			c.errorRange(stmt.Initializer.Range(), "Invalid value.")
			valueOk = false
		} else {
			if stmt.Type == nil {
				if stmt.Initializer == nil {
					c.errorToken(stmt.Name, "Variable with no initializer needs to have an explicit type.")
					valueOk = false
				} else {
					stmt.Type = stmt.Initializer.Result().Type
				}
			} else {
				if stmt.Initializer != nil && !stmt.Initializer.Result().Type.CanAssignTo(stmt.Type) {
					c.errorRange(stmt.Initializer.Range(), "Initializer with type '%s' cannot be assigned to a variable with type '%s'.", stmt.Initializer.Result().Type, stmt.Type)
				}
			}
		}
	}

	if !valueOk {
		stmt.Type = types.Primitive(types.Void, core.Range{})
	}

	// Check name collision
	if c.hasVariableInScope(stmt.Name) {
		c.errorToken(stmt.Name, "Variable with the name '%s' already exists in the current scope.", stmt.Name)
	} else {
		c.addVariable(stmt.Name, stmt.Type)
	}

	// Check void type
	if valueOk && types.IsPrimitive(stmt.Type, types.Void) {
		c.errorToken(stmt.Name, "Variable cannot be of type 'void'.")
	}
}

func (c *checker) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(c)

	// Check condition value
	if stmt.Condition.Result().Kind != ast.ValueResultKind {
		c.errorRange(stmt.Condition.Range(), "Invalid value.")
	} else {
		if !types.IsPrimitive(stmt.Condition.Result().Type, types.Bool) {
			c.errorRange(stmt.Condition.Range(), "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Result().Type)
		}
	}
}

func (c *checker) VisitFor(stmt *ast.For) {
	// Visit children
	c.loopDepth++
	stmt.AcceptChildren(c)
	c.loopDepth--

	// Check condition value
	if stmt.Condition.Result().Kind != ast.ValueResultKind {
		c.errorRange(stmt.Condition.Range(), "Invalid value.")
	} else {
		if stmt.Condition != nil && !types.IsPrimitive(stmt.Condition.Result().Type, types.Bool) {
			c.errorRange(stmt.Condition.Range(), "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Result().Type)
		}
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	// Check return value
	var type_ types.Type
	var range_ core.Range

	if stmt.Expr != nil {
		if stmt.Expr.Result().Kind != ast.ValueResultKind {
			c.errorRange(stmt.Expr.Range(), "Invalid value.")
			return
		}

		type_ = stmt.Expr.Result().Type
		range_ = stmt.Expr.Range()
	} else {
		type_ = types.Primitive(types.Void, core.Range{})
		range_ = core.TokenToRange(stmt.Token_)
	}

	if !type_.CanAssignTo(c.function.Returns) {
		c.errorRange(range_, "Cannot return type '%s' from a function with return type '%s'.", type_, c.function.Returns)
	}
}

func (c *checker) VisitBreak(stmt *ast.Break) {
	stmt.AcceptChildren(c)

	// Check if break is inside a loop
	if c.loopDepth == 0 {
		c.errorToken(stmt.Token(), "A 'break' statement needs to be inside a loop.")
	}
}

func (c *checker) VisitContinue(stmt *ast.Continue) {
	stmt.AcceptChildren(c)

	// Check if continue is inside a loop
	if c.loopDepth == 0 {
		c.errorToken(stmt.Token(), "A 'continue' statement needs to be inside a loop.")
	}
}
