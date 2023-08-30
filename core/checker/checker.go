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

	structs core.Set[string]

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
		structs:   core.NewSet[string](),
		functions: core.NewSet[string](),
		reporter:  reporter,
		decls:     decls,
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

// ast.Acceptor

func (c *checker) AcceptDecl(decl ast.Decl) {
	decl.Accept(c)
}

func (c *checker) AcceptStmt(stmt ast.Stmt) {
	stmt.Accept(c)
}

func (c *checker) AcceptExpr(expr ast.Expr) {
	expr.Accept(c)
}

// Diagnostics

func (c *checker) errorRange(range_ ast.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) errorNode(node ast.Node, format string, args ...any) {
	c.errorToken(node.Token(), format, args...)
}

func (c *checker) errorToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   ast.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningRange(range_ ast.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningNode(node ast.Node, format string, args ...any) {
	c.warningToken(node.Token(), format, args...)
}

func (c *checker) warningToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   ast.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}
