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

	function *ast.Func

	reporter core.ErrorReporter
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
}

func Check(reporter core.ErrorReporter, decls []ast.Decl) {
	c := &checker{
		reporter: reporter,
		decls:    decls,
	}

	for _, decl := range decls {
		c.acceptDecl(decl)
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
	c.variables = c.variables[:c.peekScope().variableI]
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *checker) peekScope() *scope {
	return &c.scopes[len(c.scopes)-1]
}

// Accept

func (c *checker) acceptDecl(decl ast.Decl) {
	if decl != nil {
		decl.Accept(c)
	}
}

func (c *checker) acceptStmt(stmt ast.Stmt) {
	if stmt != nil {
		stmt.Accept(c)
	}
}

func (c *checker) acceptExpr(expr ast.Expr) {
	if expr != nil {
		expr.Accept(c)
	}
}

// Errors

func (c *checker) error(node ast.Node, format string, args ...any) {
	token := node.Token()

	c.reporter.Report(core.Error{
		Message: fmt.Sprintf(format, args...),
		Line:    token.Line,
		Column:  token.Column,
	})
}
