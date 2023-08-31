package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type StmtVisitor interface {
	VisitBlock(stmt *Block)
	VisitExpression(stmt *Expression)
	VisitVariable(stmt *Variable)
	VisitIf(stmt *If)
	VisitFor(stmt *For)
	VisitReturn(stmt *Return)
	VisitBreak(stmt *Break)
	VisitContinue(stmt *Continue)
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

func (b *Block) AcceptChildren(acceptor Acceptor) {
	for _, v := range b.Stmts {
		acceptor.AcceptStmt(v)
	}
}

func (b *Block) AcceptTypes(visitor types.Visitor) {
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

func (e *Expression) AcceptChildren(acceptor Acceptor) {
	if e.Expr != nil {
		acceptor.AcceptExpr(e.Expr)
	}
}

func (e *Expression) AcceptTypes(visitor types.Visitor) {
}

type Variable struct {
	Type        types.Type
	Name        scanner.Token
	Initializer Expr
	InferType   bool
}

func (v *Variable) Token() scanner.Token {
	return v.Name
}

func (v *Variable) Accept(visitor StmtVisitor) {
	visitor.VisitVariable(v)
}

func (v *Variable) AcceptChildren(acceptor Acceptor) {
	if v.Initializer != nil {
		acceptor.AcceptExpr(v.Initializer)
	}
}

func (v *Variable) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(&v.Type)
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

func (i *If) AcceptChildren(acceptor Acceptor) {
	if i.Condition != nil {
		acceptor.AcceptExpr(i.Condition)
	}
	if i.Then != nil {
		acceptor.AcceptStmt(i.Then)
	}
	if i.Else != nil {
		acceptor.AcceptStmt(i.Else)
	}
}

func (i *If) AcceptTypes(visitor types.Visitor) {
}

type For struct {
	Token_    scanner.Token
	Condition Expr
	Body      Stmt
}

func (f *For) Token() scanner.Token {
	return f.Token_
}

func (f *For) Accept(visitor StmtVisitor) {
	visitor.VisitFor(f)
}

func (f *For) AcceptChildren(acceptor Acceptor) {
	if f.Condition != nil {
		acceptor.AcceptExpr(f.Condition)
	}
	if f.Body != nil {
		acceptor.AcceptStmt(f.Body)
	}
}

func (f *For) AcceptTypes(visitor types.Visitor) {
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

func (r *Return) AcceptChildren(acceptor Acceptor) {
	if r.Expr != nil {
		acceptor.AcceptExpr(r.Expr)
	}
}

func (r *Return) AcceptTypes(visitor types.Visitor) {
}

type Break struct {
	Token_ scanner.Token
}

func (b *Break) Token() scanner.Token {
	return b.Token_
}

func (b *Break) Accept(visitor StmtVisitor) {
	visitor.VisitBreak(b)
}

func (b *Break) AcceptChildren(acceptor Acceptor) {
}

func (b *Break) AcceptTypes(visitor types.Visitor) {
}

type Continue struct {
	Token_ scanner.Token
}

func (c *Continue) Token() scanner.Token {
	return c.Token_
}

func (c *Continue) Accept(visitor StmtVisitor) {
	visitor.VisitContinue(c)
}

func (c *Continue) AcceptChildren(acceptor Acceptor) {
}

func (c *Continue) AcceptTypes(visitor types.Visitor) {
}
