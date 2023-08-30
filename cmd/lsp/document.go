package lsp

import (
	"context"
	"errors"
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/checker"
	"fireball/core/parser"
	"fireball/core/scanner"
	"fireball/core/typeresolver"
	"go.lsp.dev/protocol"
	"sync"
)

type Documents struct {
	client protocol.Client

	docs map[protocol.URI]*Document
}

type Document struct {
	client protocol.Client

	Uri   protocol.URI
	Text  string
	Decls []ast.Decl

	parseWaitGroup sync.WaitGroup
}

// Documents

func (d *Documents) Add(uri protocol.URI) *Document {
	doc := &Document{
		client: d.client,
		Uri:    uri,
	}

	d.docs[uri] = doc
	return doc
}

func (d *Documents) Remove(uri protocol.URI) {
	delete(d.docs, uri)
}

func (d *Documents) Get(uri protocol.URI) (*Document, error) {
	if doc, ok := d.docs[uri]; ok {
		return doc, nil
	}

	return nil, errors.New("unknown document: " + uri.Filename())
}

// Document

func (d *Document) SetText(ctx context.Context, text string) error {
	d.Text = text

	reporter := &diagnosticReporter{
		diagnostics: make([]protocol.Diagnostic, 0),
	}

	d.parseWaitGroup.Add(1)
	d.Decls = parser.Parse(reporter, scanner.NewScanner(text))
	d.parseWaitGroup.Done()

	typeresolver.Resolve(reporter, d.Decls)
	checker.Check(reporter, d.Decls)

	return d.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         d.Uri,
		Diagnostics: reporter.diagnostics,
	})
}

func (d *Document) EnsureParsed() {
	d.parseWaitGroup.Wait()
}

type diagnosticReporter struct {
	diagnostics []protocol.Diagnostic
}

func (d *diagnosticReporter) Report(diag core.Diagnostic) {
	severity := protocol.DiagnosticSeverityError

	if diag.Kind == core.WarningKind {
		severity = protocol.DiagnosticSeverityWarning
	}

	d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
		Range:    convertRange(diag.Range),
		Severity: severity,
		Message:  diag.Message,
	})
}
