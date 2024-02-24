package typeresolver

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
	"fmt"
	"math"
	"strconv"
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

func (t *typeResolver) visitEnum(decl *ast.Enum) {
	// Set case values
	lastValue := int64(-1)

	for _, case_ := range decl.Cases {
		if case_.Value == nil {
			lastValue++
			case_.ActualValue = lastValue
		} else {
			var value int64
			var err error

			switch case_.Value.Token().Kind {
			case scanner.Number:
				value, err = strconv.ParseInt(case_.Value.String(), 10, 64)
			case scanner.Hex:
				value, err = strconv.ParseInt(case_.Value.String()[2:], 16, 64)
			case scanner.Binary:
				value, err = strconv.ParseInt(case_.Value.String()[2:], 2, 64)

			default:
				panic("checker.VisitEnum() - Not implemented")
			}

			if err == nil {
				lastValue = value
				case_.ActualValue = value
			} else {
				errorNode(t.reporter, case_.Value, "Failed to parse number")

				lastValue++
				case_.ActualValue = lastValue
			}
		}
	}

	// Find type
	if decl.Type == nil {
		minValue := int64(math.MaxInt64)
		maxValue := int64(math.MinInt64)

		for _, case_ := range decl.Cases {
			minValue = min(minValue, case_.ActualValue)
			maxValue = max(maxValue, case_.ActualValue)
		}

		var kind ast.PrimitiveKind

		if minValue >= 0 {
			// Unsigned
			if maxValue <= math.MaxUint8 {
				kind = ast.U8
			} else if maxValue <= math.MaxUint16 {
				kind = ast.U16
			} else if maxValue <= math.MaxUint32 {
				kind = ast.U32
			} else {
				kind = ast.U64
			}
		} else {
			// Signed
			if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
				kind = ast.I8
			} else if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
				kind = ast.I16
			} else if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
				kind = ast.I32
			} else {
				kind = ast.I64
			}
		}

		decl.ActualType = &ast.Primitive{Kind: kind}
	} else {
		decl.ActualType = decl.Type
	}
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

			errorNode(t.reporter, resolvable, "Unknown type '%s'", str.String())
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

	case *ast.Enum:
		t.visitEnum(node)

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

func errorNode(reporter utils.Reporter, node ast.Node, format string, args ...any) {
	if ast.IsNil(node) {
		return
	}

	reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Cst().Range,
		Message: fmt.Sprintf(format, args...),
	})
}

func errorSlice[T ast.Node](reporter utils.Reporter, nodes []T, format string, args ...any) {
	start := nodes[0].Cst().Range.Start
	end := nodes[len(nodes)-1].Cst().Range.End

	reporter.Report(utils.Diagnostic{
		Kind: utils.ErrorKind,
		Range: core.Range{
			Start: start,
			End:   end,
		},
		Message: fmt.Sprintf(format, args...),
	})
}
