package cst

import (
	"fireball/core/scanner"
	"fireball/core/utils"
)

var canStartExpr = []scanner.TokenKind{
	scanner.Nil,
	scanner.True,
	scanner.False,
	scanner.Number,
	scanner.Hex,
	scanner.Binary,
	scanner.String,
	scanner.Identifier,

	scanner.LeftParen,
	scanner.LeftBracket,
}

var infixAndPostfixOperators []scanner.TokenKind

// Main parse function

func parseExpr(p *parser) Node {
	return parseExprPratt(p, 0)
}

func parseExprPratt(p *parser, minPower int) Node {
	lhs := parsePrefixExprPratt(p)
	if p.recovering() {
		return lhs
	}

	for p.peekIs(infixAndPostfixOperators) {
		op := p.peek()

		if leftPower := postfixExprPower(op); leftPower != -1 {
			if leftPower < minPower {
				break
			}

			lhs = parsePostfixExprPratt(p, op, lhs)
			if p.recovering() {
				return lhs
			}
		}

		if leftPower, rightPower := infixExprPower(op); leftPower != -1 {
			if leftPower < minPower {
				break
			}

			lhs = parseInfixExprPratt(p, op, lhs, rightPower)
			if p.recovering() {
				return lhs
			}
		}
	}

	return lhs
}

// Parsing

var canStartStructFieldExpr = []scanner.TokenKind{scanner.Identifier}

func parsePrefixExprPratt(p *parser) Node {
	switch p.peek() {
	case scanner.Nil, scanner.True, scanner.False, scanner.Number, scanner.Hex, scanner.Binary, scanner.Character, scanner.String:
		return p.advanceGetLeaf()

	case scanner.Identifier:
		if p.next.Lexeme == "new" && p.peek2Is(canStartType) {
			new_ := p.advanceGetLeaf()
			type_ := parseType(p)

			if p.peek() == scanner.LeftBrace {
				// New struct
				p.begin(StructExprNode)

				p.childAdd(new_)
				if p.childAdd(type_) {
					return p.end()
				}
				if p.consume(scanner.LeftBrace) {
					return p.end()
				}
				if p.repeatSeparated(parseStructFieldExpr, canStartStructFieldExpr, scanner.Comma) {
					return p.end()
				}
				if p.consume(scanner.RightBrace) {
					return p.end()
				}

				return p.end()
			}

			// New array
			p.begin(AllocateArrayExprNode)

			p.childAdd(new_)
			if p.childAdd(type_) {
				return p.end()
			}
			if p.consume(scanner.LeftBracket) {
				return p.end()
			}
			if p.child(parseExpr) {
				return p.end()
			}
			if p.consume(scanner.RightBracket) {
				return p.end()
			}

			return p.end()
		}

		// Identifier
		p.begin(IdentifierExprNode)

		p.advanceAddChild()

		if p.peek() == scanner.Bang && p.peek2() == scanner.LeftBracket {
			p.advanceAddChild()
			p.advanceAddChild()

			if p.repeatSeparated(parseType, canStartType, scanner.Comma) {
				return p.end()
			}
			if p.consume(scanner.RightBracket) {
				return p.end()
			}
		}

		return p.end()

	case scanner.LeftParen:
		p.begin(ParenExprNode)

		p.advanceAddChild()
		if p.childAdd(parseExprPratt(p, 0)) {
			return p.end()
		}
		if p.consume(scanner.RightParen) {
			return p.end()
		}

		return p.end()

	case scanner.LeftBracket:
		p.begin(ArrayExprNode)

		p.advanceAddChild()
		if p.repeatSeparated(parseExpr, canStartExpr, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightBracket) {
			return p.end()
		}

		return p.end()

	default:
		rightPower := prefixExprPower(p.peek())

		if rightPower == -1 {
			return p.error("Cannot start an expression")
		}

		p.begin(UnaryExprNode)

		p.advanceAddChild()
		p.childAdd(parseExprPratt(p, rightPower))

		return p.end()
	}
}

