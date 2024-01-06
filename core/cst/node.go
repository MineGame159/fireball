package cst

import (
	"fireball/core"
	"fireball/core/scanner"
	"slices"
)

type NodeKind uint8

type Node struct {
	Kind  NodeKind
	Range core.Range

	Token    scanner.Token
	Children []Node
}

func (n Node) Leaf() bool {
	return n.Token.Lexeme != ""
}

func (n Node) Count(kind scanner.TokenKind) int {
	count := 0

	for _, child := range n.Children {
		if child.Token.Kind == kind {
			count++
		}
	}

	return count
}

func (n Node) Contains(kind scanner.TokenKind) bool {
	for _, child := range n.Children {
		if child.Token.Kind == kind {
			return true
		}
	}

	return false
}

func (n Node) ContainsAny(kinds []scanner.TokenKind) bool {
	for _, child := range n.Children {
		if slices.Contains(kinds, child.Token.Kind) {
			return true
		}
	}

	return false
}

func (n Node) Get(kind scanner.TokenKind) *Node {
	for i := range n.Children {
		child := &n.Children[i]

		if child.Token.Kind == kind {
			return child
		}
	}

	return nil
}

const (
	UnknownNode NodeKind = iota
	FileNode

	IdentifierTypeNode
	PointerTypeNode
	ArrayTypeNode
	FuncTypeNode
	FuncTypeParamNode

	StructNode
	StructFieldNode
	ImplNode
	EnumNode
	EnumCaseNode
	FuncNode
	FuncParamNode

	ExprStmtNode
	BlockStmtNode
	VarStmtNode
	IfStmtNode
	WhileStmtNode
	ForStmtNode
	ReturnStmtNode
	BreakStmtNode
	ContinueStmtNode

	ParenExprNode
	UnaryExprNode
	BinaryExprNode
	IndexExprNode
	CallExprNode
	TypeCallExprNode
	StructExprNode
	StructFieldExprNode
	ArrayExprNode
	AllocateArrayExprNode
	NilExprNode
	BoolExprNode
	NumberExprNode
	CharacterExprNode
	StringExprNode

	AttributesNode
	AttributeNode

	KeywordNode
	IdentifierNode
	CommentNode
	MiscNode
)

func NodeKindFromToken(token scanner.Token) NodeKind {
	switch token.Kind {
	case scanner.Identifier:
		return IdentifierNode

	case scanner.Nil:
		return NilExprNode
	case scanner.True, scanner.False:
		return BoolExprNode
	case scanner.Number, scanner.Hex, scanner.Binary:
		return NumberExprNode
	case scanner.Character:
		return CharacterExprNode
	case scanner.String:
		return StringExprNode

	case scanner.Comment:
		return CommentNode

	default:
		if scanner.IsKeyword(token.Kind) {
			return KeywordNode
		}

		return MiscNode
	}
}

func (n NodeKind) IsType() bool {
	return n >= IdentifierTypeNode && n <= FuncTypeNode
}

func (n NodeKind) IsDecl() bool {
	switch n {
	case StructNode, ImplNode, EnumNode, FuncNode:
		return true

	default:
		return false
	}
}

func (n NodeKind) IsStmt() bool {
	return n >= ExprStmtNode && n <= ContinueStmtNode
}

func (n NodeKind) IsExpr() bool {
	return (n >= ParenExprNode && n <= StringExprNode) || n == IdentifierNode
}

func (n NodeKind) String() string {
	switch n {
	case UnknownNode:
		return "Unknown"
	case FileNode:
		return "File"

	case IdentifierTypeNode:
		return "Identifier type"
	case PointerTypeNode:
		return "Pointer type"
	case ArrayTypeNode:
		return "Array type"
	case FuncTypeNode:
		return "Function type"
	case FuncTypeParamNode:
		return "Function type param"

	case StructNode:
		return "Struct"
	case StructFieldNode:
		return "Struct field"
	case ImplNode:
		return "Impl"
	case EnumNode:
		return "Enum"
	case EnumCaseNode:
		return "Enum case"
	case FuncNode:
		return "Func"
	case FuncParamNode:
		return "Func param"

	case ExprStmtNode:
		return "Expression;"
	case BlockStmtNode:
		return "Block"
	case VarStmtNode:
		return "Variable"
	case WhileStmtNode:
		return "While"
	case IfStmtNode:
		return "if"
	case ForStmtNode:
		return "For"
	case ReturnStmtNode:
		return "Return"
	case BreakStmtNode:
		return "Break"
	case ContinueStmtNode:
		return "Continue"

	case ParenExprNode:
		return "Paren"
	case UnaryExprNode:
		return "Unary"
	case BinaryExprNode:
		return "Binary"
	case IndexExprNode:
		return "Index"
	case CallExprNode:
		return "Call"
	case TypeCallExprNode:
		return "Type call"
	case StructExprNode:
		return "Struct"
	case StructFieldExprNode:
		return "Struct field"
	case ArrayExprNode:
		return "Array"
	case AllocateArrayExprNode:
		return "Allocate array"
	case NilExprNode:
		return "Nil"
	case BoolExprNode:
		return "Bool"
	case NumberExprNode:
		return "Number"
	case CharacterExprNode:
		return "Character"
	case StringExprNode:
		return "String"

	case AttributesNode:
		return "Attributes"
	case AttributeNode:
		return "Attribute"

	case KeywordNode:
		return "Keyword"
	case IdentifierNode:
		return "Identifier"
	case CommentNode:
		return "Comment"
	case MiscNode:
		return "Misc"

	default:
		panic("NodeKind.String() - Invalid value")
	}
}
