package parser

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) statement() ast.Stmt {
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

func (p *parser) block() ast.Stmt {
	token := p.current

	// Statements
	stmts := make([]ast.Stmt, 0, 4)

	for p.canLoop(scanner.RightBrace) {
		stmt := p.statement()

		if stmt == nil {
			if !p.syncToStmt() {
				break
			}
			continue
		}

		stmts = append(stmts, stmt)
	}

	// Right brace
	_ = p.consume(scanner.RightBrace, "Expected '}'.")

	// Return
	stmt := &ast.Block{
		Token_: token,
		Stmts:  stmts,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt
}

func (p *parser) expressionStmt() ast.Stmt {
	token := p.next

	// Expression
	expr := p.expression()
	if expr == nil {
		return nil
	}

	_, isAssignment := expr.(*ast.Assignment)
	_, isCall := expr.(*ast.Call)

	if !isAssignment && !isCall {
		p.error(token, "Invalid statement.")
		return nil
	}

	// Semicolon
	_ = p.consume(scanner.Semicolon, "Expected ';'.")

	// Return
	stmt := &ast.Expression{
		Token_: token,
		Expr:   expr,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt
}

func (p *parser) variable() ast.Stmt {
	start := p.current

	// Name
	name := p.consume(scanner.Identifier, "expected variable name")
	if name.IsError() {
		return nil
	}

	// Type
	var type_ types.Type

	if !p.check(scanner.Equal) {
		type__ := p.parseType()
		if type__ == nil {
			return nil
		}

		type_ = type__
	}

	// Initializer
	var initializer ast.Expr

	if !p.check(scanner.Semicolon) {
		if token := p.consume(scanner.Equal, "Expected '='."); token.IsError() {
			return nil
		}

		initializer = p.expression()
		if initializer == nil {
			return nil
		}
	}

	// Semicolon
	_ = p.consume(scanner.Semicolon, "Expected ';'.")

	// Return
	stmt := &ast.Variable{
		Type:        type_,
		Name:        name,
		Initializer: initializer,
		InferType:   type_ == nil,
	}

	stmt.SetRangeToken(start, p.current)
	stmt.SetChildrenParent()

	return stmt
}

func (p *parser) if_() ast.Stmt {
	token := p.current

	// Left paren
	if token := p.consume(scanner.LeftParen, "Expected '(' before condition."); token.IsError() {
		return nil
	}

	// Condition
	condition := p.expression()
	if condition == nil {
		return nil
	}

	// Right paren
	if token := p.consume(scanner.RightParen, "Expected ')' before condition."); token.IsError() {
		return nil
	}

	// Then
	then := p.statement()
	if then == nil {
		return nil
	}

	// Else
	var else_ ast.Stmt

	if p.match(scanner.Else) {
		else__ := p.statement()
		if else__ == nil {
			return nil
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

	return stmt
}

func (p *parser) for_() ast.Stmt {
	token := p.current

	var condition ast.Expr

	// Left paren
	if p.match(scanner.LeftParen) {
		// Condition
		expr := p.expression()
		if expr == nil {
			return nil
		}

		condition = expr

		// Right paren
		if token := p.consume(scanner.RightParen, "Expected ')' before condition."); token.IsError() {
			return nil
		}
	}

	// Body
	body := p.statement()
	if body == nil {
		return nil
	}

	// Return
	stmt := &ast.For{
		Token_:    token,
		Condition: condition,
		Body:      body,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt
}

func (p *parser) return_() ast.Stmt {
	token := p.current

	// Value
	var expr ast.Expr

	if !p.check(scanner.Semicolon) {
		expr_ := p.expression()
		if expr_ == nil {
			return nil
		}

		expr = expr_
	}

	// Semicolon
	_ = p.consume(scanner.Semicolon, "Expected ';'.")

	// Return
	stmt := &ast.Return{
		Token_: token,
		Expr:   expr,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt
}

func (p *parser) break_() ast.Stmt {
	token := p.current

	// Semicolon
	_ = p.consume(scanner.Semicolon, "Expected ';'.")

	// Return
	stmt := &ast.Break{
		Token_: token,
	}

	stmt.SetRangeToken(token, p.current)
	return stmt
}

func (p *parser) continue_() ast.Stmt {
	token := p.current

	// Semicolon
	_ = p.consume(scanner.Semicolon, "Expected ';'.")

	// Return
	stmt := &ast.Continue{
		Token_: token,
	}

	stmt.SetRangeToken(token, p.current)
	stmt.SetChildrenParent()

	return stmt
}
