package scanner

type TokenKind uint8

const (
	LeftParen TokenKind = iota
	RightParen
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket

	Dot
	Comma
	Colon
	Semicolon
	Ampersand

	Plus
	Minus
	Star
	Slash
	Percentage

	PlusEqual
	MinusEqual
	StarEqual
	SlashEqual
	PercentageEqual

	Equal
	EqualEqual
	Bang
	BangEqual
	Less
	LessEqual
	Greater
	GreaterEqual

	Nil
	True
	False
	And
	Or
	Var
	If
	Else
	While
	For
	As
	Extern
	Func
	Continue
	Break
	Return

	Number
	Character
	String
	Identifier

	Error
	Eof
)

type Token struct {
	Kind   TokenKind
	Lexeme string

	Line   int
	Column int
}

func (t Token) IsError() bool {
	return t.Kind == Error
}

func (t Token) String() string {
	return t.Lexeme
}

func IsEquality(kind TokenKind) bool {
	return kind == EqualEqual || kind == BangEqual
}

func IsComparison(kind TokenKind) bool {
	switch kind {
	case Less, LessEqual, Greater, GreaterEqual:
		return true

	default:
		return false
	}
}

func IsArithmetic(kind TokenKind) bool {
	switch kind {
	case Plus, Minus, Star, Slash, Percentage:
		return true

	default:
		return false
	}
}

func IsKeyword(kind TokenKind) bool {
	return kind >= 29 && kind <= 44
}

func IsOperator(kind TokenKind) bool {
	return kind >= 10 && kind <= 28
}
