package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
)

func getImplementations(node ast.Node, pos core.Pos, resolver ast.Resolver) []protocol.Location {
	leaf := ast.GetLeaf(node, pos)

	builder := implementationBuilder{}

	switch parent := leaf.Parent().(type) {
	case *ast.Interface:
		builder.target = parent
		builder.visit(resolver)

	case *ast.Resolvable:
		if inter, ok := ast.As[*ast.Interface](parent); ok {
			builder.target = inter
			builder.visit(resolver)
		}
	}

	return builder.locations
}

type implementationBuilder struct {
	target    *ast.Interface
	locations []protocol.Location
}

func (i *implementationBuilder) visit(resolver ast.Resolver) {
	resolver.GetSymbols(i)

	for _, child := range resolver.GetChildren() {
		i.visit(resolver.GetChild(child))
	}
}

func (i *implementationBuilder) VisitSymbol(node ast.Node) {
	if impl, ok := node.(*ast.Impl); ok && i.target.Equals(impl.Implements) {
		file := ast.GetParent[*ast.File](node)
		if file == nil {
			return
		}

		i.locations = append(i.locations, protocol.Location{
			URI:   uri.New(file.Path),
			Range: convertRange(impl.Struct.Cst().Range),
		})
	}
}
