package checker

import (
	"fireball/core/ast"
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
	resolver *ast.CombinedResolver
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  *ast.Token
	type_ ast.Type

	node ast.Node

	param bool
	used  bool
}

func Check(reporter utils.Reporter, root ast.RootResolver, file *ast.File) {
	c := &checker{
		reporter: reporter,
		resolver: ast.NewCombinedResolver(root),
	}

	for _, decl := range file.Decls {
		if using, ok := decl.(*ast.Using); ok {
			if resolver := root.GetResolver(using.Name); resolver != nil {
				c.resolver.Add(resolver)
			} else {
				c.error(using.Name, "Unknown namespace")
			}
		}
	}

	reset(c, file)
	c.VisitNode(file)
}

// Scope / Variables

func (c *checker) hasVariableInScope(name *ast.Token) bool {
	if name != nil {
		scope := c.peekScope()

		for i := scope.variableCount - 1; i >= 0; i-- {
			if c.variables[scope.variableI+i].name.String() == name.String() {
				return true
			}
		}
	}

	return false
}

func (c *checker) getVariable(name string) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.String() == name {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *checker) addVariable(name *ast.Token, type_ ast.Type, node ast.Node) *variable {
	if name == nil || type_ == nil {
		return nil
	}

	c.variables = append(c.variables, variable{
		name:  name,
		type_: type_,
		node:  node,
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

// Other

func (c *checker) expectPrimitiveValue(expr ast.Expr, kind ast.PrimitiveKind) {
	if expr == nil {
		return
	}

	if expr.Result().Kind == ast.InvalidResultKind {
		// Do not cascade errors
		return
	}

	if expr.Result().Kind != ast.ValueResultKind {
		c.error(expr, "Invalid value")
		return
	}

	if !ast.IsPrimitive(expr.Result().Type, kind) {
		c.error(expr, "Expected a '%s' but got a '%s'", kind.String(), ast.PrintType(expr.Result().Type))
	}
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
	if ast.IsNil(node) {
		return
	}

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
