package ast

import (
	"fireball/core"
	"fireball/core/cst"
	"fireball/core/scanner"
)

type Node interface {
	Cst() *cst.Node
	Token() scanner.Token

	Parent() Node
	SetParent(parent Node)

	AcceptChildren(visitor Visitor)

	String() string
}

type Visitor interface {
	VisitNode(node Node)
}

// GetLeaf()

func GetLeaf(node Node, pos core.Pos) Node {
	g := get{
		pos: pos,
	}

	g.VisitNode(node)
	return g.node
}

type get struct {
	pos  core.Pos
	node Node
}

func (g *get) VisitNode(node Node) {
	// Propagate node up the tree
	if g.node != nil {
		return
	}

	// Check if node contains target position
	if node.Cst() == nil || node.Cst().Range.Contains(g.pos) {
		// Check if node is a leaf node
		if !node.Token().IsError() {
			g.node = node
			return
		}

		// Children
		node.AcceptChildren(g)

		// Propagate node up the tree
		if g.node != nil {
			return
		}
	}
}

// GetParent()

func GetParent[T Node](node Node) T {
	for node != nil {
		if n, ok := node.(T); ok {
			return n
		}

		node = node.Parent()
	}

	var empty T
	return empty
}
