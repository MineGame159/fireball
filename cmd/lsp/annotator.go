package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fmt"
	"github.com/MineGame159/protocol"
)

type annotator struct {
	functions map[string]*ast.Func
	hints     []protocol.InlayHint
}

func annotate(node ast.Node) []protocol.InlayHint {
	a := annotator{
		functions: make(map[string]*ast.Func),
		hints:     make([]protocol.InlayHint, 0, 16),
	}

	for _, decl := range node.(*ast.File).Decls {
		if f, ok := decl.(*ast.Func); ok {
			a.functions[f.Name.String()] = f
		}
	}

	a.VisitNode(node)

	return a.hints
}

// Declarations

func (a *annotator) visitEnum(decl *ast.Enum) {
	if decl.Type == nil {
		a.addToken(decl.Name, " : "+ast.PrintType(decl.ActualType), protocol.InlayHintKindType)
	}

	for _, case_ := range decl.Cases {
		if case_.Value == nil {
			a.addToken(case_.Name, fmt.Sprintf(" = %d", case_.ActualValue), protocol.InlayHintKindParameter)
		}
	}

	decl.AcceptChildren(a)
}

// Statements

func (a *annotator) visitVar(stmt *ast.Var) {
	if stmt.Type == nil {
		a.addToken(stmt.Name, " "+ast.PrintType(stmt.ActualType), protocol.InlayHintKindType)
	}

	stmt.AcceptChildren(a)
}

// Expressions

func (a *annotator) visitCall(expr *ast.Call) {
	if false {
		if function, ok := expr.Callee.Result().Type.(*ast.Func); ok {
			for i, arg := range expr.Args {
				if i >= len(function.Params) {
					break
				}

				param := function.Params[i]

				if arg.Cst() != nil {
					a.add(arg.Cst().Range.Start, param.Name.String()+": ", protocol.InlayHintKindParameter)
				}
			}
		}
	}

	expr.AcceptChildren(a)
}

// ast.Visitor

func (a *annotator) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Enum:
		a.visitEnum(node)

	case *ast.Var:
		a.visitVar(node)

	case *ast.Call:
		a.visitCall(node)

	default:
		node.AcceptChildren(a)
	}
}

// Utils

func (a *annotator) addToken(node ast.Node, text string, kind protocol.InlayHintKind) {
	if node.Cst() != nil {
		a.add(node.Cst().Range.End, text, kind)
	}
}

func (a *annotator) add(pos core.Pos, text string, kind protocol.InlayHintKind) {
	a.hints = append(a.hints, protocol.InlayHint{
		Position: convertPos(pos),
		Label:    text,
		Kind:     kind,
	})
}
