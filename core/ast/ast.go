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
