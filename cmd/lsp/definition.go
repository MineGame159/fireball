package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
)

func getDefinition(node ast.Node, pos core.Pos) []protocol.Location {
	// Get node under cursor
	node = ast.GetLeaf(node, pos)

	// Get definition based on the leaf node
	switch node := node.(type) {
	case *ast.Identifier:
		return getDefinitionExprResult(node.Result())
	case *ast.Resolvable:
		return newDefinition(node.Resolved())
	case *ast.Token:
		return getDefinitionToken(node)
	}

	return nil
}

func getDefinitionToken(token *ast.Token) []protocol.Location {
	// Get definition based on the token's parent
	switch parent := token.Parent().(type) {
	case *ast.Resolvable:
		if token == parent.Parts[len(parent.Parts)-1] {
			return newDefinition(parent.Resolved())
		}

	case *ast.Member:
		return getDefinitionExprResult(parent.Result())

	case *ast.InitField:
		if s, ok := ast.As[*ast.Struct](parent.Parent().(*ast.StructInitializer).Type); ok {
			if _, field := s.GetField(token.String()); field != nil {
				return newDefinition(field)
			}
		}
	}

	return nil
}

func getDefinitionExprResult(result *ast.ExprResult) []protocol.Location {
	switch result.Kind {
	case ast.TypeResultKind:
		return newDefinition(result.Type)
	case ast.ValueResultKind:
		return newDefinition(result.Value())
	case ast.CallableResultKind:
		return newDefinition(result.Callable())

	default:
		return nil
	}
}

func newDefinition(node ast.Node) []protocol.Location {
	if nodeCst(node) == nil {
		return nil
	}

	file := ast.GetParent[*ast.File](node)
	if file == nil {
		return nil
	}

	return []protocol.Location{{
		URI:   uri.New(file.Path),
		Range: convertRange(node.Cst().Range),
	}}
}
