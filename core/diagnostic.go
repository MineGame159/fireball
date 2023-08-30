package core

import (
	"fireball/core/ast"
	"fmt"
)

type DiagnosticKind uint8

const (
	ErrorKind DiagnosticKind = iota
	WarningKind
)

type Diagnostic struct {
	Kind    DiagnosticKind
	Range   ast.Range
	Message string
}

func (p *Diagnostic) String() string {
	return fmt.Sprintf("[%d:%d] %s", p.Range.Start.Line, p.Range.Start.Column, p.Message)
}

type Reporter interface {
	Report(diag Diagnostic)
}

type NopReporter struct {
}

func (n NopReporter) Report(error Diagnostic) {
}
