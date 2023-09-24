package checker

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"fmt"
)

type checker struct {
	scopes    []scope
	variables []variable

	function *ast.Func

	loopDepth int

	typeExpr ast.Expr

	reporter utils.Reporter
	resolver utils.Resolver
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

func Check(reporter utils.Reporter, resolver utils.Resolver, decls []ast.Decl) {
	c := &checker{
		reporter: reporter,
		resolver: resolver,
		decls:    decls,
	}

	resolveTypes(c, decls)

	for _, decl := range decls {
		c.AcceptDecl(decl)
	}
}

// Scope / Variables

func (c *checker) hasVariableInScope(name scanner.Token) bool {
	scope := c.peekScope()

	for i := scope.variableCount - 1; i >= 0; i-- {
		if c.variables[scope.variableI+i].name.Lexeme == name.Lexeme {
			return true
		}
	}

	return false
}

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
	if c.function != nil {
		for i := len(c.variables) - 1; i >= c.peekScope().variableI; i-- {
			v := c.variables[i]

			if !v.used && v.name.Lexeme[0] != '_' && (!c.function.IsExtern() || !v.param) {
				c.warningToken(v.name, "Unused variable '%s'. Prefix with '_' to ignore.", v.name)
			}
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

func (c *checker) errorRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) errorToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.WarningKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.WarningKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}
