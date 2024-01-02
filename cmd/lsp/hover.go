package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"github.com/MineGame159/protocol"
	"strconv"
	"strings"
)

func getHover(node ast.Node, pos core.Pos) *protocol.Hover {
	// Get node under cursor
	node = ast.GetLeaf(node, pos)

	if t, ok := node.(*ast.Token); ok {
		node = t.Parent()
	}

	if expr, ok := node.(ast.Expr); ok && expr.Result().Kind != ast.InvalidResultKind {
		if l, ok := node.(*ast.Literal); ok {
			// ast.Literal
			text := ""

			if strings.HasPrefix(l.String(), "0x") || strings.HasPrefix(l.String(), "0X") {
				v, err := strconv.ParseUint(l.String()[2:], 16, 64)
				if err == nil {
					text = strconv.FormatUint(v, 10)
				}
			} else if strings.HasPrefix(l.String(), "0b") || strings.HasPrefix(l.String(), "0B") {
				v, err := strconv.ParseUint(l.String()[2:], 2, 64)
				if err == nil {
					text = strconv.FormatUint(v, 10)
				}
			}

			if text != "" && l.Cst() != nil {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: text,
					},
					Range: convertRangePtr(l.Cst().Range),
				}
			}

			return nil
		} else if i, ok := node.(*ast.StructInitializer); ok {
			// ast.Initializer
			for _, field := range i.Fields {
				if field.Cst() != nil && field.Name.Cst().Range.Contains(pos) {
					if struct_, ok := ast.As[*ast.Struct](i.Type); ok {
						if _, f := struct_.GetField(field.Name.String()); f != nil {
							return &protocol.Hover{
								Contents: protocol.MarkupContent{
									Kind:  protocol.PlainText,
									Value: ast.PrintType(f.Type),
								},
								Range: convertRangePtr(field.Name.Cst().Range),
							}
						}
					}
				}
			}
		} else if m, ok := node.(*ast.Member); ok {
			// ast.Member that is an enum
			if i, ok := m.Value.(*ast.Identifier); ok && i.Kind == ast.EnumKind {
				if e, ok := ast.As[*ast.Enum](m.Result().Type); ok {
					case_ := e.GetCase(m.Name.String())

					if case_ != nil && m.Name.Cst() != nil {
						return &protocol.Hover{
							Contents: protocol.MarkupContent{
								Kind:  protocol.PlainText,
								Value: strconv.FormatInt(case_.ActualValue, 10),
							},
							Range: convertRangePtr(m.Name.Cst().Range),
						}
					}
				}
			}
		} else if t, ok := node.(*ast.TypeCall); ok {
			// ast.TypeCall
			value := uint32(0)

			if t.Callee.String() == "sizeof" {
				value = t.Arg.Size()
			} else {
				value = t.Arg.Align()
			}

			if t.Callee.Cst() != nil {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: strconv.FormatUint(uint64(value), 10),
					},
					Range: convertRangePtr(t.Callee.Cst().Range),
				}
			}
		}

		// ast.Expr
		text := ast.PrintType(expr.Result().Type.Resolved())

		// Ignore literal expressions
		if _, ok := expr.(*ast.Literal); ok {
			text = ""
		}

		// Return
		if text != "" && expr.Cst() != nil {
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.PlainText,
					Value: text,
				},
				Range: convertRangePtr(expr.Cst().Range),
			}
		}
	} else if variable, ok := node.(*ast.Var); ok && variable.Name.Cst() != nil {
		// ast.Var
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.PlainText,
				Value: ast.PrintType(variable.ActualType),
			},
			Range: convertRangePtr(variable.Name.Cst().Range),
		}
	} else if enum, ok := node.(*ast.Enum); ok {
		// ast.Enum

		for _, case_ := range enum.Cases {
			if case_.Name.Cst() != nil && case_.Name.Cst().Range.Contains(pos) {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: strconv.FormatInt(case_.ActualValue, 10),
					},
					Range: convertRangePtr(case_.Name.Cst().Range),
				}
			}
		}
	}

	return nil
}
