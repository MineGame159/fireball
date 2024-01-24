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

	if stmt.Name == nil {
		return
	}

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
			if stmt.Type != nil {
				stmt.ActualType = stmt.Type
			}

			if stmt.ActualType == nil {
				if stmt.Value == nil {
					c.error(stmt.Name, "Variable with no initializer needs to have an explicit type")
					valueOk = false
				} else {
					stmt.ActualType = stmt.Value.Result().Type
				}
			} else {
				c.checkRequired(stmt.ActualType, stmt.Value)
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
		c.addVariable(stmt.Name, stmt.ActualType, stmt)
	}

	// Check void type
	if valueOk && ast.IsPrimitive(stmt.ActualType, ast.Void) {
		c.error(stmt.Name, "Variable cannot be of type 'void'")
	}
}

func (c *checker) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(c)

	required := ast.Primitive{Kind: ast.Bool}
	c.checkRequired(&required, stmt.Condition)
}

func (c *checker) VisitWhile(stmt *ast.While) {
	c.loopDepth++
	stmt.AcceptChildren(c)
	c.loopDepth--

	// Check condition value
	required := ast.Primitive{Kind: ast.Bool}
	c.checkRequired(&required, stmt.Condition)
}

func (c *checker) VisitFor(stmt *ast.For) {
	// Visit children
	c.pushScope()
	c.loopDepth++

	stmt.AcceptChildren(c)

	c.loopDepth--
	c.popScope()

	// Check condition value
	required := ast.Primitive{Kind: ast.Bool}
	c.checkRequired(&required, stmt.Condition)
}

func (c *checker) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(c)

	// Check return value
	if stmt.Value != nil {
		if stmt.Value.Result().Kind != ast.ValueResultKind {
			c.error(stmt.Value, "Invalid value")
			return
		}

		c.checkRequired(c.function.Returns, stmt.Value)
	} else {
		type_ := ast.Primitive{Kind: ast.Void}

		if !c.function.Returns.Equals(&type_) {
			c.error(stmt, "Expected a '%s' but got a 'void'", ast.PrintType(c.function.Returns))
		}
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
