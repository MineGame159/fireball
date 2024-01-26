package typeresolver

import (
	"fireball/core/ast"
	"fireball/core/utils"
	"fmt"
	"strings"
)

type typeResolver struct {
	expr ast.Expr

	reporter utils.Reporter
	resolver *ast.CombinedResolver
}

func Resolve(reporter utils.Reporter, root ast.RootResolver, file *ast.File) {
	r := typeResolver{
		reporter: reporter,
		resolver: ast.NewCombinedResolver(root),
	}

	for _, decl := range file.Decls {
		if using, ok := decl.(*ast.Using); ok {
			if resolver := root.GetResolver(using.Name); resolver != nil {
				r.resolver.Add(resolver)
			}
		}
	}

	r.VisitNode(file)
}

// Declarations

func (t *typeResolver) visitImpl(decl *ast.Impl) {
	// Struct
	if decl.Struct != nil {
		type_ := t.resolver.GetType(decl.Struct.Token().Lexeme)

		if s, ok := type_.(*ast.Struct); ok {
			decl.Type = s
		} else {
			t.reporter.Report(utils.Diagnostic{
				Kind:    utils.ErrorKind,
				Range:   decl.Struct.Cst().Range,
				Message: fmt.Sprintf("Unknown struct '%s'", decl.Struct.Token()),
			})

			decl.Type = nil
		}
	} else {
		decl.Type = nil
	}

	// Children
	decl.AcceptChildren(t)
}

// Types

func (t *typeResolver) visitType(type_ ast.Type) {
	if resolvable, ok := type_.(*ast.Resolvable); ok {
		// Resolve type
		var resolver ast.Resolver = t.resolver
		var resolved ast.Type

		for i, part := range resolvable.Parts {
			if i == len(resolvable.Parts)-1 {
				resolved = resolver.GetType(part.String())
			} else {
				if child := resolver.GetChild(part.String()); child != nil {
					resolver = child
				}
			}
		}

		if resolved != nil {
			// Store resolved type
			resolvable.Type = resolved
		} else {
			// Report an error
			str := strings.Builder{}

			for i, part := range resolvable.Parts {
				if i > 0 {
					str.WriteRune('.')
				}

				str.WriteString(part.String())
			}

			t.reporter.Report(utils.Diagnostic{
				Kind:    utils.ErrorKind,
				Range:   resolvable.Cst().Range,
				Message: fmt.Sprintf("Unknown type '%s'", str.String()),
			})

			resolvable.Type = &ast.Primitive{Kind: ast.Void}

			if t.expr != nil {
				t.expr.Result().SetInvalid()
			}
		}
	}

	// Visit children
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
