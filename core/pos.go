package core

import "fireball/core/scanner"

type Range struct {
	Start Pos
	End   Pos
}

type Pos struct {
	Line   int
	Column int
}

func (r Range) Valid() bool {
	return r.End.Line > 0 && r.End.Column > 0
}

func (r Range) Contains(pos Pos) bool {
	// Check line
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}

	// Check start column
	if pos.Line == r.Start.Line && pos.Column < r.Start.Column {
		return false
	}

	// Check end column
	if pos.Line == r.End.Line && pos.Column > r.End.Column {
		return false
	}

	// True
	return true
}

func TokensToRange(start, end scanner.Token) Range {
	return Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func TokenToRange(token scanner.Token) Range {
	return Range{
		Start: TokenToPos(token, false),
		End:   TokenToPos(token, true),
	}
}

func TokenToPos(token scanner.Token, end bool) Pos {
	offset := 0

	if end {
		offset = len(token.Lexeme)
	}

	return Pos{
		Line:   token.Line(),
		Column: token.Column() + offset,
	}
}
