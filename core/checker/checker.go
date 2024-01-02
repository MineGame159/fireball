package checker

import (
	"fireball/core/ast"
	"fireball/core/fuckoff"
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
	resolver fuckoff.Resolver
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  *ast.Token
	type_ ast.Type

	param bool
	used  bool
}

func Check(reporter utils.Reporter, resolver fuckoff.Resolver, node ast.Node) {
	c := &checker{
		reporter: reporter,
		resolver: resolver,
	}

	reset(c, node)
	c.VisitNode(node)
}

// Scope / Variables

func (c *checker) hasVariableInScope(name *ast.Token) bool {
	scope := c.peekScope()

	for i := scope.variableCount - 1; i >= 0; i-- {
		if c.variables[scope.variableI+i].name.String() == name.String() {
			return true
		}
	}

	return false
}

func (c *checker) getVariable(name ast.Node) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.String() == name.String() {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *checker) addVariable(name *ast.Token, type_ ast.Type) *variable {
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

			if !v.used && v.name.String()[0] != '_' && (c.function.HasBody() || !v.param) {
				c.warning(v.name, "Unused variable '%s'. Prefix with '_' to ignore.", v.name)
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

// ast.Visit

func (c *checker) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case ast.Decl:
		node.AcceptDecl(c)

	case ast.Stmt:
		node.AcceptStmt(c)

	case ast.Expr:
		node.AcceptExpr(c)

	default:
		node.AcceptChildren(c)
	}
}

// Diagnostics

func (c *checker) error(node ast.Node, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Cst().Range,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warning(node ast.Node, format string, args ...any) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.WarningKind,
		Range:   node.Cst().Range,
		Message: fmt.Sprintf(format, args...),
	})
}
