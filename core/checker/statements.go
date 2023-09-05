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

	// Check initializer type
	if stmt.Type == nil {
		if stmt.Initializer == nil {
			c.errorToken(stmt.Name, "Variable with no initializer needs to have an explicit type.")
			stmt.Type = types.Primitive(types.Void, core.Range{})
		} else {
			stmt.Type = stmt.Initializer.Type().WithoutRange()
		}
	} else {
		if stmt.Initializer != nil && !stmt.Initializer.Type().CanAssignTo(stmt.Type) {
			c.errorRange(stmt.Initializer.Range(), "Initializer with type '%s' cannot be assigned to a variable with type '%s'.", stmt.Initializer.Type(), stmt.Type)
		}
	}

	// Check name collision
	if c.hasVariableInScope(stmt.Name) {
		c.errorToken(stmt.Name, "Variable with the name '%s' already exists in the current scope.", stmt.Name)
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
		c.errorRange(stmt.Condition.Range(), "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitFor(stmt *ast.For) {
	// Visit children
	c.loopDepth++
	stmt.AcceptChildren(c)
	c.loopDepth--

	// Check condition type
	if stmt.Condition != nil && !types.IsPrimitive(stmt.Condition.Type(), types.Bool) {
		c.errorRange(stmt.Condition.Range(), "Condition needs to be of type 'bool' but got '%s'.", stmt.Condition.Type())
	}
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	// Check return type
	var type_ types.Type
	var range_ core.Range

	if stmt.Expr != nil {
		type_ = stmt.Expr.Type().WithoutRange()
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
