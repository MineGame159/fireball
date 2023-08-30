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
