package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) statement() (ast.Stmt, *core.Error) {
	if p.match(scanner.LeftBrace) {
		return p.block()
	}
	if p.match(scanner.Var) {
		return p.variable()
	}
	if p.match(scanner.If) {
		return p.if_()
	}
	if p.match(scanner.Return) {
		return p.return_()
	}

	return p.expressionStmt()
}

func (p *parser) block() (ast.Stmt, *core.Error) {
	token := p.current
	stmts := make([]ast.Stmt, 0, 4)

	for !p.check(scanner.RightBrace) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	if _, err := p.consume(scanner.RightBrace, "Expected '}'."); err != nil {
		return nil, err
	}

	return &ast.Block{
		Token_: token,
		Stmts:  stmts,
	}, nil
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
	var initializer ast.Expr = nil

	if !p.check(scanner.Semicolon) {
		if _, err := p.consume(scanner.Equal, "Expected '='."); err != nil {
			return nil, err
		}

		if !p.check(scanner.Semicolon) {
			initializer, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
	}

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	return &ast.Variable{
		Type:        type_,
		Name:        name,
		Initializer: initializer,
	}, nil
}

func (p *parser) if_() (ast.Stmt, *core.Error) {
	token := p.current

	// Condition
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Then
	then, err := p.statement()
	if err != nil {
		return nil, err
	}

	// Else
	var else_ ast.Stmt = nil

	if p.match(scanner.Else) {
		else__, err := p.statement()
		if err != nil {
			return nil, err
		}

		else_ = else__
	}

	// Return
	return &ast.If{
		Token_:    token,
		Condition: condition,
		Then:      then,
		Else:      else_,
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
