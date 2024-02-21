package typeresolver

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/utils"
	"fmt"
	"strings"
)

type typeResolver struct {
	expr ast.Expr

	reporter utils.Reporter
	resolver ast.Resolver
}

func Resolve(reporter utils.Reporter, root ast.RootResolver, file *ast.File) {
	resolver := ast.NewCombinedResolver(root)

	r := typeResolver{
		reporter: reporter,
		resolver: resolver,
	}

	for _, decl := range file.Decls {
		if using, ok := decl.(*ast.Using); ok {
			if resolver2 := root.GetResolver(using.Name); resolver2 != nil {
				resolver.Add(resolver2)
			}
		}
	}

	r.VisitNode(file)
}

// Declarations

func (t *typeResolver) visitStruct(decl *ast.Struct) {
	prevResolver := t.resolver
	if len(decl.GenericParams) != 0 {
		t.resolver = ast.NewGenericResolver(t.resolver, decl.GenericParams)
	}

	decl.AcceptChildren(t)

	t.resolver = prevResolver
}

func (t *typeResolver) visitImpl(decl *ast.Impl) {
	prevResolver := t.resolver

	// Struct
	if decl.Struct != nil {
		type_ := t.resolver.GetType(decl.Struct.Token().Lexeme)

		if s, ok := type_.(*ast.Struct); ok {
			decl.Type = s

			if len(s.GenericParams) > 0 {
				t.resolver = ast.NewGenericResolver(t.resolver, s.GenericParams)
			}
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

	t.resolver = prevResolver
}

func (t *typeResolver) visitFunc(decl *ast.Func) {
	prevResolver := t.resolver
	if len(decl.GenericParams) != 0 {
		t.resolver = ast.NewGenericResolver(t.resolver, decl.GenericParams)
	}

	decl.AcceptChildren(t)

	t.resolver = prevResolver
}

// Types

func (t *typeResolver) visitType(type_ ast.Type) {
	if resolvable, ok := type_.(*ast.Resolvable); ok {
		// Resolve type
		resolver := t.resolver
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
			// Visit children
			type_.AcceptChildren(t)

			// Specialize struct if needed
			if s, ok := ast.As[*ast.Struct](resolved); ok {
				if len(resolvable.GenericArgs) != len(s.GenericParams) {
					if resolvable.GenericArgs != nil {
						errorSlice(t, resolvable.GenericArgs, "Got '%d' generic arguments but struct takes '%d'", len(resolvable.GenericArgs), len(s.GenericParams))
					} else {
						t.error(resolvable.Parts[len(resolvable.Parts)-1], "Got '%d' generic arguments but struct takes '%d'", len(resolvable.GenericArgs), len(s.GenericParams))
					}
				}

				if len(resolvable.GenericArgs) != 0 {
					resolved = s.Specialize(resolvable.GenericArgs)
				}
			} else if len(resolvable.GenericArgs) != 0 {
				errorSlice(t, resolvable.GenericArgs, "This type doesn't have any generic parameters")
			}

			// Store resolved type
			resolvable.Type = resolved

			return
		} else {
			// Report an error
			str := strings.Builder{}

			for i, part := range resolvable.Parts {
				if i > 0 {
					str.WriteRune('.')
				}

				str.WriteString(part.String())
			}

			t.error(resolvable, "Unknown type '%s'", str.String())
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
	case *ast.Struct:
		t.visitStruct(node)

	case *ast.Impl:
		t.visitImpl(node)

	case *ast.Func:
		t.visitFunc(node)

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

// Utils

func (t *typeResolver) error(node ast.Node, format string, args ...any) {
	if ast.IsNil(node) {
		return
	}

	t.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Cst().Range,
		Message: fmt.Sprintf(format, args...),
	})
}

func errorSlice[T ast.Node](t *typeResolver, nodes []T, format string, args ...any) {
	start := nodes[0].Cst().Range.Start
	end := nodes[len(nodes)-1].Cst().Range.End

	t.reporter.Report(utils.Diagnostic{
		Kind: utils.ErrorKind,
		Range: core.Range{
			Start: start,
			End:   end,
		},
		Message: fmt.Sprintf(format, args...),
	})
}
