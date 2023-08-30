package lsp

import (
	"fireball/core/ast"
	"fireball/core/scanner"
)

type highlighter struct {
	functions map[string]struct{}
	params    []ast.Param

	data []uint32

	lastLine   int
	lastColumn int
}

func highlight(decls []ast.Decl) []uint32 {
	h := &highlighter{
		functions: make(map[string]struct{}),
		data:      make([]uint32, 0, 256),
	}

	for _, decl := range decls {
		if v, ok := decl.(*ast.Func); ok {
			h.functions[v.Name.Lexeme] = struct{}{}
		}
	}

	for _, decl := range decls {
		h.AcceptDecl(decl)
	}

	return h.data
}

// Declarations

func (h *highlighter) VisitStruct(decl *ast.Struct) {
	h.add(decl.Name, classKind)

	for _, field := range decl.Fields {
		h.add(field.Name, propertyKind)
	}

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitFunc(decl *ast.Func) {
	h.add(decl.Name, functionKind)

	for _, param := range decl.Params {
		h.add(param.Name, parameterKind)
	}

	h.params = decl.Params
	decl.AcceptChildren(h)
	h.params = nil
}

// Statements

func (h *highlighter) VisitBlock(stmt *ast.Block) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitVariable(stmt *ast.Variable) {
	h.add(stmt.Name, variableKind)

	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitFor(stmt *ast.For) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(h)
}

// Expressions

func (h *highlighter) VisitGroup(expr *ast.Group) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitIdentifier(expr *ast.Identifier) {
	if _, ok := h.functions[expr.Identifier.Lexeme]; ok {
		h.add(expr.Identifier, functionKind)
	} else if h.isParameter(expr.Identifier) {
		h.add(expr.Identifier, parameterKind)
	} else {
		h.add(expr.Identifier, variableKind)
	}

	expr.AcceptChildren(h)
}

func (h *highlighter) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitMember(expr *ast.Member) {
	expr.AcceptChildren(h)

	h.add(expr.Name, propertyKind)
}

// ast.Acceptor

func (h *highlighter) AcceptDecl(decl ast.Decl) {
	decl.Accept(h)
}

func (h *highlighter) AcceptStmt(stmt ast.Stmt) {
	stmt.Accept(h)
}

func (h *highlighter) AcceptExpr(expr ast.Expr) {
	expr.Accept(h)
}

// Tokens

type semanticKind uint32

const (
	functionKind semanticKind = iota
	parameterKind
	variableKind
	classKind
	propertyKind
)

func (h *highlighter) add(token scanner.Token, kind semanticKind) {
	if h.lastLine != token.Line-1 {
		h.lastColumn = 0
	}

	h.data = append(h.data, uint32(token.Line-1-h.lastLine))
	h.data = append(h.data, uint32(token.Column-h.lastColumn))
	h.data = append(h.data, uint32(len(token.Lexeme)))
	h.data = append(h.data, uint32(kind))
	h.data = append(h.data, 0)

	h.lastLine = token.Line - 1
	h.lastColumn = token.Column
}

// Utils

func (h *highlighter) isParameter(name scanner.Token) bool {
	for _, param := range h.params {
		if param.Name.Lexeme == name.Lexeme {
			return true
		}
	}

	return false
}
