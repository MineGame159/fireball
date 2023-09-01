package checker

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
)

type checker struct {
	scopes    []scope
	variables []variable

	structs map[string]*types.StructType

	functions core.Set[string]
	function  *ast.Func

	loopDepth int

	reporter core.Reporter
	decls    []ast.Decl
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  scanner.Token
	type_ types.Type

	param bool
	used  bool
}

func Check(reporter core.Reporter, decls []ast.Decl) {
	c := &checker{
		structs:   make(map[string]*types.StructType),
		functions: core.NewSet[string](),
		reporter:  reporter,
		decls:     decls,
	}

	// Collect structs
	for _, decl := range decls {
		if s, ok := decl.(*ast.Struct); ok {
			// Create type
			fields := make([]types.Field, len(s.Fields))

			for i, field := range s.Fields {
				fields[i] = types.Field{
					Name: field.Name.Lexeme,
					Type: field.Type,
				}
			}

			type_ := types.Struct(s.Name.Lexeme, fields, core.Range{})
			s.Type = type_

			// Save in map and check name collision
			if _, ok := c.structs[s.Name.Lexeme]; ok {
				c.errorToken(s.Name, "Struct with the name '%s' already exists.", s.Name)
			}

			c.structs[s.Name.Lexeme] = type_
		}
	}

	for _, decl := range decls {
		c.AcceptDecl(decl)
	}
}

func (c *checker) getFunction(name scanner.Token) *ast.Func {
	for _, decl := range c.decls {
		if function, ok := decl.(*ast.Func); ok {
			if function.Name.Lexeme == name.Lexeme {
				return function
			}
		}
	}

	return nil
}

// Scope / Variables

func (c *checker) getVariable(name scanner.Token) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.Lexeme == name.Lexeme {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *checker) addVariable(name scanner.Token, type_ types.Type) *variable {
	c.variables = append(c.variables, variable{
		name:  name,
		type_: type_,
	})

	c.peekScope().variableCount++
	return &c.variables[len(c.variables)-1]
}

func (c *checker) pushScope() {
	c.scopes = append(c.scopes, scope{
		variableI:     len(c.variables),
		variableCount: 0,
	})
}

func (c *checker) popScope() {
	// Check unused variables
	for i := len(c.variables) - 1; i >= c.peekScope().variableI; i-- {
		v := c.variables[i]

		if !v.used && v.name.Lexeme[0] != '_' && (!c.function.Extern || !v.param) {
			c.warningToken(v.name, "Unused variable '%s'. Prefix with '_' to ignore.", v.name)
		}
	}

	// Pop scope
	c.variables = c.variables[:c.peekScope().variableI]
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *checker) peekScope() *scope {
	return &c.scopes[len(c.scopes)-1]
}

// types.PtrVisitor

func (c *checker) VisitType(type_ *types.Type) {
	if v, ok := (*type_).(*types.UnresolvedType); ok {
		t, ok := c.structs[v.Identifier.Lexeme]

		if !ok {
			c.errorRange(v.Range(), "Unknown type '%s'.", v)
			*type_ = types.Primitive(types.Void, v.Range())
		} else {
			*type_ = types.Struct(t.Name, t.Fields, v.Range())
		}
	}

	if *type_ != nil {
		(*type_).AcceptChildrenPtr(c)
	}
}

// ast.Acceptor

func (c *checker) AcceptDecl(decl ast.Decl) {
	decl.Accept(c)
	decl.AcceptTypesPtr(c)
}

func (c *checker) AcceptStmt(stmt ast.Stmt) {
	stmt.Accept(c)
	stmt.AcceptTypesPtr(c)
}

func (c *checker) AcceptExpr(expr ast.Expr) {
	expr.Accept(c)
	expr.AcceptTypesPtr(c)
}

// Diagnostics

func (c *checker) errorRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) errorToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}
