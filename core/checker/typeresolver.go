package checker

import (
	"fireball/core/ast"
	"fireball/core/types"
)

type typeResolver struct {
	checker *checker
	expr    ast.Expr
}

func resolveTypes(checker *checker, decls []ast.Decl) {
	r := &typeResolver{
		checker: checker,
	}

	for _, decl := range decls {
		r.AcceptDecl(decl)
	}
}

// types.PtrVisitor

func (r *typeResolver) VisitType(type_ *types.Type) {
	if v, ok := (*type_).(*types.UnresolvedType); ok {
		if t, _ := r.checker.resolver.GetType(v.Identifier.Lexeme); t != nil {
			*type_ = t.WithRange(v.Range())
		} else {
			r.checker.errorRange(v.Range(), "Unknown type '%s'.", v)
			*type_ = types.Primitive(types.Void, v.Range())

			if r.expr != nil {
				r.expr.Result().SetInvalid()
			}
		}
	}

	if *type_ != nil {
		(*type_).AcceptTypesPtr(r)
	}
}

// ast.Acceptor

func (r *typeResolver) AcceptDecl(decl ast.Decl) {
	decl.AcceptChildren(r)
	decl.AcceptTypesPtr(r)
}

func (r *typeResolver) AcceptStmt(stmt ast.Stmt) {
	stmt.AcceptChildren(r)
	stmt.AcceptTypesPtr(r)
}

func (r *typeResolver) AcceptExpr(expr ast.Expr) {
	expr.AcceptChildren(r)

	prevTypeExpr := r.expr
	r.expr = expr

	expr.AcceptTypesPtr(r)

	r.expr = prevTypeExpr
}
