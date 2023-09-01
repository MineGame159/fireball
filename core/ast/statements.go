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

// Block

type Block struct {
	range_ Range

	Token_ scanner.Token
	Stmts  []Stmt
}

func (b *Block) Token() scanner.Token {
	return b.Token_
}

func (b *Block) Range() Range {
	return b.range_
}

func (b *Block) SetRangeToken(start, end scanner.Token) {
	b.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (b *Block) SetRangePos(start, end Pos) {
	b.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (b *Block) SetRangeNode(start, end Node) {
	b.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (b *Block) Leaf() bool {
	return false
}

// Expression

type Expression struct {
	range_ Range

	Token_ scanner.Token
	Expr   Expr
}

func (e *Expression) Token() scanner.Token {
	return e.Token_
}

func (e *Expression) Range() Range {
	return e.range_
}

func (e *Expression) SetRangeToken(start, end scanner.Token) {
	e.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (e *Expression) SetRangePos(start, end Pos) {
	e.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (e *Expression) SetRangeNode(start, end Node) {
	e.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (e *Expression) Leaf() bool {
	return false
}

// Variable

type Variable struct {
	range_ Range

	Type        types.Type
	Name        scanner.Token
	Initializer Expr
	InferType   bool
}

func (v *Variable) Token() scanner.Token {
	return v.Name
}

func (v *Variable) Range() Range {
	return v.range_
}

func (v *Variable) SetRangeToken(start, end scanner.Token) {
	v.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (v *Variable) SetRangePos(start, end Pos) {
	v.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (v *Variable) SetRangeNode(start, end Node) {
	v.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (v *Variable) Leaf() bool {
	return false
}

// If

type If struct {
	range_ Range

	Token_    scanner.Token
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (i *If) Token() scanner.Token {
	return i.Token_
}

func (i *If) Range() Range {
	return i.range_
}

func (i *If) SetRangeToken(start, end scanner.Token) {
	i.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (i *If) SetRangePos(start, end Pos) {
	i.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (i *If) SetRangeNode(start, end Node) {
	i.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (i *If) Leaf() bool {
	return false
}

// For

type For struct {
	range_ Range

	Token_    scanner.Token
	Condition Expr
	Body      Stmt
}

func (f *For) Token() scanner.Token {
	return f.Token_
}

func (f *For) Range() Range {
	return f.range_
}

func (f *For) SetRangeToken(start, end scanner.Token) {
	f.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (f *For) SetRangePos(start, end Pos) {
	f.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (f *For) SetRangeNode(start, end Node) {
	f.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (f *For) Leaf() bool {
	return false
}

// Return

type Return struct {
	range_ Range

	Token_ scanner.Token
	Expr   Expr
}

func (r *Return) Token() scanner.Token {
	return r.Token_
}

func (r *Return) Range() Range {
	return r.range_
}

func (r *Return) SetRangeToken(start, end scanner.Token) {
	r.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (r *Return) SetRangePos(start, end Pos) {
	r.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (r *Return) SetRangeNode(start, end Node) {
	r.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (r *Return) Leaf() bool {
	return false
}

// Break

type Break struct {
	range_ Range

	Token_ scanner.Token
}

func (b *Break) Token() scanner.Token {
	return b.Token_
}

func (b *Break) Range() Range {
	return b.range_
}

func (b *Break) SetRangeToken(start, end scanner.Token) {
	b.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (b *Break) SetRangePos(start, end Pos) {
	b.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (b *Break) SetRangeNode(start, end Node) {
	b.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (b *Break) Accept(visitor StmtVisitor) {
	visitor.VisitBreak(b)
}

func (b *Break) AcceptChildren(acceptor Acceptor) {
}

func (b *Break) AcceptTypes(visitor types.Visitor) {
}

func (b *Break) Leaf() bool {
	return true
}

// Continue

type Continue struct {
	range_ Range

	Token_ scanner.Token
}

func (c *Continue) Token() scanner.Token {
	return c.Token_
}

func (c *Continue) Range() Range {
	return c.range_
}

func (c *Continue) SetRangeToken(start, end scanner.Token) {
	c.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (c *Continue) SetRangePos(start, end Pos) {
	c.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (c *Continue) SetRangeNode(start, end Node) {
	c.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Continue) Accept(visitor StmtVisitor) {
	visitor.VisitContinue(c)
}

func (c *Continue) AcceptChildren(acceptor Acceptor) {
}

func (c *Continue) AcceptTypes(visitor types.Visitor) {
}

func (c *Continue) Leaf() bool {
	return true
}
