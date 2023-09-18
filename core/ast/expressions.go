package ast

import "log"
import "fireball/core"
import "fireball/core/types"
import "fireball/core/scanner"

//go:generate go run ../../gen/ast.go

type ExprVisitor interface {
	VisitGroup(expr *Group)
	VisitLiteral(expr *Literal)
	VisitStructInitializer(expr *StructInitializer)
	VisitArrayInitializer(expr *ArrayInitializer)
	VisitUnary(expr *Unary)
	VisitBinary(expr *Binary)
	VisitLogical(expr *Logical)
	VisitIdentifier(expr *Identifier)
	VisitAssignment(expr *Assignment)
	VisitCast(expr *Cast)
	VisitSizeof(expr *Sizeof)
	VisitCall(expr *Call)
	VisitIndex(expr *Index)
	VisitMember(expr *Member)
}

type Expr interface {
	Node

	Accept(visitor ExprVisitor)

	Result() *ExprResult
}

// Group

type Group struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Expr   Expr
}

func (g *Group) Token() scanner.Token {
	return g.Token_
}

func (g *Group) Range() core.Range {
	return g.range_
}

func (g *Group) SetRangeToken(start, end scanner.Token) {
	g.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (g *Group) SetRangePos(start, end core.Pos) {
	g.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (g *Group) SetRangeNode(start, end Node) {
	g.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (g *Group) Parent() Node {
	return g.parent
}

func (g *Group) SetParent(parent Node) {
	if g.parent != nil && parent != nil {
		log.Fatalln("Group.SetParent() - Node already has a parent")
	}
	g.parent = parent
}

func (g *Group) Accept(visitor ExprVisitor) {
	visitor.VisitGroup(g)
}

func (g *Group) AcceptChildren(visitor Acceptor) {
	if g.Expr != nil {
		visitor.AcceptExpr(g.Expr)
	}
}

func (g *Group) AcceptTypes(visitor types.Visitor) {
	if g.result.Type != nil {
		visitor.VisitType(g.result.Type)
	}
}

func (g *Group) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&g.result.Type)
}

func (g *Group) Leaf() bool {
	return false
}

func (g *Group) String() string {
	return g.Token().Lexeme
}

func (g *Group) Result() *ExprResult {
	return &g.result
}

func (g *Group) SetChildrenParent() {
	if g.Expr != nil {
		g.Expr.SetParent(g)
	}
}

// Literal

type Literal struct {
	range_ core.Range
	parent Node
	result ExprResult

	Value scanner.Token
}

func (l *Literal) Token() scanner.Token {
	return l.Value
}

func (l *Literal) Range() core.Range {
	return l.range_
}

func (l *Literal) SetRangeToken(start, end scanner.Token) {
	l.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (l *Literal) SetRangePos(start, end core.Pos) {
	l.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (l *Literal) SetRangeNode(start, end Node) {
	l.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (l *Literal) Parent() Node {
	return l.parent
}

func (l *Literal) SetParent(parent Node) {
	if l.parent != nil && parent != nil {
		log.Fatalln("Literal.SetParent() - Node already has a parent")
	}
	l.parent = parent
}

func (l *Literal) Accept(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) AcceptChildren(visitor Acceptor) {
}

func (l *Literal) AcceptTypes(visitor types.Visitor) {
	if l.result.Type != nil {
		visitor.VisitType(l.result.Type)
	}
}

func (l *Literal) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.result.Type)
}

func (l *Literal) Leaf() bool {
	return true
}

func (l *Literal) String() string {
	return l.Token().Lexeme
}

func (l *Literal) Result() *ExprResult {
	return &l.result
}

func (l *Literal) SetChildrenParent() {
}

// StructInitializer

type StructInitializer struct {
	range_ core.Range
	parent Node
	result ExprResult

	Name   scanner.Token
	Fields []InitField
}

func (s *StructInitializer) Token() scanner.Token {
	return s.Name
}

func (s *StructInitializer) Range() core.Range {
	return s.range_
}

func (s *StructInitializer) SetRangeToken(start, end scanner.Token) {
	s.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (s *StructInitializer) SetRangePos(start, end core.Pos) {
	s.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (s *StructInitializer) SetRangeNode(start, end Node) {
	s.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (s *StructInitializer) Parent() Node {
	return s.parent
}

func (s *StructInitializer) SetParent(parent Node) {
	if s.parent != nil && parent != nil {
		log.Fatalln("StructInitializer.SetParent() - Node already has a parent")
	}
	s.parent = parent
}

func (s *StructInitializer) Accept(visitor ExprVisitor) {
	visitor.VisitStructInitializer(s)
}

func (s *StructInitializer) AcceptChildren(visitor Acceptor) {
	for i_ := range s.Fields {
		if s.Fields[i_].Value != nil {
			visitor.AcceptExpr(s.Fields[i_].Value)
		}
	}
}

func (s *StructInitializer) AcceptTypes(visitor types.Visitor) {
	if s.result.Type != nil {
		visitor.VisitType(s.result.Type)
	}
}

func (s *StructInitializer) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&s.result.Type)
}

func (s *StructInitializer) Leaf() bool {
	return false
}

func (s *StructInitializer) String() string {
	return s.Token().Lexeme
}

func (s *StructInitializer) Result() *ExprResult {
	return &s.result
}

func (s *StructInitializer) SetChildrenParent() {
	for i_ := range s.Fields {
		if s.Fields[i_].Value != nil {
			s.Fields[i_].Value.SetParent(s)
		}
	}
}

// InitField

type InitField struct {
	Name  scanner.Token
	Value Expr
}

// ArrayInitializer

type ArrayInitializer struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Values []Expr
}

func (a *ArrayInitializer) Token() scanner.Token {
	return a.Token_
}

func (a *ArrayInitializer) Range() core.Range {
	return a.range_
}

func (a *ArrayInitializer) SetRangeToken(start, end scanner.Token) {
	a.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (a *ArrayInitializer) SetRangePos(start, end core.Pos) {
	a.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (a *ArrayInitializer) SetRangeNode(start, end Node) {
	a.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (a *ArrayInitializer) Parent() Node {
	return a.parent
}

func (a *ArrayInitializer) SetParent(parent Node) {
	if a.parent != nil && parent != nil {
		log.Fatalln("ArrayInitializer.SetParent() - Node already has a parent")
	}
	a.parent = parent
}

func (a *ArrayInitializer) Accept(visitor ExprVisitor) {
	visitor.VisitArrayInitializer(a)
}

func (a *ArrayInitializer) AcceptChildren(visitor Acceptor) {
	for i_ := range a.Values {
		if a.Values[i_] != nil {
			visitor.AcceptExpr(a.Values[i_])
		}
	}
}

func (a *ArrayInitializer) AcceptTypes(visitor types.Visitor) {
	if a.result.Type != nil {
		visitor.VisitType(a.result.Type)
	}
}

func (a *ArrayInitializer) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&a.result.Type)
}

func (a *ArrayInitializer) Leaf() bool {
	return false
}

func (a *ArrayInitializer) String() string {
	return a.Token().Lexeme
}

func (a *ArrayInitializer) Result() *ExprResult {
	return &a.result
}

func (a *ArrayInitializer) SetChildrenParent() {
	for i_ := range a.Values {
		if a.Values[i_] != nil {
			a.Values[i_].SetParent(a)
		}
	}
}

// Unary

type Unary struct {
	range_ core.Range
	parent Node
	result ExprResult

	Op     scanner.Token
	Value  Expr
	Prefix bool
}

func (u *Unary) Token() scanner.Token {
	return u.Op
}

func (u *Unary) Range() core.Range {
	return u.range_
}

func (u *Unary) SetRangeToken(start, end scanner.Token) {
	u.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (u *Unary) SetRangePos(start, end core.Pos) {
	u.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (u *Unary) SetRangeNode(start, end Node) {
	u.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (u *Unary) Parent() Node {
	return u.parent
}

func (u *Unary) SetParent(parent Node) {
	if u.parent != nil && parent != nil {
		log.Fatalln("Unary.SetParent() - Node already has a parent")
	}
	u.parent = parent
}

func (u *Unary) Accept(visitor ExprVisitor) {
	visitor.VisitUnary(u)
}

func (u *Unary) AcceptChildren(visitor Acceptor) {
	if u.Value != nil {
		visitor.AcceptExpr(u.Value)
	}
}

func (u *Unary) AcceptTypes(visitor types.Visitor) {
	if u.result.Type != nil {
		visitor.VisitType(u.result.Type)
	}
}

func (u *Unary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&u.result.Type)
}

func (u *Unary) Leaf() bool {
	return false
}

func (u *Unary) String() string {
	return u.Token().Lexeme
}

func (u *Unary) Result() *ExprResult {
	return &u.result
}

func (u *Unary) SetChildrenParent() {
	if u.Value != nil {
		u.Value.SetParent(u)
	}
}

// Binary

type Binary struct {
	range_ core.Range
	parent Node
	result ExprResult

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (b *Binary) Token() scanner.Token {
	return b.Op
}

func (b *Binary) Range() core.Range {
	return b.range_
}

func (b *Binary) SetRangeToken(start, end scanner.Token) {
	b.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (b *Binary) SetRangePos(start, end core.Pos) {
	b.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (b *Binary) SetRangeNode(start, end Node) {
	b.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (b *Binary) Parent() Node {
	return b.parent
}

func (b *Binary) SetParent(parent Node) {
	if b.parent != nil && parent != nil {
		log.Fatalln("Binary.SetParent() - Node already has a parent")
	}
	b.parent = parent
}

func (b *Binary) Accept(visitor ExprVisitor) {
	visitor.VisitBinary(b)
}

func (b *Binary) AcceptChildren(visitor Acceptor) {
	if b.Left != nil {
		visitor.AcceptExpr(b.Left)
	}
	if b.Right != nil {
		visitor.AcceptExpr(b.Right)
	}
}

func (b *Binary) AcceptTypes(visitor types.Visitor) {
	if b.result.Type != nil {
		visitor.VisitType(b.result.Type)
	}
}

func (b *Binary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&b.result.Type)
}

func (b *Binary) Leaf() bool {
	return false
}

func (b *Binary) String() string {
	return b.Token().Lexeme
}

func (b *Binary) Result() *ExprResult {
	return &b.result
}

func (b *Binary) SetChildrenParent() {
	if b.Left != nil {
		b.Left.SetParent(b)
	}
	if b.Right != nil {
		b.Right.SetParent(b)
	}
}

// Logical

type Logical struct {
	range_ core.Range
	parent Node
	result ExprResult

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (l *Logical) Token() scanner.Token {
	return l.Op
}

func (l *Logical) Range() core.Range {
	return l.range_
}

func (l *Logical) SetRangeToken(start, end scanner.Token) {
	l.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (l *Logical) SetRangePos(start, end core.Pos) {
	l.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (l *Logical) SetRangeNode(start, end Node) {
	l.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (l *Logical) Parent() Node {
	return l.parent
}

func (l *Logical) SetParent(parent Node) {
	if l.parent != nil && parent != nil {
		log.Fatalln("Logical.SetParent() - Node already has a parent")
	}
	l.parent = parent
}

func (l *Logical) Accept(visitor ExprVisitor) {
	visitor.VisitLogical(l)
}

func (l *Logical) AcceptChildren(visitor Acceptor) {
	if l.Left != nil {
		visitor.AcceptExpr(l.Left)
	}
	if l.Right != nil {
		visitor.AcceptExpr(l.Right)
	}
}

func (l *Logical) AcceptTypes(visitor types.Visitor) {
	if l.result.Type != nil {
		visitor.VisitType(l.result.Type)
	}
}

func (l *Logical) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.result.Type)
}

func (l *Logical) Leaf() bool {
	return false
}

func (l *Logical) String() string {
	return l.Token().Lexeme
}

func (l *Logical) Result() *ExprResult {
	return &l.result
}

func (l *Logical) SetChildrenParent() {
	if l.Left != nil {
		l.Left.SetParent(l)
	}
	if l.Right != nil {
		l.Right.SetParent(l)
	}
}

// Identifier

type Identifier struct {
	range_ core.Range
	parent Node
	result ExprResult

	Identifier scanner.Token
	Kind       IdentifierKind
}

func (i *Identifier) Token() scanner.Token {
	return i.Identifier
}

func (i *Identifier) Range() core.Range {
	return i.range_
}

func (i *Identifier) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *Identifier) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *Identifier) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *Identifier) Parent() Node {
	return i.parent
}

func (i *Identifier) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("Identifier.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *Identifier) Accept(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) AcceptChildren(visitor Acceptor) {
}

func (i *Identifier) AcceptTypes(visitor types.Visitor) {
	if i.result.Type != nil {
		visitor.VisitType(i.result.Type)
	}
}

func (i *Identifier) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.result.Type)
}

func (i *Identifier) Leaf() bool {
	return true
}

func (i *Identifier) String() string {
	return i.Token().Lexeme
}

func (i *Identifier) Result() *ExprResult {
	return &i.result
}

func (i *Identifier) SetChildrenParent() {
}

// IdentifierKind

type IdentifierKind uint8

const (
	FunctionKind  IdentifierKind = 0
	EnumKind      IdentifierKind = 1
	VariableKind  IdentifierKind = 2
	ParameterKind IdentifierKind = 3
)

// Assignment

type Assignment struct {
	range_ core.Range
	parent Node
	result ExprResult

	Assignee Expr
	Op       scanner.Token
	Value    Expr
}

func (a *Assignment) Token() scanner.Token {
	return a.Op
}

func (a *Assignment) Range() core.Range {
	return a.range_
}

func (a *Assignment) SetRangeToken(start, end scanner.Token) {
	a.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (a *Assignment) SetRangePos(start, end core.Pos) {
	a.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (a *Assignment) SetRangeNode(start, end Node) {
	a.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (a *Assignment) Parent() Node {
	return a.parent
}

func (a *Assignment) SetParent(parent Node) {
	if a.parent != nil && parent != nil {
		log.Fatalln("Assignment.SetParent() - Node already has a parent")
	}
	a.parent = parent
}

func (a *Assignment) Accept(visitor ExprVisitor) {
	visitor.VisitAssignment(a)
}

func (a *Assignment) AcceptChildren(visitor Acceptor) {
	if a.Assignee != nil {
		visitor.AcceptExpr(a.Assignee)
	}
	if a.Value != nil {
		visitor.AcceptExpr(a.Value)
	}
}

func (a *Assignment) AcceptTypes(visitor types.Visitor) {
	if a.result.Type != nil {
		visitor.VisitType(a.result.Type)
	}
}

func (a *Assignment) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&a.result.Type)
}

func (a *Assignment) Leaf() bool {
	return false
}

func (a *Assignment) String() string {
	return a.Token().Lexeme
}

func (a *Assignment) Result() *ExprResult {
	return &a.result
}

func (a *Assignment) SetChildrenParent() {
	if a.Assignee != nil {
		a.Assignee.SetParent(a)
	}
	if a.Value != nil {
		a.Value.SetParent(a)
	}
}

// Cast

type Cast struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Expr   Expr
}

func (c *Cast) Token() scanner.Token {
	return c.Token_
}

func (c *Cast) Range() core.Range {
	return c.range_
}

func (c *Cast) SetRangeToken(start, end scanner.Token) {
	c.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (c *Cast) SetRangePos(start, end core.Pos) {
	c.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (c *Cast) SetRangeNode(start, end Node) {
	c.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Cast) Parent() Node {
	return c.parent
}

func (c *Cast) SetParent(parent Node) {
	if c.parent != nil && parent != nil {
		log.Fatalln("Cast.SetParent() - Node already has a parent")
	}
	c.parent = parent
}

func (c *Cast) Accept(visitor ExprVisitor) {
	visitor.VisitCast(c)
}

func (c *Cast) AcceptChildren(visitor Acceptor) {
	if c.Expr != nil {
		visitor.AcceptExpr(c.Expr)
	}
}

func (c *Cast) AcceptTypes(visitor types.Visitor) {
	if c.result.Type != nil {
		visitor.VisitType(c.result.Type)
	}
}

func (c *Cast) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.result.Type)
}

func (c *Cast) Leaf() bool {
	return false
}

func (c *Cast) String() string {
	return c.Token().Lexeme
}

func (c *Cast) Result() *ExprResult {
	return &c.result
}

func (c *Cast) SetChildrenParent() {
	if c.Expr != nil {
		c.Expr.SetParent(c)
	}
}

// Sizeof

type Sizeof struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Target types.Type
}

func (s *Sizeof) Token() scanner.Token {
	return s.Token_
}

func (s *Sizeof) Range() core.Range {
	return s.range_
}

func (s *Sizeof) SetRangeToken(start, end scanner.Token) {
	s.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (s *Sizeof) SetRangePos(start, end core.Pos) {
	s.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (s *Sizeof) SetRangeNode(start, end Node) {
	s.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (s *Sizeof) Parent() Node {
	return s.parent
}

func (s *Sizeof) SetParent(parent Node) {
	if s.parent != nil && parent != nil {
		log.Fatalln("Sizeof.SetParent() - Node already has a parent")
	}
	s.parent = parent
}

func (s *Sizeof) Accept(visitor ExprVisitor) {
	visitor.VisitSizeof(s)
}

func (s *Sizeof) AcceptChildren(visitor Acceptor) {
}

func (s *Sizeof) AcceptTypes(visitor types.Visitor) {
	if s.result.Type != nil {
		visitor.VisitType(s.result.Type)
	}
	if s.Target != nil {
		visitor.VisitType(s.Target)
	}
}

func (s *Sizeof) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&s.result.Type)
	visitor.VisitType(&s.Target)
}

func (s *Sizeof) Leaf() bool {
	return true
}

func (s *Sizeof) String() string {
	return s.Token().Lexeme
}

func (s *Sizeof) Result() *ExprResult {
	return &s.result
}

func (s *Sizeof) SetChildrenParent() {
}

// Call

type Call struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Callee Expr
	Args   []Expr
}

func (c *Call) Token() scanner.Token {
	return c.Token_
}

func (c *Call) Range() core.Range {
	return c.range_
}

func (c *Call) SetRangeToken(start, end scanner.Token) {
	c.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (c *Call) SetRangePos(start, end core.Pos) {
	c.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (c *Call) SetRangeNode(start, end Node) {
	c.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Call) Parent() Node {
	return c.parent
}

func (c *Call) SetParent(parent Node) {
	if c.parent != nil && parent != nil {
		log.Fatalln("Call.SetParent() - Node already has a parent")
	}
	c.parent = parent
}

func (c *Call) Accept(visitor ExprVisitor) {
	visitor.VisitCall(c)
}

func (c *Call) AcceptChildren(visitor Acceptor) {
	if c.Callee != nil {
		visitor.AcceptExpr(c.Callee)
	}
	for i_ := range c.Args {
		if c.Args[i_] != nil {
			visitor.AcceptExpr(c.Args[i_])
		}
	}
}

func (c *Call) AcceptTypes(visitor types.Visitor) {
	if c.result.Type != nil {
		visitor.VisitType(c.result.Type)
	}
}

func (c *Call) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.result.Type)
}

func (c *Call) Leaf() bool {
	return false
}

func (c *Call) String() string {
	return c.Token().Lexeme
}

func (c *Call) Result() *ExprResult {
	return &c.result
}

func (c *Call) SetChildrenParent() {
	if c.Callee != nil {
		c.Callee.SetParent(c)
	}
	for i_ := range c.Args {
		if c.Args[i_] != nil {
			c.Args[i_].SetParent(c)
		}
	}
}

// Index

type Index struct {
	range_ core.Range
	parent Node
	result ExprResult

	Token_ scanner.Token
	Value  Expr
	Index  Expr
}

func (i *Index) Token() scanner.Token {
	return i.Token_
}

func (i *Index) Range() core.Range {
	return i.range_
}

func (i *Index) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *Index) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *Index) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *Index) Parent() Node {
	return i.parent
}

func (i *Index) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("Index.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *Index) Accept(visitor ExprVisitor) {
	visitor.VisitIndex(i)
}

func (i *Index) AcceptChildren(visitor Acceptor) {
	if i.Value != nil {
		visitor.AcceptExpr(i.Value)
	}
	if i.Index != nil {
		visitor.AcceptExpr(i.Index)
	}
}

func (i *Index) AcceptTypes(visitor types.Visitor) {
	if i.result.Type != nil {
		visitor.VisitType(i.result.Type)
	}
}

func (i *Index) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.result.Type)
}

