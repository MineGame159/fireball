package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type ExprVisitor interface {
	VisitGroup(expr *Group)
	VisitLiteral(expr *Literal)
	VisitUnary(expr *Unary)
	VisitBinary(expr *Binary)
	VisitIdentifier(expr *Identifier)
	VisitAssignment(expr *Assignment)
	VisitCast(expr *Cast)
	VisitCall(expr *Call)
}

type Expr interface {
	Node

	Accept(visitor ExprVisitor)

	Type() types.Type
	SetType(type_ types.Type)
}

type Group struct {
	type_ types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (g *Group) Token() scanner.Token {
	return g.Token_
}

func (g *Group) Accept(visitor ExprVisitor) {
	visitor.VisitGroup(g)
}

func (g *Group) Type() types.Type {
	return g.type_
}

func (g *Group) SetType(type_ types.Type) {
	g.type_ = type_
}

type Literal struct {
	type_ types.Type

	Value scanner.Token
}

func (l *Literal) Token() scanner.Token {
	return l.Value
}

func (l *Literal) Accept(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) Type() types.Type {
	return l.type_
}

func (l *Literal) SetType(type_ types.Type) {
	l.type_ = type_
}

type Unary struct {
	type_ types.Type

	Op    scanner.Token
	Right Expr
}

func (u *Unary) Token() scanner.Token {
	return u.Op
}

func (u *Unary) Accept(visitor ExprVisitor) {
	visitor.VisitUnary(u)
}

func (u *Unary) Type() types.Type {
	return u.type_
}

func (u *Unary) SetType(type_ types.Type) {
	u.type_ = type_
}

type Binary struct {
	type_ types.Type

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (b *Binary) Token() scanner.Token {
	return b.Op
}

func (b *Binary) Accept(visitor ExprVisitor) {
	visitor.VisitBinary(b)
}

func (b *Binary) Type() types.Type {
	return b.type_
}

func (b *Binary) SetType(type_ types.Type) {
	b.type_ = type_
}

type Identifier struct {
	type_ types.Type

	Identifier scanner.Token
}

func (i *Identifier) Token() scanner.Token {
	return i.Identifier
}

func (i *Identifier) Accept(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) Type() types.Type {
	return i.type_
}

func (i *Identifier) SetType(type_ types.Type) {
	i.type_ = type_
}

type Assignment struct {
	type_ types.Type

	Assignee Expr
	Op       scanner.Token
	Value    Expr
}

func (a *Assignment) Token() scanner.Token {
	return a.Op
}

func (a *Assignment) Accept(visitor ExprVisitor) {
	visitor.VisitAssignment(a)
}

func (a *Assignment) Type() types.Type {
	return a.type_
}

func (a *Assignment) SetType(type_ types.Type) {
	a.type_ = type_
}

type Cast struct {
	type_ types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (c *Cast) Token() scanner.Token {
	return c.Token_
}

func (c *Cast) Accept(visitor ExprVisitor) {
	visitor.VisitCast(c)
}

func (c *Cast) Type() types.Type {
	return c.type_
}

func (c *Cast) SetType(type_ types.Type) {
	c.type_ = type_
}

type Call struct {
	type_ types.Type

	Token_ scanner.Token
	Callee Expr
	Args   []Expr
}

func (c *Call) Token() scanner.Token {
	return c.Token_
}

func (c *Call) Accept(visitor ExprVisitor) {
	visitor.VisitCall(c)
}

func (c *Call) Type() types.Type {
	return c.type_
}

func (c *Call) SetType(type_ types.Type) {
	c.type_ = type_
}
