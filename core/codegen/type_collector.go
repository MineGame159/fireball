package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

type typeValuePair struct {
	type_ types.Type
	val   value
}

type typeCollector struct {
	globals *values
	types   []typeValuePair
}

func collectTypes(globals *values, decls []ast.Decl) []typeValuePair {
	t := &typeCollector{
		globals: globals,
		types:   make([]typeValuePair, 0, 64),
	}

	for _, decl := range decls {
		t.AcceptDecl(decl)
	}

	return t.types
}

// types.Visitor

func (t *typeCollector) addType(type_ types.Type) {
	// Try cache
	for _, pair := range t.types {
		if pair.type_.Equals(type_) {
			return
		}
	}

	// Struct
	if v, ok := type_.(*ast.Struct); ok {
		val := t.globals.constant("%struct."+v.Name.Lexeme, v)

		t.types = append(t.types, typeValuePair{
			type_: type_,
			val:   val,
		})
	}
}

func (t *typeCollector) VisitType(type_ types.Type) {
	t.addType(type_)
	type_.AcceptTypes(t)
}

// ast.Acceptor

func (t *typeCollector) AcceptDecl(decl ast.Decl) {
	decl.AcceptTypes(t)
	decl.AcceptChildren(t)
}

func (t *typeCollector) AcceptStmt(stmt ast.Stmt) {
	stmt.AcceptTypes(t)
	stmt.AcceptChildren(t)
}

func (t *typeCollector) AcceptExpr(expr ast.Expr) {
	expr.AcceptTypes(t)
	expr.AcceptChildren(t)
}
