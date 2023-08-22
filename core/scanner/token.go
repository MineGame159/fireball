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
	Func
	Continue
	Break
	Return

	Number
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
