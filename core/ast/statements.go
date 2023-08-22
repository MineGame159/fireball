package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type StmtVisitor interface {
	VisitExpression(stmt *Expression)
	VisitVariable(stmt *Variable)
	VisitReturn(stmt *Return)
}

type Stmt interface {
	Node

	Accept(visitor StmtVisitor)
}

type Expression struct {
	Token_ scanner.Token
	Expr   Expr
}

func (e *Expression) Token() scanner.Token {
	return e.Token_
}

func (e *Expression) Accept(visitor StmtVisitor) {
	visitor.VisitExpression(e)
}

type Variable struct {
	Type        types.Type
	Name        scanner.Token
	Initializer Expr
}

func (v *Variable) Token() scanner.Token {
	return v.Name
}

func (v *Variable) Accept(visitor StmtVisitor) {
	visitor.VisitVariable(v)
}

type Return struct {
	Token_ scanner.Token
	Expr   Expr
}

func (r *Return) Token() scanner.Token {
	return r.Token_
}

func (r *Return) Accept(visitor StmtVisitor) {
	visitor.VisitReturn(r)
}
