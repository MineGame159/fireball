package lsp

import (
	"context"
	"errors"
	"fireball/core"
	"fireball/core/checker"
	"fireball/core/parser"
	"fireball/core/scanner"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
	client protocol.Client

	files map[protocol.URI]string
}

type SemanticTokensOptions struct {
	protocol.WorkDoneProgressOptions

	Legend protocol.SemanticTokensLegend `json:"legend"`
	Range  bool                          `json:"range,omitempty"`
	Full   *SemanticTokensFull           `json:"full,omitempty"`
}

type SemanticTokensFull struct {
	Delta bool `json:"delta,omitempty"`
}

func (h *handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (result *protocol.InitializeResult, err error) {
	h.logger.Debug("handle Initialize")

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			SemanticTokensProvider: &SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenTypes: []protocol.SemanticTokenTypes{
						protocol.SemanticTokenKeyword,
						protocol.SemanticTokenOperator,
						protocol.SemanticTokenNumber,
						protocol.SemanticTokenString,
						protocol.SemanticTokenFunction,
						protocol.SemanticTokenVariable,
					},
					TokenModifiers: []protocol.SemanticTokenModifiers{},
				},
				Full: &SemanticTokensFull{},
			},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "fireball",
			Version: "0.1.0",
		},
	}, nil
}

func (h *handler) Initialized(ctx context.Context, params *protocol.InitializedParams) (err error) {
	h.logger.Debug("handle Initialized")

	return nil
}

func (h *handler) Shutdown(ctx context.Context) (err error) {
	h.logger.Debug("handle Shutdown")
	return nil
}

func (h *handler) Exit(ctx context.Context) (err error) {
	h.logger.Debug("handle Exit")
	return nil
}

func (h *handler) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) LogTrace(ctx context.Context, params *protocol.LogTraceParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) SetTrace(ctx context.Context, params *protocol.SetTraceParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) CodeAction(ctx context.Context, params *protocol.CodeActionParams) (result []protocol.CodeAction, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) CodeLens(ctx context.Context, params *protocol.CodeLensParams) (result []protocol.CodeLens, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (result *protocol.CodeLens, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) (result []protocol.ColorPresentation, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (result *protocol.CompletionItem, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) (result []protocol.Location, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	h.logger.Debug("handle DidChange")
	h.files[params.TextDocument.URI] = params.ContentChanges[0].Text

	return h.parse(ctx, params.TextDocument.URI)
}

func (h *handler) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) (err error) {
	h.logger.Debug("handle DidClose")
	delete(h.files, params.TextDocument.URI)

	return nil
}

func (h *handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	h.logger.Debug("handle DidOpen")
	h.files[params.TextDocument.URI] = params.TextDocument.Text

	return h.parse(ctx, params.TextDocument.URI)
}

func (h *handler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) (result []protocol.ColorInformation, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) (result []protocol.DocumentHighlight, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) (result []protocol.DocumentLink, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (result *protocol.DocumentLink, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (result []interface{}, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (result interface{}, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) (result []protocol.FoldingRange, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) (result []protocol.TextEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Hover(ctx context.Context, params *protocol.HoverParams) (result *protocol.Hover, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) (result []protocol.Location, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) (result []protocol.TextEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (result *protocol.Range, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) (result []protocol.TextEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Rename(ctx context.Context, params *protocol.RenameParams) (result *protocol.WorkspaceEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (result *protocol.SignatureHelp, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) (result []protocol.SymbolInformation, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) (result []protocol.Location, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) (result []protocol.TextEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (result *protocol.ShowDocumentResult, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (result *protocol.WorkspaceEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (result *protocol.WorkspaceEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (result *protocol.WorkspaceEdit, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (err error) {
	return errors.New("not implemented")
}

func (h *handler) CodeLensRefresh(ctx context.Context) (err error) {
	return errors.New("not implemented")
}

func (h *handler) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) (result []protocol.CallHierarchyItem, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) (result []protocol.CallHierarchyIncomingCall, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) (result []protocol.CallHierarchyOutgoingCall, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (result *protocol.SemanticTokens, err error) {
	h.logger.Debug("handle SemanticTokensFull")

	s := scanner.NewScanner(h.files[params.TextDocument.URI])
	data := make([]uint32, 0, 64)

	lastTokenKind := scanner.Eof

	lastLine := 0
	lastColumn := 0

	for {
		token := s.Next()
		if token.Kind == scanner.Eof {
			break
		}

		kind := uint32(0)

		if scanner.IsKeyword(token.Kind) {
			kind = 1
		} else if scanner.IsOperator(token.Kind) {
			kind = 2
		} else {
			switch token.Kind {
			case scanner.Number:
				kind = 3

			case scanner.Character, scanner.String:
				kind = 4

			case scanner.Identifier:
				switch lastTokenKind {
				case scanner.Func:
					kind = 5

				case scanner.Var:
					kind = 6
				}
			}
		}

		if kind != 0 {
			if lastLine != token.Line-1 {
				lastColumn = 0
			}

			data = append(data, uint32(token.Line-1-lastLine))
			data = append(data, uint32(token.Column-lastColumn))
			data = append(data, uint32(len(token.Lexeme)))
			data = append(data, kind-1)
			data = append(data, 0)

			lastLine = token.Line - 1
			lastColumn = token.Column
		}

		lastTokenKind = token.Kind
	}

	return &protocol.SemanticTokens{
		Data: data,
	}, nil
}

func (h *handler) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (result interface{}, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (result *protocol.SemanticTokens, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) SemanticTokensRefresh(ctx context.Context) (err error) {
	return errors.New("not implemented")
}

func (h *handler) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (result *protocol.LinkedEditingRanges, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Moniker(ctx context.Context, params *protocol.MonikerParams) (result []protocol.Moniker, err error) {
	return nil, errors.New("not implemented")
}

func (h *handler) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	return nil, errors.New("not implemented")
}

// Utils

func (h *handler) parse(ctx context.Context, uri uri.URI) error {
	reporter := &diagnosticReporter{
		diagnostics: make([]protocol.Diagnostic, 0),
	}

	decls := parser.Parse(reporter, scanner.NewScanner(h.files[uri]))
	checker.Check(reporter, decls)

	return h.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: reporter.diagnostics,
	})
}

type diagnosticReporter struct {
	diagnostics []protocol.Diagnostic
}

func (d *diagnosticReporter) Report(error core.Error) {
	pos := protocol.Position{
		Line:      uint32(error.Line - 1),
		Character: uint32(error.Column),
	}

	d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
		Range: protocol.Range{
			Start: pos,
			End: protocol.Position{
				Line:      pos.Line,
				Character: pos.Character + 1,
			},
		},
		Severity: protocol.DiagnosticSeverityError,
		Message:  error.Message,
	})
}
