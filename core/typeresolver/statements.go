package typeresolver

import (
	"fireball/core/ast"
)

func (r *resolver) VisitBlock(stmt *ast.Block) {
	stmt.AcceptChildren(r)
}

func (r *resolver) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(r)
}

func (r *resolver) VisitVariable(stmt *ast.Variable) {
	stmt.AcceptChildren(r)
}

func (r *resolver) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(r)
}

func (r *resolver) VisitFor(stmt *ast.For) {
	stmt.AcceptChildren(r)
}

func (r *resolver) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(r)
}