func (i *Index) Leaf() bool {
	return false
}

func (i *Index) String() string {
	return i.Token().Lexeme
}

func (i *Index) Result() *ExprResult {
	return &i.result
}

func (i *Index) SetChildrenParent() {
	if i.Value != nil {
		i.Value.SetParent(i)
	}
	if i.Index != nil {
		i.Index.SetParent(i)
	}
}

// Member

type Member struct {
	range_ core.Range
	parent Node
	result ExprResult

	Value Expr
	Name  scanner.Token
}

func (m *Member) Token() scanner.Token {
	return m.Name
}

func (m *Member) Range() core.Range {
	return m.range_
}

func (m *Member) SetRangeToken(start, end scanner.Token) {
	m.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (m *Member) SetRangePos(start, end core.Pos) {
	m.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (m *Member) SetRangeNode(start, end Node) {
	m.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (m *Member) Parent() Node {
	return m.parent
}

func (m *Member) SetParent(parent Node) {
	if m.parent != nil && parent != nil {
		log.Fatalln("Member.SetParent() - Node already has a parent")
	}
	m.parent = parent
}

func (m *Member) Accept(visitor ExprVisitor) {
	visitor.VisitMember(m)
}

func (m *Member) AcceptChildren(visitor Acceptor) {
	if m.Value != nil {
		visitor.AcceptExpr(m.Value)
	}
}

func (m *Member) AcceptTypes(visitor types.Visitor) {
	if m.result.Type != nil {
		visitor.VisitType(m.result.Type)
	}
}

func (m *Member) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&m.result.Type)
}

func (m *Member) Leaf() bool {
	return false
}

func (m *Member) String() string {
	return m.Token().Lexeme
}

func (m *Member) Result() *ExprResult {
	return &m.result
}

func (m *Member) SetChildrenParent() {
	if m.Value != nil {
		m.Value.SetParent(m)
	}
}
