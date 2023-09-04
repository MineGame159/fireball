package ast

import (
	"fireball/core"
	"log"
)
import "fireball/core/types"
import "fireball/core/scanner"

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
	range_ core.Range
	parent Node

	Token_ scanner.Token
	Stmts  []Stmt
}

func (b *Block) Token() scanner.Token {
	return b.Token_
}

func (b *Block) Range() core.Range {
	return b.range_
}

func (b *Block) SetRangeToken(start, end scanner.Token) {
	b.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (b *Block) SetRangePos(start, end core.Pos) {
	b.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (b *Block) SetRangeNode(start, end Node) {
	b.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (b *Block) Parent() Node {
	return b.parent
}

func (b *Block) SetParent(parent Node) {
	if b.parent != nil && parent != nil {
		log.Fatalln("Block.SetParent() - Node already has a parent")
	}
	b.parent = parent
}

func (b *Block) Accept(visitor StmtVisitor) {
	visitor.VisitBlock(b)
}

func (b *Block) AcceptChildren(visitor Acceptor) {
	for i_ := range b.Stmts {
		if b.Stmts[i_] != nil {
			visitor.AcceptStmt(b.Stmts[i_])
		}
	}
}

func (b *Block) AcceptTypes(visitor types.Visitor) {
}

func (b *Block) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (b *Block) Leaf() bool {
	return false
}

func (b *Block) SetChildrenParent() {
	for i_ := range b.Stmts {
		if b.Stmts[i_] != nil {
			b.Stmts[i_].SetParent(b)
		}
	}
}

// Expression

type Expression struct {
	range_ core.Range
	parent Node

	Token_ scanner.Token
	Expr   Expr
}

func (e *Expression) Token() scanner.Token {
	return e.Token_
}

func (e *Expression) Range() core.Range {
	return e.range_
}

func (e *Expression) SetRangeToken(start, end scanner.Token) {
	e.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (e *Expression) SetRangePos(start, end core.Pos) {
	e.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (e *Expression) SetRangeNode(start, end Node) {
	e.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (e *Expression) Parent() Node {
	return e.parent
}

func (e *Expression) SetParent(parent Node) {
	if e.parent != nil && parent != nil {
		log.Fatalln("Expression.SetParent() - Node already has a parent")
	}
	e.parent = parent
}

func (e *Expression) Accept(visitor StmtVisitor) {
	visitor.VisitExpression(e)
}

func (e *Expression) AcceptChildren(visitor Acceptor) {
	if e.Expr != nil {
		visitor.AcceptExpr(e.Expr)
	}
}

func (e *Expression) AcceptTypes(visitor types.Visitor) {
}

func (e *Expression) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (e *Expression) Leaf() bool {
	return false
}

func (e *Expression) SetChildrenParent() {
	if e.Expr != nil {
		e.Expr.SetParent(e)
	}
}

// Variable

type Variable struct {
	range_ core.Range
	parent Node

	Type        types.Type
	Name        scanner.Token
	Initializer Expr
	InferType   bool
}

func (v *Variable) Token() scanner.Token {
	return v.Name
}

func (v *Variable) Range() core.Range {
	return v.range_
}

func (v *Variable) SetRangeToken(start, end scanner.Token) {
	v.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (v *Variable) SetRangePos(start, end core.Pos) {
	v.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (v *Variable) SetRangeNode(start, end Node) {
	v.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (v *Variable) Parent() Node {
	return v.parent
}

func (v *Variable) SetParent(parent Node) {
	if v.parent != nil && parent != nil {
		log.Fatalln("Variable.SetParent() - Node already has a parent")
	}
	v.parent = parent
}

func (v *Variable) Accept(visitor StmtVisitor) {
	visitor.VisitVariable(v)
}

func (v *Variable) AcceptChildren(visitor Acceptor) {
	if v.Initializer != nil {
		visitor.AcceptExpr(v.Initializer)
	}
}

func (v *Variable) AcceptTypes(visitor types.Visitor) {
	if v.Type != nil {
		visitor.VisitType(v.Type)
	}
}

func (v *Variable) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&v.Type)
}

func (v *Variable) Leaf() bool {
	return false
}

func (v *Variable) SetChildrenParent() {
	if v.Initializer != nil {
		v.Initializer.SetParent(v)
	}
}

// If

type If struct {
	range_ core.Range
	parent Node

	Token_    scanner.Token
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (i *If) Token() scanner.Token {
	return i.Token_
}

func (i *If) Range() core.Range {
	return i.range_
}

func (i *If) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *If) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *If) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *If) Parent() Node {
	return i.parent
}

func (i *If) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("If.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *If) Accept(visitor StmtVisitor) {
	visitor.VisitIf(i)
}

func (i *If) AcceptChildren(visitor Acceptor) {
	if i.Condition != nil {
		visitor.AcceptExpr(i.Condition)
	}
	if i.Then != nil {
		visitor.AcceptStmt(i.Then)
	}
	if i.Else != nil {
		visitor.AcceptStmt(i.Else)
	}
}

func (i *If) AcceptTypes(visitor types.Visitor) {
}

func (i *If) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (i *If) Leaf() bool {
	return false
}

func (i *If) SetChildrenParent() {
	if i.Condition != nil {
		i.Condition.SetParent(i)
	}
	if i.Then != nil {
		i.Then.SetParent(i)
	}
	if i.Else != nil {
		i.Else.SetParent(i)
	}
}

// For

type For struct {
	range_ core.Range
	parent Node

	Token_    scanner.Token
	Condition Expr
	Body      Stmt
}

func (f *For) Token() scanner.Token {
	return f.Token_
}

func (f *For) Range() core.Range {
	return f.range_
}

func (f *For) SetRangeToken(start, end scanner.Token) {
	f.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (f *For) SetRangePos(start, end core.Pos) {
	f.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (f *For) SetRangeNode(start, end Node) {
	f.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (f *For) Parent() Node {
	return f.parent
}

func (f *For) SetParent(parent Node) {
	if f.parent != nil && parent != nil {
		log.Fatalln("For.SetParent() - Node already has a parent")
	}
	f.parent = parent
}

func (f *For) Accept(visitor StmtVisitor) {
	visitor.VisitFor(f)
}

func (f *For) AcceptChildren(visitor Acceptor) {
	if f.Condition != nil {
		visitor.AcceptExpr(f.Condition)
	}
	if f.Body != nil {
		visitor.AcceptStmt(f.Body)
	}
}

func (f *For) AcceptTypes(visitor types.Visitor) {
}

func (f *For) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (f *For) Leaf() bool {
	return false
}

func (f *For) SetChildrenParent() {
	if f.Condition != nil {
		f.Condition.SetParent(f)
	}
	if f.Body != nil {
		f.Body.SetParent(f)
	}
}

// Return

type Return struct {
	range_ core.Range
	parent Node

	Token_ scanner.Token
	Expr   Expr
}

func (r *Return) Token() scanner.Token {
	return r.Token_
}

func (r *Return) Range() core.Range {
	return r.range_
}

func (r *Return) SetRangeToken(start, end scanner.Token) {
	r.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (r *Return) SetRangePos(start, end core.Pos) {
	r.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (r *Return) SetRangeNode(start, end Node) {
	r.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (r *Return) Parent() Node {
	return r.parent
}

func (r *Return) SetParent(parent Node) {
	if r.parent != nil && parent != nil {
		log.Fatalln("Return.SetParent() - Node already has a parent")
	}
	r.parent = parent
}

func (r *Return) Accept(visitor StmtVisitor) {
	visitor.VisitReturn(r)
}

func (r *Return) AcceptChildren(visitor Acceptor) {
	if r.Expr != nil {
		visitor.AcceptExpr(r.Expr)
	}
}

func (r *Return) AcceptTypes(visitor types.Visitor) {
}

func (r *Return) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (r *Return) Leaf() bool {
	return false
}

func (r *Return) SetChildrenParent() {
	if r.Expr != nil {
		r.Expr.SetParent(r)
	}
}

// Break

type Break struct {
	range_ core.Range
	parent Node

	Token_ scanner.Token
}

func (b *Break) Token() scanner.Token {
	return b.Token_
}

func (b *Break) Range() core.Range {
	return b.range_
}

func (b *Break) SetRangeToken(start, end scanner.Token) {
	b.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (b *Break) SetRangePos(start, end core.Pos) {
	b.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (b *Break) SetRangeNode(start, end Node) {
	b.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (b *Break) Parent() Node {
	return b.parent
}

func (b *Break) SetParent(parent Node) {
	if b.parent != nil && parent != nil {
		log.Fatalln("Break.SetParent() - Node already has a parent")
	}
	b.parent = parent
}

func (b *Break) Accept(visitor StmtVisitor) {
	visitor.VisitBreak(b)
}

func (b *Break) AcceptChildren(visitor Acceptor) {
}

func (b *Break) AcceptTypes(visitor types.Visitor) {
}

func (b *Break) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (b *Break) Leaf() bool {
	return true
}

func (b *Break) SetChildrenParent() {
}

// Continue

type Continue struct {
	range_ core.Range
	parent Node

	Token_ scanner.Token
}

func (c *Continue) Token() scanner.Token {
	return c.Token_
}

func (c *Continue) Range() core.Range {
	return c.range_
}

func (c *Continue) SetRangeToken(start, end scanner.Token) {
	c.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (c *Continue) SetRangePos(start, end core.Pos) {
	c.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (c *Continue) SetRangeNode(start, end Node) {
	c.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Continue) Parent() Node {
	return c.parent
}

func (c *Continue) SetParent(parent Node) {
	if c.parent != nil && parent != nil {
		log.Fatalln("Continue.SetParent() - Node already has a parent")
	}
	c.parent = parent
}

func (c *Continue) Accept(visitor StmtVisitor) {
	visitor.VisitContinue(c)
}

func (c *Continue) AcceptChildren(visitor Acceptor) {
}

func (c *Continue) AcceptTypes(visitor types.Visitor) {
}

func (c *Continue) AcceptTypesPtr(visitor types.PtrVisitor) {
}

func (c *Continue) Leaf() bool {
	return true
}

func (c *Continue) SetChildrenParent() {
}