func parseStructFieldExpr(p *parser) Node {
	p.begin(StructFieldExprNode)

	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.Colon) {
		return p.end()
	}
	if p.child(parseExpr) {
		return p.end()
	}

	return p.end()
}

func parseInfixExprPratt(p *parser, op scanner.TokenKind, lhs Node, rightPower int) Node {
	switch op {
	case scanner.As, scanner.Is:
		p.begin(BinaryExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()
		p.child(parseType)

		return p.end()

	case scanner.Dot:
		p.begin(BinaryExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()
		p.consume(scanner.Identifier)

		if p.peek() == scanner.Bang && p.peek2() == scanner.LeftBracket {
			p.advanceAddChild()
			p.advanceAddChild()

			if p.repeatSeparated(parseType, canStartType, scanner.Comma) {
				return p.end()
			}
			if p.consume(scanner.RightBracket) {
				return p.end()
			}
		}

		return p.end()

	default:
		p.begin(BinaryExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()
		p.childAdd(parseExprPratt(p, rightPower))

		return p.end()
	}
}

func parsePostfixExprPratt(p *parser, op scanner.TokenKind, lhs Node) Node {
	switch op {
	case scanner.LeftBracket:
		p.begin(IndexExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()
		if p.childAdd(parseExprPratt(p, 0)) {
			return p.end()
		}
		if p.consume(scanner.RightBracket) {
			return p.end()
		}

		return p.end()

	case scanner.LeftParen:
		lhsLexeme := identifierExprOrTokenLexeme(lhs)

		// Type call
		if lhsLexeme == "sizeof" || lhsLexeme == "alignof" {
			p.begin(TypeCallExprNode)

			p.childAdd(lhs)
			p.advanceAddChild()
			if p.child(parseType) {
				return p.end()
			}
			if p.consume(scanner.RightParen) {
				return p.end()
			}

			return p.end()
		}

		// Typeof
		if lhsLexeme == "typeof" {
			p.begin(TypeofExprNode)

			p.childAdd(lhs)
			p.advanceAddChild()
			if p.child(parseExpr) {
				return p.end()
			}
			if p.consume(scanner.RightParen) {
				return p.end()
			}

			return p.end()
		}

		// Call
		p.begin(CallExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()
		if p.repeatSeparated(parseExpr, canStartExpr, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightParen) {
			return p.end()
		}

		return p.end()

	case scanner.LeftBrace:
		p.begin(StructExprNode)

		if p.childAdd(convertExprToIdentifierType(p, lhs)) {
			return p.end()
		}
		if p.consume(scanner.LeftBrace) {
			return p.end()
		}
		if p.repeatSeparated(parseStructFieldExpr, canStartStructFieldExpr, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightBrace) {
			return p.end()
		}

		return p.end()

	default:
		p.begin(UnaryExprNode)

		p.childAdd(lhs)
		p.advanceAddChild()

		return p.end()
	}
}

func convertExprToIdentifierType(p *parser, node Node) Node {
	p.begin(IdentifierTypeNode)

	convertExprToIdentifierTypePart(p, node)

	for _, child := range node.Children {
		if child.Token.Kind == scanner.LeftBracket || child.Token.Kind == scanner.RightBracket || child.Kind.IsType() {
			p.childAdd(child)
		}
	}

	return p.end()
}

func convertExprToIdentifierTypePart(p *parser, node Node) {
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch node.Kind {
	case TokenNode:
		p.childAdd(node)
		return

	case IdentifierExprNode:
		p.childAdd(node.Children[0])
		return

	case BinaryExprNode:
		if node.Contains(scanner.Dot) {
			convertExprToIdentifierTypePart(p, node.Children[0])

			p.childAdd(node.Children[1])

			if len(node.Children) == 3 {
				convertExprToIdentifierTypePart(p, node.Children[2])
			}

			return
		}
	}

	p.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Range,
		Message: "Invalid type",
	})
}

// Powers

type tokenPowers struct {
	prefixRightPower int

	infixLeftPower  int
	infixRightPower int

	postfixLeftPower int
}

var tokenPowerTable = make([]tokenPowers, scanner.Eof+1)
var tokenPowerTableCount = 0

func init() {
	// Set every power to -1
	for i := 0; i < len(tokenPowerTable); i++ {
		tokenPowerTable[i] = tokenPowers{
			prefixRightPower: -1,
			infixLeftPower:   -1,
			infixRightPower:  -1,
			postfixLeftPower: -1,
		}
	}

	// =, +=, -=, *=, /=, %=, |=, ^=, &=, <<=, >>=
	infix(false, scanner.Equal, scanner.PlusEqual, scanner.MinusEqual, scanner.StarEqual, scanner.SlashEqual, scanner.PercentageEqual, scanner.PipeEqual, scanner.XorEqual, scanner.AmpersandEqual, scanner.LessLessEqual, scanner.GreaterGreaterEqual)
	// ||
	infix(false, scanner.Or)
	// &&
	infix(false, scanner.And)
	// |
	infix(false, scanner.Pipe)
	// ^
	infix(false, scanner.Xor)
	// &
	infix(false, scanner.Ampersand)
	// ==, !=
	infix(false, scanner.EqualEqual, scanner.BangEqual)
	// >, <=, >, >=, as, is
	infix(false, scanner.Less, scanner.LessEqual, scanner.Greater, scanner.GreaterEqual, scanner.As, scanner.Is)
	// <<, >>
	infix(false, scanner.LessLess, scanner.GreaterGreater)
	// +, -
	infix(false, scanner.Plus, scanner.Minus)
	// *, /, %
	infix(false, scanner.Star, scanner.Slash, scanner.Percentage)
	// -x, !x, ++x, --x, &x, *x, => x
	prefix(scanner.Minus, scanner.Bang, scanner.PlusPlus, scanner.MinusMinus, scanner.Ampersand, scanner.Star, scanner.FuncPtr)
	// x++, x--
	postfix(scanner.PlusPlus, scanner.MinusMinus)
	// x[], x(), x {}
	postfix(scanner.LeftBracket, scanner.LeftParen, scanner.LeftBrace)
	// x.y
	infix(false, scanner.Dot)
}

func prefix(kinds ...scanner.TokenKind) {
	for _, kind := range kinds {
		tokenPowerTable[kind].prefixRightPower = (tokenPowerTableCount * 2) + 1

		canStartExpr = append(canStartExpr, kind)
	}

	tokenPowerTableCount++
}

func infix(rightAssociative bool, kinds ...scanner.TokenKind) {
	for _, kind := range kinds {
		if rightAssociative {
			tokenPowerTable[kind].infixLeftPower = (tokenPowerTableCount * 2) + 2
			tokenPowerTable[kind].infixRightPower = (tokenPowerTableCount * 2) + 1
		} else {
			tokenPowerTable[kind].infixLeftPower = (tokenPowerTableCount * 2) + 1
			tokenPowerTable[kind].infixRightPower = (tokenPowerTableCount * 2) + 2
		}

		infixAndPostfixOperators = append(infixAndPostfixOperators, kind)
	}

	tokenPowerTableCount++
}

func postfix(kinds ...scanner.TokenKind) {
	for _, kind := range kinds {
		tokenPowerTable[kind].postfixLeftPower = (tokenPowerTableCount * 2) + 1

		infixAndPostfixOperators = append(infixAndPostfixOperators, kind)
	}

	tokenPowerTableCount++
}

func prefixExprPower(kind scanner.TokenKind) int {
	return tokenPowerTable[kind].prefixRightPower
}

func infixExprPower(kind scanner.TokenKind) (int, int) {
	powers := tokenPowerTable[kind]
	return powers.infixLeftPower, powers.infixRightPower
}

func postfixExprPower(kind scanner.TokenKind) int {
	return tokenPowerTable[kind].postfixLeftPower
}
