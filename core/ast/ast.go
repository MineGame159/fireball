package ast

import (
	"fireball/core/scanner"
	"fireball/core/types"
)

type Node interface {
	Token() scanner.Token

	AcceptChildren(acceptor Acceptor)
	AcceptTypes(visitor types.Visitor)
}

type Acceptor interface {
	AcceptDecl(decl Decl)
	AcceptStmt(stmt Stmt)
	AcceptExpr(expr Expr)
}

type Range struct {
	Start Pos
	End   Pos
}

type Pos struct {
	Line   int
	Column int
}

func TokenToRange(token scanner.Token) Range {
	return Range{
		Start: TokenToPos(token, false),
		End:   TokenToPos(token, true),
	}
}

func TokenToPos(token scanner.Token, end bool) Pos {
	offset := 0

	if end {
		offset = len(token.Lexeme)
	}

	return Pos{
		Line:   token.Line,
		Column: token.Column + offset,
	}
}

func VisitStmts[T Stmt](node Node, callback func(stmt T)) {
	visitor[T]{
		expressions: false,
		callback:    callback,
	}.Accept(node)
}

func VisitExprs[T Expr](node Node, callback func(expr T)) {
	visitor[T]{
		expressions: true,
		callback:    callback,
	}.Accept(node)
}

type visitor[T Node] struct {
	expressions bool
	callback    func(node T)
}

func (v visitor[T]) Accept(node Node) {
	if v2, ok := node.(T); ok {
		v.callback(v2)
	}

	node.AcceptChildren(v)
}

func (v visitor[T]) AcceptDecl(decl Decl) {
	v.Accept(decl)
}

func (v visitor[T]) AcceptStmt(stmt Stmt) {
	v.Accept(stmt)
}

func (v visitor[T]) AcceptExpr(expr Expr) {
	if v.expressions {
		v.Accept(expr)
	}
}
