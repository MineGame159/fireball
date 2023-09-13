package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"github.com/MineGame159/protocol"
	"strconv"
	"strings"
)

func getHover(decls []ast.Decl, pos core.Pos) *protocol.Hover {
	for _, decl := range decls {
		// Get node under cursor
		node := ast.GetLeaf(decl, pos)

		if expr, ok := node.(ast.Expr); ok && expr.Result().Kind != ast.InvalidResultKind {
			if l, ok := node.(*ast.Literal); ok {
				// ast.Literal
				text := ""

				if strings.HasPrefix(l.Value.Lexeme, "0x") || strings.HasPrefix(l.Value.Lexeme, "0X") {
					v, err := strconv.ParseUint(l.Value.Lexeme[2:], 16, 64)
					if err == nil {
						text = strconv.FormatUint(v, 10)
					}
				} else if strings.HasPrefix(l.Value.Lexeme, "0b") || strings.HasPrefix(l.Value.Lexeme, "0B") {
					v, err := strconv.ParseUint(l.Value.Lexeme[2:], 2, 64)
					if err == nil {
						text = strconv.FormatUint(v, 10)
					}
				}

				if text != "" {
					return &protocol.Hover{
						Contents: protocol.MarkupContent{
							Kind:  protocol.PlainText,
							Value: text,
						},
						Range: convertRangePtr(l.Range()),
					}
				}

				return nil
			} else if i, ok := node.(*ast.Initializer); ok {
				// ast.Initializer
				for _, field := range i.Fields {
					range_ := core.TokenToRange(field.Name)

					if range_.Contains(pos) {
						_, f := i.Result().Type.(*ast.Struct).GetField(field.Name.Lexeme)

						if f != nil {
							return &protocol.Hover{
								Contents: protocol.MarkupContent{
									Kind:  protocol.PlainText,
									Value: f.Type.String(),
								},
								Range: convertRangePtr(range_),
							}
						}
					}
				}
			} else if m, ok := node.(*ast.Member); ok {
				// ast.Member that is an enum
				if i, ok := m.Value.(*ast.Identifier); ok && i.Kind == ast.EnumKind {
					if e, ok := m.Result().Type.(*ast.Enum); ok {
						case_ := e.GetCase(m.Name.Lexeme)

						if case_ != nil {
							return &protocol.Hover{
								Contents: protocol.MarkupContent{
									Kind:  protocol.PlainText,
									Value: strconv.Itoa(case_.Value),
								},
								Range: convertRangePtr(expr.Range()),
							}
						}
					}
				}
			}

			// ast.Expr
			text := expr.Result().Type.String()

			// Ignore literal expressions
			if _, ok := expr.(*ast.Literal); ok {
				text = ""
			}

			// Return
			if text != "" {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: text,
					},
					Range: convertRangePtr(expr.Range()),
				}
			}
		} else if variable, ok := node.(*ast.Variable); ok {
			// ast.Variable
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.PlainText,
					Value: variable.Type.String(),
				},
				Range: convertRangePtr(core.TokenToRange(variable.Name)),
			}
		} else if enum, ok := node.(*ast.Enum); ok {
			// ast.Enum

			for _, case_ := range enum.Cases {
				range_ := core.TokenToRange(case_.Name)

				if range_.Contains(pos) {
					return &protocol.Hover{
						Contents: protocol.MarkupContent{
							Kind:  protocol.PlainText,
							Value: strconv.Itoa(case_.Value),
						},
						Range: convertRangePtr(range_),
					}
				}
			}
		}
	}

	return nil
}
