package checker

import "fireball/core/ast"

type resetter struct {
	c *checker
}

func reset(c *checker, decls []ast.Decl) {
	r := &resetter{c: c}

	for _, decl := range decls {
		r.AcceptDecl(decl)
	}
}

// ast.Acceptor

func (r *resetter) AcceptDecl(decl ast.Decl) {
	decl.AcceptChildren(r)
}

func (r *resetter) AcceptStmt(stmt ast.Stmt) {
	stmt.AcceptChildren(r)

	switch stmt := stmt.(type) {
	case *ast.Variable:
		if stmt.InferType {
			stmt.Type = nil
		}
	}
}

func (r *resetter) AcceptExpr(expr ast.Expr) {
	expr.AcceptChildren(r)

	expr.Result().SetInvalid()
}
