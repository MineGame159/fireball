package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/fuckoff"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
)

func getDefinition(resolver fuckoff.Resolver, node ast.Node, pos core.Pos) []protocol.Location {
	// Get node under cursor
	node = ast.GetLeaf(node, pos)

	// Get definition based on the leaf node
	switch node := node.(type) {
	case *ast.Identifier:
		return getDefinitionIdentifier(pos, node)

	case *ast.Resolvable:
		return newDefinition(node.Resolved())

	case *ast.Token:
		return getDefinitionToken(resolver, node)
	}

	return nil
}

func getDefinitionIdentifier(pos core.Pos, identifier *ast.Identifier) []protocol.Location {
	switch identifier.Kind {
	case ast.StructKind, ast.EnumKind:
		return newDefinition(identifier.Result().Type)

	case ast.FunctionKind:
		return newDefinition(identifier.Result().Function)

	case ast.ParameterKind:
		function := ast.GetParent[*ast.Func](identifier)
		if function == nil {
			return nil
		}

		for _, param := range function.Params {
			if param.Name.String() == identifier.String() {
				return newDefinition(param)
			}
		}

	case ast.VariableKind:
		function := ast.GetParent[*ast.Func](identifier)
		if function == nil {
			return nil
		}

		resolver := variableResolver{target: pos, targetVariableName: identifier}
		resolver.VisitNode(function)

		if resolver.variable != nil {
			return newDefinition(resolver.variable)
		}
	}

	return nil
}

func getDefinitionToken(resolver fuckoff.Resolver, token *ast.Token) []protocol.Location {
	// Get definition based on the token's parent
	switch parent := token.Parent().(type) {
	case *ast.Member:
		if s, ok := asThroughPointer[*ast.Struct](parent.Value.Result().Type); ok {
			if parentWantsFunction(parent) {
				method, _ := resolver.GetMethod(s, token.String(), false)
				if method == nil {
					method, _ = resolver.GetMethod(s, token.String(), true)
				}

				if method != nil {
					return newDefinition(method)
				}
			} else {
				_, field := s.GetField(token.String())
				if field == nil {
					_, field = s.GetStaticField(token.String())
				}

				if field != nil {
					return newDefinition(field)
				}
			}
		} else if e, ok := ast.As[*ast.Enum](parent.Value.Result().Type); ok {
			case_ := e.GetCase(token.String())

			if case_ != nil {
				return newDefinition(case_)
			}
		}

	case *ast.InitField:
		if s, ok := ast.As[*ast.Struct](parent.Parent().(*ast.StructInitializer).Type); ok {
			if _, field := s.GetField(token.String()); field != nil {
				return newDefinition(field)
			}
		}
	}

	return nil
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
