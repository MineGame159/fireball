package checker

import (
	"fireball/core/ast"
)

func (c *checker) VisitFunc(decl *ast.Func) {
	c.pushScope()

	for _, param := range decl.Params {
		c.addVariable(param.Name, param.Type).param = true
	}

	for _, stmt := range decl.Body {
		c.acceptStmt(stmt)
	}

	c.popScope()
}
