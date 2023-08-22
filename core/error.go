package core

import "fmt"

type Error struct {
	Message string

	Line   int
	Column int
}

func (p *Error) Error() string {
	return fmt.Sprintf("[%d:%d] %s", p.Line, p.Column, p.Message)
}

type ErrorReporter interface {
	Report(error Error)
}
