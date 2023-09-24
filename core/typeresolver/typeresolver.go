package typeresolver

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
	"fireball/core/utils"
	"fmt"
)

type typeResolver struct {
	expr ast.Expr

	reporter utils.Reporter
	resolver utils.Resolver
}

func Resolve(reporter utils.Reporter, resolver utils.Resolver, decls []ast.Decl) {
	r := &typeResolver{
		reporter: reporter,
		resolver: resolver,
	}

	for _, decl := range decls {
		r.AcceptDecl(decl)
	}
}

// Declarations

func (r *typeResolver) visitImpl(decl *ast.Impl) {
	type_, _ := r.resolver.GetType(decl.Struct.Lexeme)

	if s, ok := type_.(*ast.Struct); ok {
		decl.Type_ = s
	} else {
		r.reporter.Report(utils.Diagnostic{
			Kind:    utils.ErrorKind,
			Range:   core.TokenToRange(decl.Struct),
			Message: fmt.Sprintf("Struct with the name '%s' does not exist.", decl.Struct),
		})

		decl.Type_ = nil
	}
}

// types.PtrVisitor

func (r *typeResolver) VisitType(type_ *types.Type) {
	if v, ok := (*type_).(*types.UnresolvedType); ok {
		if t, _ := r.resolver.GetType(v.Identifier.Lexeme); t != nil {
			*type_ = t.WithRange(v.Range())
		} else {
			r.reporter.Report(utils.Diagnostic{
				Kind:    utils.ErrorKind,
				Range:   v.Range(),
				Message: fmt.Sprintf("Unknown type '%s'.", v),
			})

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
	switch decl := decl.(type) {
	case *ast.Impl:
		r.visitImpl(decl)
	}

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
