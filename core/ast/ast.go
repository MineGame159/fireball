package ast

import (
	"fireball/core"
	"fireball/core/scanner"
	"fireball/core/types"
)

type Node interface {
	Token() scanner.Token
	Range() core.Range

	Parent() Node
	SetParent(parent Node)

	AcceptChildren(acceptor Acceptor)

	AcceptTypes(visitor types.Visitor)
	AcceptTypesPtr(visitor types.PtrVisitor)

	Leaf() bool
	String() string
}

type Acceptor interface {
	AcceptDecl(decl Decl)
	AcceptStmt(stmt Stmt)
	AcceptExpr(expr Expr)
}

// Utils

func IsIdentifierKindVariable(kind IdentifierKind) bool {
	return kind == VariableKind || kind == ParameterKind
}

// Visit

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

// GetLeaf

func GetLeaf(node Node, pos core.Pos) Node {
	g := get{
		pos: pos,
	}

	g.Accept(node)
	return g.node
}

type get struct {
	pos  core.Pos
	node Node
}

func (v *get) Accept(node Node) {
	// Propagate node up the tree
	if v.node != nil {
		return
	}

	// Check if node contains target position
	if node.Range().Contains(v.pos) {
		// Check if node is a leaf node
		if node.Leaf() {
			v.node = node
			return
		}

		// Children
		node.AcceptChildren(v)

		// Propagate node up the tree
		if v.node != nil {
			return
		}

		// TODO: Work around the fact that some nodes store names as scanner.Token which does not inherit ast.Node
		if initializer, ok := node.(*StructInitializer); ok {
			for _, field := range initializer.Fields {
				if core.TokenToRange(field.Name).Contains(v.pos) {
					v.node = node
					return
				}
			}
		}

		if variable, ok := node.(*Variable); ok && core.TokenToRange(variable.Name).Contains(v.pos) {
			v.node = node
			return
		}

		if member, ok := node.(*Member); ok && core.TokenToRange(member.Name).Contains(v.pos) {
			v.node = node
			return
		}
	}
}

func (v *get) AcceptDecl(decl Decl) {
	v.Accept(decl)
}

func (v *get) AcceptStmt(stmt Stmt) {
	v.Accept(stmt)
}

func (v *get) AcceptExpr(expr Expr) {
	v.Accept(expr)
}
