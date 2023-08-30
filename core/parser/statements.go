package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) statement() (ast.Stmt, *core.Diagnostic) {
	if p.match(scanner.LeftBrace) {
		return p.block()
	}
	if p.match(scanner.Var) {
		return p.variable()
	}
	if p.match(scanner.If) {
		return p.if_()
	}
	if p.match(scanner.For) {
		return p.for_()
	}
	if p.match(scanner.Return) {
		return p.return_()
	}
	if p.match(scanner.Break) {
		return p.break_()
	}
	if p.match(scanner.Continue) {
		return p.continue_()
	}

	return p.expressionStmt()
}

func (p *parser) block() (ast.Stmt, *core.Diagnostic) {
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

func (p *parser) expressionStmt() (ast.Stmt, *core.Diagnostic) {
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

func (p *parser) variable() (ast.Stmt, *core.Diagnostic) {
	// Name
	name, err := p.consume(scanner.Identifier, "expected variable name")
	if err != nil {
		return nil, err
	}

	// Type
	var type_ types.Type

	if !p.check(scanner.Equal) {
		type__, err := p.parseType()
		if err != nil {
			return nil, err
		}

		type_ = type__
	}

	// Initializer
	var initializer ast.Expr

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

func (p *parser) if_() (ast.Stmt, *core.Diagnostic) {
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
	var else_ ast.Stmt

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

func (p *parser) for_() (ast.Stmt, *core.Diagnostic) {
	token := p.current

	// Condition
	var condition ast.Expr

	if !p.check(scanner.LeftBrace) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		condition = expr
	}

	// Body
	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// Return
	return &ast.For{
		Token_:    token,
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *parser) return_() (ast.Stmt, *core.Diagnostic) {
	token := p.current

	// Value
	var expr ast.Expr

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

	// Return
	return &ast.Return{
		Token_: token,
		Expr:   expr,
	}, nil
}

func (p *parser) break_() (ast.Stmt, *core.Diagnostic) {
	token := p.current

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	return &ast.Break{
		Token_: token,
	}, nil
}

func (p *parser) continue_() (ast.Stmt, *core.Diagnostic) {
	token := p.current

	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	return &ast.Continue{
		Token_: token,
	}, nil
}
