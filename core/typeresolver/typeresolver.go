package typeresolver

import (
	"fireball/core/ast"
	"fireball/core/utils"
	"fmt"
)

type typeResolver struct {
	expr ast.Expr

	reporter utils.Reporter
	resolver ast.Resolver
}

func Resolve(reporter utils.Reporter, resolver ast.Resolver, node ast.Node) {
	r := typeResolver{
		reporter: reporter,
		resolver: resolver,
	}

	r.VisitNode(node)
}

// Declarations

func (t *typeResolver) visitImpl(decl *ast.Impl) {
	if decl.Struct != nil {
		type_ := t.resolver.GetType(decl.Struct.Token().Lexeme)

		if s, ok := type_.(*ast.Struct); ok {
			decl.Type = s
		} else {
			t.reporter.Report(utils.Diagnostic{
				Kind:    utils.ErrorKind,
				Range:   decl.Struct.Cst().Range,
				Message: fmt.Sprintf("Struct with the name '%s' does not exist", decl.Struct.Token()),
			})

			decl.Type = nil
		}
	}

	decl.AcceptChildren(t)
}

// Types

func (t *typeResolver) visitType(type_ ast.Type) {
	if resolvable, ok := type_.(*ast.Resolvable); ok {
		if resolved := t.resolver.GetType(resolvable.Token().Lexeme); resolved != nil {
			resolvable.Type = resolved
		} else {
			t.reporter.Report(utils.Diagnostic{
				Kind:    utils.ErrorKind,
				Range:   resolvable.Cst().Range,
				Message: fmt.Sprintf("Unknown type '%s'", resolvable.Token()),
			})

			resolvable.Type = &ast.Primitive{Kind: ast.Void}

			if t.expr != nil {
				t.expr.Result().SetInvalid()
			}
		}
	}

	type_.AcceptChildren(t)
}

// ast.Visitor

func (t *typeResolver) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Impl:
		t.visitImpl(node)

	case ast.Expr:
		prevTypeExpr := t.expr
		t.expr = node

		node.AcceptChildren(t)

		t.expr = prevTypeExpr

	case ast.Type:
		t.visitType(node)

	default:
		node.AcceptChildren(t)
	}
}
