package cst

import "fireball/core/scanner"

var canStartStmt = []scanner.TokenKind{
	scanner.LeftBrace,
	scanner.Var,
	scanner.If,
	scanner.For,
	scanner.Return,
	scanner.Break,
	scanner.Continue,
}

func init() {
	canStartStmt = append(canStartStmt, canStartExpr...)
}

func parseStmt(p *parser) Node {
	if p.peekIs(canStartExpr) {
		return parseExprStmt(p)
	}

	switch p.peek() {
	case scanner.LeftBrace:
		return parseBlockStmt(p)
	case scanner.Var:
		return parseVarStmt(p)
	case scanner.If:
		return parseIfStmt(p)
	case scanner.For:
		return parseForStmt(p)
	case scanner.Return:
		return parseReturnStmt(p)
	case scanner.Break:
		return parseBreakStmt(p)
	case scanner.Continue:
		return parseContinueStmt(p)

	default:
		return p.error("Cannot start a statement")
	}
}

func parseExprStmt(p *parser) Node {
	p.begin(ExprStmtNode)

	if p.child(parseExpr) {
		return p.end()
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

func parseBlockStmt(p *parser) Node {
	p.begin(BlockStmtNode)

	if p.consume(scanner.LeftBrace) {
		return p.end()
	}
	if p.repeat(parseStmt, canStartStmt...) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

func parseVarStmt(p *parser) Node {
	p.begin(VarStmtNode)

	if p.consume(scanner.Var) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.peekIs(canStartType) {
		if p.child(parseType) {
			return p.end()
		}
	}
	if p.optional(scanner.Equal) {
		if p.child(parseExpr) {
			return p.end()
		}
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

func parseIfStmt(p *parser) Node {
	p.begin(IfStmtNode)

	if p.consume(scanner.If) {
		return p.end()
	}
	if p.consume(scanner.LeftParen) {
		return p.end()
	}
	if p.child(parseExpr) {
		return p.end()
	}
	if p.consume(scanner.RightParen) {
		return p.end()
	}
	if p.child(parseStmt) {
		return p.end()
	}
	if p.optional(scanner.Else) {
		if p.child(parseStmt) {
			return p.end()
		}
	}

	return p.end()
}

func parseForStmt(p *parser) Node {
	p.begin(ForStmtNode)

	if p.consume(scanner.For) {
		return p.end()
	}
	if p.consume(scanner.LeftParen) {
		return p.end()
	}

	if p.peekIs(canStartStmt) {
		if p.child(parseStmt) {
			return p.end()
		}
	} else {
		if p.consume(scanner.Semicolon) {
			return p.end()
		}
	}

	if p.peekIs(canStartExpr) {
		if p.child(parseExpr) {
			return p.end()
		}
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	if p.peekIs(canStartExpr) {
		if p.child(parseExpr) {
			return p.end()
		}
	}

	if p.consume(scanner.RightParen) {
		return p.end()
	}
	if p.child(parseStmt) {
		return p.end()
	}

	return p.end()
}

func parseReturnStmt(p *parser) Node {
	p.begin(ReturnStmtNode)

	if p.consume(scanner.Return) {
		return p.end()
	}
	if p.peekIs(canStartExpr) {
		if p.child(parseExpr) {
			return p.end()
		}
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

func parseBreakStmt(p *parser) Node {
	p.begin(BreakStmtNode)

	p.consume(scanner.Break)

	return p.end()
}

func parseContinueStmt(p *parser) Node {
	p.begin(ContinueStmtNode)

	p.consume(scanner.Continue)

	return p.end()
}
