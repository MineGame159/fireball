package checker

import (
	"fireball/core/ast"
)

func (c *checker) VisitFunc(decl *ast.Func) {
	// Push scope
	c.function = decl
	c.pushScope()

	// Params
	for _, param := range decl.Params {
		c.addVariable(param.Name, param.Type).param = true
	}

	// Body
	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	// Pop scope
	c.function = nil
	c.popScope()
}
