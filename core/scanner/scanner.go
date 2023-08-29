package scanner

type Scanner struct {
	text string

	startI   int
	currentI int

	line   int
	column int
}

func NewScanner(text string) *Scanner {
	return &Scanner{
		text:     text,
		startI:   0,
		currentI: 0,
		line:     1,
	}
}

func (s *Scanner) Next() Token {
	s.skipWhitespace()
	s.startI = s.currentI

	if s.isAtEnd() {
		return s.make(Eof)
	}

	c := s.advance()

	if isAlpha(c) {
		return s.identifier()
	}
	if isDigit(c) {
		return s.number()
	}

	switch c {
	case '(':
		return s.make(LeftParen)
	case ')':
		return s.make(RightParen)
	case '{':
		return s.make(LeftBrace)
	case '}':
		return s.make(RightBrace)
	case '[':
		return s.make(LeftBracket)
	case ']':
		return s.make(RightBracket)

	case '.':
		return s.make(Dot)
	case ',':
		return s.make(Comma)
	case ':':
		return s.make(Colon)
	case ';':
		return s.make(Semicolon)
	case '&':
		return s.make(Ampersand)

	case '+':
		return s.matchToken('=', PlusEqual, Plus)
	case '-':
		return s.matchToken('=', MinusEqual, Minus)
	case '*':
		return s.matchToken('=', StarEqual, Star)
	case '/':
		return s.matchToken('=', SlashEqual, Slash)
	case '%':
		return s.matchToken('=', PercentageEqual, Percentage)

	case '=':
		return s.matchToken('=', EqualEqual, Equal)
	case '!':
		return s.matchToken('=', BangEqual, Bang)
	case '<':
		return s.matchToken('=', LessEqual, Less)
	case '>':
		return s.matchToken('=', GreaterEqual, Greater)

	case '"':
		return s.string()
	}

	return s.error("Unexpected character.")
}

func (s *Scanner) identifier() Token {
	for isAlpha(s.peek()) || isDigit(s.peek()) {
		s.advance()
	}

	return s.make(s.identifierKind())
}

func (s *Scanner) identifierKind() TokenKind {
	switch s.text[s.startI] {
	case 'a':
		if s.currentI-s.startI > 1 {
			switch s.text[s.startI+1] {
			case 'n':
				return s.checkKeyword(2, "d", And)
			case 's':
				return As
			}
		}
	case 'b':
		return s.checkKeyword(1, "reak", Break)
	case 'c':
		return s.checkKeyword(1, "ontinue", Continue)
	case 'e':
		if s.currentI-s.startI > 1 {
			switch s.text[s.startI+1] {
			case 'l':
				return s.checkKeyword(2, "se", Else)
			case 'x':
				return s.checkKeyword(2, "tern", Extern)
			}
		}
	case 'f':
		if s.currentI-s.startI > 1 {
			switch s.text[s.startI+1] {
			case 'a':
				return s.checkKeyword(2, "lse", False)
			case 'o':
				return s.checkKeyword(2, "r", For)
			case 'u':
				return s.checkKeyword(2, "nc", Func)
			}
		}
	case 'i':
		return s.checkKeyword(1, "f", If)
	case 'n':
		return s.checkKeyword(1, "il", Nil)
	case 'o':
		return s.checkKeyword(1, "r", Or)
	case 'r':
		return s.checkKeyword(1, "eturn", Return)
	case 't':
		return s.checkKeyword(1, "rue", True)
	case 'v':
		return s.checkKeyword(1, "ar", Var)
	case 'w':
		return s.checkKeyword(1, "hile", While)
	}

	return Identifier
}

func (s *Scanner) checkKeyword(start int, rest string, kind TokenKind) TokenKind {
	if s.currentI-s.startI == start+len(rest) && s.text[s.startI+start:s.startI+start+len(rest)] == rest {
		return kind
	}

	return Identifier
}

func (s *Scanner) number() Token {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	s.match('f')

	return s.make(Number)
}

func (s *Scanner) string() Token {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}

		s.advance()
	}

	if s.isAtEnd() {
		return s.error("Unterminated string")
	}

	s.advance()
	return s.make(String)
}

func (s *Scanner) matchToken(expected uint8, kindTrue TokenKind, kindFalse TokenKind) Token {
	if s.match(expected) {
		return s.make(kindTrue)
	}

	return s.make(kindFalse)
}

func (s *Scanner) match(expected uint8) bool {
	if s.isAtEnd() {
		return false
	}

	if s.peek() != expected {
		return false
	}

	s.advance()
	return true
}

func (s *Scanner) skipWhitespace() {
	for {
		if s.isAtEnd() {
			return
		}

		c := s.peek()

		switch c {
		case ' ', '\r', '\t':
			s.advance()
		case '\n':
			s.advance()
			s.line++
			s.column = 0
		case '/':
			if s.peekNext() == '/' {
				for s.peek() != '\n' && !s.isAtEnd() {
					s.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *Scanner) peek() uint8 {
	if s.isAtEnd() {
		return '\000'
	}

	return s.text[s.currentI]
}

func (s *Scanner) peekNext() uint8 {
	if s.isAtEnd() {
		return '\000'
	}

	return s.text[s.currentI+1]
}

func (s *Scanner) advance() uint8 {
	s.currentI++
	s.column++

	return s.text[s.currentI-1]
}

func (s *Scanner) isAtEnd() bool {
	return s.currentI >= len(s.text)
}

func (s *Scanner) make(kind TokenKind) Token {
	lexeme := s.text[s.startI:s.currentI]

	return Token{
		Kind:   kind,
		Lexeme: lexeme,
		Line:   s.line,
		Column: s.column - len(lexeme),
	}
}

func (s *Scanner) error(msg string) Token {
	return Token{
		Kind:   Error,
		Lexeme: msg,
		Line:   s.line,
		Column: s.column,
	}
}

func isAlpha(c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isDigit(c uint8) bool {
	return c >= '0' && c <= '9'
}
