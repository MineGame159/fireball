package scanner

import (
	"fireball/core"
	"slices"
)

type TokenKind uint8

const (
	Error TokenKind = iota

	LeftParen
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
	DotDotDot
	And
	Or

	Nil
	True
	False
	Var
	If
	Else
	While
	For
	As
	Is
	Static
	Func
	Fn
	Continue
	Break
	Return
	Namespace
	Using
	Struct
	Impl
	Enum
	Interface

	Number
	Hex
	Binary
	Character
	String
	Identifier

	Comment

	Eof
)

func (t TokenKind) IsAny(kinds []TokenKind) bool {
	return slices.Contains(kinds, t)
}

var LogicalOperators = []TokenKind{
	And,
	Or,
}

var AssignmentOperators = []TokenKind{
	Equal,

	PlusEqual,
	MinusEqual,
	StarEqual,
	SlashEqual,
	PercentageEqual,

	PipeEqual,
	XorEqual,
	AmpersandEqual,

	LessLessEqual,
	GreaterGreaterEqual,
}

var CastOperators = []TokenKind{
	As,
	Is,
}

type Token struct {
	Kind   TokenKind
	Lexeme string
}

func (t Token) IsError() bool {
	return t.Kind == Error
}

func (t Token) String() string {
	return t.Lexeme
}

func (t Token) IsEmpty() bool {
	return t.Kind == Error && t.Lexeme == ""
}

type PositionedToken struct {
	Token

	Pos core.Pos
}

func (p PositionedToken) Range() core.Range {
	return core.Range{
		Start: p.Pos,
		End: core.Pos{
			Line:   p.Pos.Line,
			Column: p.Pos.Column + uint16(len(p.Lexeme)),
		},
	}
}

func IsKeyword(kind TokenKind) bool {
	return kind >= Nil && kind <= Interface
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

func TokenKindStr(kind TokenKind) string {
	switch kind {
	case Error:
		return "<error>"

	case LeftParen:
		return "'('"
	case RightParen:
		return "')'"
	case LeftBrace:
		return "'{'"
	case RightBrace:
		return "'}'"
	case LeftBracket:
		return "'['"
	case RightBracket:
		return "']'"

	case Dot:
		return "'.'"
	case Comma:
		return "','"
	case Colon:
		return "':'"
	case Semicolon:
		return "';'"

	case Plus:
		return "'+'"
	case Minus:
		return "'-'"
	case Star:
		return "'*'"
	case Slash:
		return "'/'"
	case Percentage:
		return "'%'"

	case PlusEqual:
		return "'+='"
	case MinusEqual:
		return "'-='"
	case StarEqual:
		return "'*='"
	case SlashEqual:
		return "'/='"
	case PercentageEqual:
		return "'%='"

	case PlusPlus:
		return "'++'"
	case MinusMinus:
		return "'--'"

	case Equal:
		return "'='"
	case EqualEqual:
		return "'=='"
	case Bang:
		return "'!'"
	case BangEqual:
		return "'!='"
	case Less:
		return "'<'"
	case LessEqual:
		return "'<='"
	case Greater:
		return "'>'"
	case GreaterEqual:
		return "'>='"

	case Pipe:
		return "'|'"
	case PipeEqual:
		return "'|='"
	case Ampersand:
		return "'&'"
	case AmpersandEqual:
		return "'&='"
	case Xor:
		return "'^'"
	case XorEqual:
		return "'^='"
	case LessLess:
		return "'<<'"
	case LessLessEqual:
		return "'<<='"
	case GreaterGreater:
		return "'>>'"
	case GreaterGreaterEqual:
		return "'>>='"
	case FuncPtr:
		return "'fn'"
	case Hashtag:
		return "'#'"

	case Nil:
		return "'nil'"
	case True:
		return "'true'"
	case False:
		return "'false'"
	case And:
		return "'and'"
	case Or:
		return "'or'"
	case Var:
		return "'var'"
	case If:
		return "'if'"
	case Else:
		return "'else'"
	case While:
		return "'while'"
	case For:
		return "'for'"
	case As:
		return "'as'"
	case Is:
		return "'is'"
	case Static:
		return "'static'"
	case Func:
		return "'func'"
	case Fn:
		return "'fn'"
	case Continue:
		return "'continue'"
	case Break:
		return "'break'"
	case Return:
		return "'return'"
	case Namespace:
		return "'namespace'"
	case Using:
		return "'using'"
	case Struct:
		return "'struct'"
	case Impl:
		return "'impl'"
	case Enum:
		return "'enum'"
	case Interface:
		return "'interface'"

	case Number:
		return "number"
	case Hex:
		return "hexadecimal number"
	case Binary:
		return "binary number"
	case Character:
		return "character"
	case String:
		return "string"
	case Identifier:
		return "identifier"

	case Comment:
		return "<comment>"

	case Eof:
		return "<eof>"

	default:
		panic("scanner.TokenKindStr() - Not implemented: " + string(kind))
	}
}
