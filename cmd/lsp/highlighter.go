package lsp

import (
	"cmp"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
	"slices"
)

type highlighter struct {
	enums utils.Set[string]

	functions utils.Set[string]
	params    []*ast.Param

	tokens []semantic
}

func highlight(node ast.Node) []uint32 {
	h := highlighter{
		enums:     utils.NewSet[string](),
		functions: utils.NewSet[string](),
		tokens:    make([]semantic, 0, 256),
	}

	for _, decl := range node.(*ast.File).Decls {
		if v, ok := decl.(*ast.Enum); ok {
			if v.Name != nil {
				h.enums.Add(v.Name.String())
			}
		} else if v, ok := decl.(*ast.Func); ok {
			if v.Name != nil {
				h.functions.Add(v.Name.String())
			}
		}
	}

	h.VisitNode(node)

	return h.data()
}

// Declarations

func (h *highlighter) VisitNamespace(decl *ast.Namespace) {
	if decl.Name != nil {
		for _, part := range decl.Name.Parts {
			h.add(part, namespaceKind)
		}
	}
}

func (h *highlighter) VisitUsing(decl *ast.Using) {
	if decl.Name != nil {
		for _, part := range decl.Name.Parts {
			h.add(part, namespaceKind)
		}
	}
}

func (h *highlighter) VisitStruct(decl *ast.Struct) {
	h.add(decl.Name, classKind)

	for _, field := range decl.Fields {
		h.add(field.Name, propertyKind)
	}

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitImpl(decl *ast.Impl) {
	h.add(decl.Struct, classKind)

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitEnum(decl *ast.Enum) {
	h.add(decl.Name, enumKind)

	for _, case_ := range decl.Cases {
		h.add(case_.Name, enumMemberKind)
	}

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitFunc(decl *ast.Func) {
	if decl.Name != nil {
		h.add(decl.Name, functionKind)
	}

	for _, param := range decl.Params {
		h.add(param.Name, parameterKind)
	}

	h.params = decl.Params
	decl.AcceptChildren(h)
	h.params = nil
}

func (h *highlighter) VisitGlobalVar(decl *ast.GlobalVar) {
	h.add(decl.Name, variableKind)

	decl.AcceptChildren(h)
}

// Statements

func (h *highlighter) VisitBlock(stmt *ast.Block) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitVar(stmt *ast.Var) {
	h.add(stmt.Name, variableKind)

	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitIf(stmt *ast.If) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitWhile(stmt *ast.While) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitFor(stmt *ast.For) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitReturn(stmt *ast.Return) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitBreak(stmt *ast.Break) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitContinue(stmt *ast.Continue) {
	stmt.AcceptChildren(h)
}

// Expressions

func (h *highlighter) VisitParen(expr *ast.Paren) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitStructInitializer(expr *ast.StructInitializer) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitAllocateArray(expr *ast.AllocateArray) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitIdentifier(expr *ast.Identifier) {
	h.visitExprResult(expr, expr.Result())

	expr.AcceptChildren(h)
}

func (h *highlighter) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitTypeCall(expr *ast.TypeCall) {
	h.add(expr.Callee, functionKind)

	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitMember(expr *ast.Member) {
	h.visitExprResult(expr.Name, expr.Result())

	expr.AcceptChildren(h)
}

func (h *highlighter) visitExprResult(node ast.Node, result *ast.ExprResult) {
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch result.Kind {
	case ast.TypeResultKind:
		switch result.Type.(type) {
		case *ast.Struct:
			h.add(node, classKind)
		case *ast.Enum:
			h.add(node, enumKind)
		}

	case ast.ResolverResultKind:
		h.add(node, namespaceKind)

	case ast.ValueResultKind:
		switch result.Value().(type) {
		case *ast.Field:
			h.add(node, propertyKind)
		case *ast.Param:
			h.add(node, parameterKind)
		case *ast.Var:
			h.add(node, variableKind)
		case *ast.EnumCase:
			h.add(node, enumMemberKind)
		}

	case ast.CallableResultKind:
		h.add(node, functionKind)
	}
}

// Types

func (h *highlighter) visitType(type_ ast.Type) {
	if type_.Cst() != nil {
		switch type_ := type_.(type) {
		case *ast.Primitive:
			h.add(type_, typeKind)

		case *ast.Resolvable:
			switch type_.Resolved().(type) {
			case *ast.Struct:
				h.add(type_, classKind)
			case *ast.Enum:
				h.add(type_, enumKind)
			}
		}
	}

	type_.AcceptChildren(h)
}

// ast.Visitor

func (h *highlighter) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case ast.Decl:
		node.AcceptDecl(h)
	case ast.Stmt:
		node.AcceptStmt(h)
	case ast.Expr:
		node.AcceptExpr(h)

	case ast.Type:
		h.visitType(node)

	default:
		node.AcceptChildren(h)
	}
}

// Tokens

type semanticKind uint8

const (
	functionKind semanticKind = iota
	parameterKind
	variableKind
	typeKind
	classKind
	enumKind
	propertyKind
	enumMemberKind
	namespaceKind
)

type semantic struct {
	line   uint16
	column uint8

	length uint8
	kind   semanticKind
}

func newSemantic(line, column, length uint16, kind semanticKind) semantic {
	return semantic{
		line:   line - 1,
		column: uint8(column),
		length: uint8(length),
		kind:   kind,
	}
}

func (h *highlighter) add(node ast.Node, kind semanticKind) {
	if !ast.IsNil(node) && node.Cst().Range.Start.Column < 256 {
		range_ := node.Cst().Range
		h.tokens = append(h.tokens, newSemantic(range_.Start.Line, range_.Start.Column, range_.End.Column-range_.Start.Column, kind))
	}
}

func (h *highlighter) data() []uint32 {
	// Sort tokens
	slices.SortFunc(h.tokens, func(a, b semantic) int {
		if a.line == b.line {
			return cmp.Compare(a.column, b.column)
		}

		if a.line < b.line {
			return -1
		}

		return 1
	})

	// Get data
	data := make([]uint32, len(h.tokens)*5)

	lastLine := uint16(0)
	lastColumn := uint8(0)

	for i, token := range h.tokens {
		if lastLine != token.line {
			lastColumn = 0
		}

		j := i * 5

		data[j+0] = uint32(token.line - lastLine)
		data[j+1] = uint32(token.column - lastColumn)
		data[j+2] = uint32(token.length)
		data[j+3] = uint32(token.kind)
		data[j+4] = uint32(0)

		lastLine = token.line
		lastColumn = token.column
	}

	return data
}

// Utils

func (h *highlighter) isParameter(name scanner.Token) bool {
	for _, param := range h.params {
		if param.Name.String() == name.Lexeme {
			return true
		}
	}

	return false
}
