package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type StmtVisitor interface {
	VisitBlock(stmt *Block)
	VisitExpression(stmt *Expression)
	VisitVariable(stmt *Variable)
	VisitIf(stmt *If)
	VisitReturn(stmt *Return)
}

type Stmt interface {
	Node

	Accept(visitor StmtVisitor)
}

type Block struct {
	Token_ scanner.Token
	Stmts  []Stmt
}

func (b *Block) Token() scanner.Token {
	return b.Token_
}

func (b *Block) Accept(visitor StmtVisitor) {
	visitor.VisitBlock(b)
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

type If struct {
	Token_    scanner.Token
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (i *If) Token() scanner.Token {
	return i.Token_
}

func (i *If) Accept(visitor StmtVisitor) {
	visitor.VisitIf(i)
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
