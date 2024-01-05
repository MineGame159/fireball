package core

type Pos struct {
	Line   uint16
	Column uint16
}

func (p Pos) IsAfter(after Pos) bool {
	if p.Line == after.Line {
		return p.Column > after.Column
	}

	return p.Line > after.Line
}

type Range struct {
	Start Pos
	End   Pos
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
