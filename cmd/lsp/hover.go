package lsp

import (
	"fireball/core"
	"fireball/core/abi"
	"fireball/core/ast"
	"github.com/MineGame159/protocol"
	"strconv"
	"strings"
)

func getHover(node ast.Node, pos core.Pos) *protocol.Hover {
	// Get node under cursor
	node = ast.GetLeaf(node, pos)
	if node == nil || node.Cst() == nil {
		return nil
	}

	// Switch based on the leaf node
	switch node := node.(type) {
	case *ast.Identifier:
		return getHoverExprResult(node.Result(), node)

	case *ast.Literal:
		if strings.HasPrefix(node.String(), "0x") || strings.HasPrefix(node.String(), "0X") {
			v, err := strconv.ParseUint(node.String()[2:], 16, 64)
			if err == nil {
				return newHover(node, strconv.FormatUint(v, 10))
			}
		} else if strings.HasPrefix(node.String(), "0b") || strings.HasPrefix(node.String(), "0B") {
			v, err := strconv.ParseUint(node.String()[2:], 2, 64)
			if err == nil {
				return newHover(node, strconv.FormatUint(v, 10))
			}
		}

	case *ast.Token:
		return getHoverToken(node)
	}

	return nil
}

func getHoverToken(token *ast.Token) *protocol.Hover {
	// Switch based on the token's parent node
	switch parent := token.Parent().(type) {
	case *ast.Field:
		return newHover(token, printType(parent.Type))

	case *ast.EnumCase:
		return newHover(token, strconv.FormatInt(parent.ActualValue, 10))

	case *ast.Param:
		return newHover(token, printType(parent.Type))

	case *ast.Var:
		return newHover(token, printType(parent.ActualType))

	case *ast.Member:
		return getHoverExprResult(parent.Result(), token)

	case *ast.TypeCall:
		if parent.Callee == nil || parent.Arg == nil {
			return nil
		}

		value := uint32(0)

		if parent.Callee.String() == "sizeof" {
			value = abi.GetTargetAbi().Size(parent.Arg)
		} else {
			value = abi.GetTargetAbi().Align(parent.Arg)
		}

		return newHover(token, strconv.FormatUint(uint64(value), 10))

	case *ast.InitField:
		if s, ok := ast.As[*ast.Struct](parent.Parent().(*ast.StructInitializer).Type); ok {
			if _, field := s.GetField(token.String()); field != nil {
				return newHover(token, ast.PrintType(field.Type))
			}
		}
	}

	return nil
}

func getHoverExprResult(result *ast.ExprResult, node ast.Node) *protocol.Hover {
	if result.Kind == ast.ValueResultKind {
		if case_, ok := result.Value().(*ast.EnumCase); ok {
			return newHover(node, strconv.FormatInt(case_.ActualValue, 10))
		}
	}

	return newHover(node, printType(result.Type))
}

func newHover(node ast.Node, text string) *protocol.Hover {
	if text == "" {
		return nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.PlainText,
			Value: text,
		},
		Range: convertRangePtr(node.Cst().Range),
	}
}
