package typeresolver

import (
	"fireball/core/ast"
)

func (r *resolver) VisitGroup(expr *ast.Group) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitIdentifier(expr *ast.Identifier) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(r)
}

func (r *resolver) VisitMember(expr *ast.Member) {
	expr.AcceptChildren(r)
}
