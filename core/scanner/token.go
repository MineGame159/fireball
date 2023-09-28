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

	PlusPlus
	MinusMinus

	Equal
	EqualEqual
	Bang
	BangEqual
	Less
	LessEqual
	Greater
	GreaterEqual

	Pipe
	PipeEqual
	Ampersand
	AmpersandEqual
	Xor
	XorEqual
	LessLess
	LessLessEqual
	GreaterGreater
	GreaterGreaterEqual
	FuncPtr
	Hashtag

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
	Static
	Func
	Continue
	Break
	Return
	Struct
	Impl
	Enum

	Number
	Hex
	Binary
	Character
	String
	Identifier

	Error
	Eof
)

type Token struct {
	Kind   TokenKind
	Lexeme string

	line   int
	column int
}

func (t Token) IsError() bool {
	return t.Kind == Error
}

func (t Token) Line() int {
	return t.line
}

func (t Token) Column() int {
	return t.column
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
	case Plus, PlusEqual, Minus, MinusEqual, Star, StarEqual, Slash, SlashEqual, Percentage, PercentageEqual:
		return true

	default:
		return false
	}
}

func IsBitwise(kind TokenKind) bool {
	switch kind {
	case Pipe, PipeEqual, Xor, XorEqual, Ampersand, AmpersandEqual, LessLess, LessLessEqual, GreaterGreater, GreaterGreaterEqual:
		return true

	default:
		return false
	}
}
