package lsp

import (
	"cmp"
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"fmt"
	"slices"
)

type highlighter struct {
	enums utils.Set[string]

	functions utils.Set[string]
	params    []ast.Param

	tokens []semantic
}

func highlight(decls []ast.Decl) []uint32 {
	h := &highlighter{
		enums:     utils.NewSet[string](),
		functions: utils.NewSet[string](),
		tokens:    make([]semantic, 0, 256),
	}

	for _, decl := range decls {
		if v, ok := decl.(*ast.Enum); ok {
			h.enums.Add(v.Name.Lexeme)
		} else if v, ok := decl.(*ast.Func); ok {
			h.functions.Add(v.Name.Lexeme)
		}
	}

	for _, decl := range decls {
		h.AcceptDecl(decl)
	}

	return h.data()
}

// Declarations

func (h *highlighter) VisitStruct(decl *ast.Struct) {
	h.addToken(decl.Name, classKind)

	for _, field := range decl.Fields {
		h.addToken(field.Name, propertyKind)
	}

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitEnum(decl *ast.Enum) {
	h.addToken(decl.Name, enumKind)

	for _, case_ := range decl.Cases {
		h.addToken(case_.Name, enumMemberKind)
	}

	decl.AcceptChildren(h)
}

func (h *highlighter) VisitFunc(decl *ast.Func) {
	h.addToken(decl.Name, functionKind)

	for _, param := range decl.Params {
		h.addToken(param.Name, parameterKind)
	}

	h.params = decl.Params
	decl.AcceptChildren(h)
	h.params = nil
}

// Statements

func (h *highlighter) VisitBlock(stmt *ast.Block) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitExpression(stmt *ast.Expression) {
	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitVariable(stmt *ast.Variable) {
	h.addToken(stmt.Name, variableKind)

	stmt.AcceptChildren(h)
}

func (h *highlighter) VisitIf(stmt *ast.If) {
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

func (h *highlighter) VisitGroup(expr *ast.Group) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitInitializer(expr *ast.Initializer) {
	h.addToken(expr.Name, classKind)

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
	var kind semanticKind

	switch expr.Kind {
	case ast.FunctionKind:
		kind = functionKind

	case ast.EnumKind:
		kind = enumKind

	case ast.VariableKind:
		kind = variableKind

	case ast.ParameterKind:
		kind = parameterKind
	}

	h.addToken(expr.Identifier, kind)

	expr.AcceptChildren(h)
}

func (h *highlighter) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(h)
}

func (h *highlighter) VisitMember(expr *ast.Member) {
	if i, ok := expr.Value.(*ast.Identifier); ok && i.Kind == ast.EnumKind {
		h.addToken(expr.Name, enumMemberKind)
	} else {
		h.addToken(expr.Name, propertyKind)
	}

	expr.AcceptChildren(h)
}

// types.Visitor

func (h *highlighter) VisitType(type_ types.Type) {
	if type_ == nil {
		fmt.Println()
	}

	if _, ok := type_.(ast.Decl); !ok && type_.Range().Valid() {
		if _, ok := type_.(*types.PrimitiveType); ok {
			h.addRange(type_.Range(), typeKind)
		} else if _, ok := type_.(*ast.Struct); ok {
			h.addRange(type_.Range(), classKind)
		} else if _, ok := type_.(*ast.Enum); ok {
			h.addRange(type_.Range(), enumKind)
		} else {
			type_.AcceptTypes(h)
		}
	}
}

// ast.Acceptor

func (h *highlighter) AcceptDecl(decl ast.Decl) {
	decl.Accept(h)
	decl.AcceptTypes(h)
}

func (h *highlighter) AcceptStmt(stmt ast.Stmt) {
	stmt.Accept(h)
	stmt.AcceptTypes(h)
}

func (h *highlighter) AcceptExpr(expr ast.Expr) {
	expr.Accept(h)
	expr.AcceptTypes(h)
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
)

type semantic struct {
	line   uint16
	column uint8

	// length - 0b00011111
	// kind   - 0b11100000
	lengthKind uint8
}

func newSemantic(line, column, length int, kind semanticKind) semantic {
	length = min(length, 31)

	return semantic{
		line:       uint16(line - 1),
		column:     uint8(column),
		lengthKind: (uint8(length) & 0b00011111) | ((uint8(kind) << 5) & 0b11100000),
	}
}

func (h *highlighter) addToken(token scanner.Token, kind semanticKind) {
	if token.Column < 256 {
		h.tokens = append(h.tokens, newSemantic(token.Line, token.Column, len(token.Lexeme), kind))
	}
}

func (h *highlighter) addRange(range_ core.Range, kind semanticKind) {
	if range_.Start.Line == range_.End.Line {
		if range_.Start.Column < 256 {
			h.tokens = append(h.tokens, newSemantic(range_.Start.Line, range_.Start.Column, range_.End.Column-range_.Start.Column, kind))
		}
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
		data[j+2] = uint32(token.lengthKind & 0b00011111)
		data[j+3] = uint32((token.lengthKind & 0b11100000) >> 5)
		data[j+4] = uint32(0)

		lastLine = token.line
		lastColumn = token.column
	}

	return data
}

// Utils

func (h *highlighter) isParameter(name scanner.Token) bool {
	for _, param := range h.params {
		if param.Name.Lexeme == name.Lexeme {
			return true
		}
	}

	return false
}
