package checker

import "fireball/core/ast"

type resetter struct {
	c *checker
}

func reset(c *checker, node ast.Node) {
	r := resetter{c: c}
	r.VisitNode(node)
}

// ast.Visitor

func (r *resetter) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Var:
		node.ActualType = nil

	case ast.Expr:
		node.AcceptChildren(r)
		node.Result().SetInvalid()

	default:
		node.AcceptChildren(r)
	}
}
