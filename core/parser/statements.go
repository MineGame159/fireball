package parser

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
)

func (p *parser) statement() (ast.Stmt, *utils.Diagnostic) {
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

func (p *parser) block() (ast.Stmt, *utils.Diagnostic) {
	token := p.current

	// Statements
	stmts := make([]ast.Stmt, 0, 4)

	for !p.check(scanner.RightBrace) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	// Right brace
	if _, err := p.consume(scanner.RightBrace, "Expected '}'."); err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.Block{
		Token_: token,
		Stmts:  stmts,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) expressionStmt() (ast.Stmt, *utils.Diagnostic) {
	token := p.next

	// Expression
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, isAssignment := expr.(*ast.Assignment)
	_, isCall := expr.(*ast.Call)

	if !isAssignment && !isCall {
		return nil, p.error(token, "Invalid statement.")
	}

	// Semicolon
	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.Expression{
		Token_: token,
		Expr:   expr,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) variable() (ast.Stmt, *utils.Diagnostic) {
	start := p.current

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

		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Semicolon
	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.Variable{
		Type:        type_,
		Name:        name,
		Initializer: initializer,
		InferType:   type_ == nil,
	}

	stmt.SetRangeToken(start, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) if_() (ast.Stmt, *utils.Diagnostic) {
	token := p.current

	// Left paren
	if _, err := p.consume(scanner.LeftParen, "Expected '(' before condition."); err != nil {
		return nil, err
	}

	// Condition
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Right paren
	if _, err := p.consume(scanner.RightParen, "Expected ')' before condition."); err != nil {
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
	stmt := &ast.If{
		Token_:    token,
		Condition: condition,
		Then:      then,
		Else:      else_,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) for_() (ast.Stmt, *utils.Diagnostic) {
	token := p.current

	var condition ast.Expr

	// Left paren
	if p.match(scanner.LeftParen) {
		// Condition
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		condition = expr

		// Right paren
		if _, err := p.consume(scanner.RightParen, "Expected ')' before condition."); err != nil {
			return nil, err
		}
	}

	// Body
	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.For{
		Token_:    token,
		Condition: condition,
		Body:      body,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) return_() (ast.Stmt, *utils.Diagnostic) {
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
	stmt := &ast.Return{
		Token_: token,
		Expr:   expr,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}

func (p *parser) break_() (ast.Stmt, *utils.Diagnostic) {
	token := p.current

	// Semicolon
	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.Break{
		Token_: token,
	}

	stmt.SetRangeToken(token, p.current)
	return stmt, nil
}

func (p *parser) continue_() (ast.Stmt, *utils.Diagnostic) {
	token := p.current

	// Semicolon
	if _, err := p.consume(scanner.Semicolon, "Expected ';'."); err != nil {
		return nil, err
	}

	// Return
	stmt := &ast.Continue{
		Token_: token,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt, nil
}
