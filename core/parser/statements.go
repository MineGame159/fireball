package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) statement() (ast.Stmt, *core.Error) {
	if p.match(scanner.Var) {
		return p.variable()
	}
	if p.match(scanner.Return) {
		return p.return_()
	}

	return p.expressionStmt()
}

func (p *parser) expressionStmt() (ast.Stmt, *core.Error) {
	token := p.next

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, isAssignment := expr.(*ast.Assignment)
	_, isCall := expr.(*ast.Call)

	if !isAssignment && !isCall {
		return nil, p.error(token, "Invalid statement.")
	}

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	return &ast.Expression{
		Token_: token,
		Expr:   expr,
	}, nil
}

func (p *parser) variable() (ast.Stmt, *core.Error) {
	// Name
	name, err := p.consume(scanner.Identifier, "expected variable name")
	if err != nil {
		return nil, err
	}

	// Type
	var type_ types.Type = nil

	if !p.check(scanner.Equal) {
		type__, err := p.parseType()
		if err != nil {
			return nil, err
		}

		type_ = type__
	}

	// Initializer
	if _, err := p.consume(scanner.Equal, "Expected '='."); err != nil {
		return nil, err
	}

	var expr ast.Expr = nil

	if !p.check(scanner.Semicolon) {
		expr, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	return &ast.Variable{
		Type:        type_,
		Name:        name,
		Initializer: expr,
	}, nil
}

func (p *parser) return_() (ast.Stmt, *core.Error) {
	token := p.current
	var expr ast.Expr = nil

	if !p.check(scanner.Semicolon) {
		expr_, err := p.expression()
		if err != nil {
			return nil, err
		}

		expr = expr_
	}

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	return &ast.Return{
		Token_: token,
		Expr:   expr,
	}, nil
}
