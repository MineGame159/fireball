package lsp

import (
	"context"
	"errors"
	"fireball/core/ast"
	"fireball/core/types"
	"github.com/MineGame159/protocol"
	"go.uber.org/zap"
	"strings"
)

type handler struct {
	logger *zap.Logger
	client protocol.Client

	docs *Documents
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
						protocol.SemanticTokenFunction,
						protocol.SemanticTokenParameter,
						protocol.SemanticTokenVariable,
						protocol.SemanticTokenClass,
						protocol.SemanticTokenProperty,
					},
					TokenModifiers: []protocol.SemanticTokenModifiers{},
				},
				Full: &SemanticTokensFull{},
			},
			DocumentSymbolProvider: &protocol.DocumentSymbolOptions{
				Label: "Fireball",
			},
			HoverProvider:     true,
			InlayHintProvider: true,
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

	doc, err := h.docs.Get(params.TextDocument.URI)
	if err != nil {
		return err
	}

	return doc.SetText(ctx, params.ContentChanges[0].Text)
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
	h.docs.Remove(params.TextDocument.URI)

	return nil
}

func (h *handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	h.logger.Debug("handle DidOpen")
	doc := h.docs.Add(params.TextDocument.URI)

	return doc.SetText(ctx, params.TextDocument.Text)
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
	h.logger.Debug("handle DocumentSymbol")

	// GetLeaf document
	doc, err := h.docs.Get(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	doc.EnsureParsed()

	// GetLeaf symbols
	symbols := make([]interface{}, 0, 8)

	for _, decl := range doc.Decls {
		if v, ok := decl.(*ast.Struct); ok {
			// Struct
			symbols = append(symbols, &protocol.DocumentSymbol{
				Name:           v.Name.Lexeme,
				Kind:           protocol.SymbolKindStruct,
				Range:          convertRange(v.Range()),
				SelectionRange: convertRange(ast.TokenToRange(v.Name)),
			})
		} else if v, ok := decl.(*ast.Func); ok {
			// Function
			signature := strings.Builder{}
			signature.WriteRune('(')

			for i, param := range v.Params {
				if i > 0 {
					signature.WriteString(", ")
				}

				signature.WriteString(param.Name.Lexeme)
				signature.WriteRune(' ')
				signature.WriteString(param.Type.String())
			}

			signature.WriteRune(')')

			if !types.IsPrimitive(v.Returns, types.Void) {
				signature.WriteRune(' ')
				signature.WriteString(v.Returns.String())
			}

			symbols = append(symbols, &protocol.DocumentSymbol{
				Name:           v.Name.Lexeme,
				Detail:         signature.String(),
				Kind:           protocol.SymbolKindFunction,
				Range:          convertRange(v.Range()),
				SelectionRange: convertRange(ast.TokenToRange(v.Name)),
			})
		}
	}

	return symbols, nil
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
	h.logger.Debug("handle Hover")

	// GetLeaf document
	doc, err := h.docs.Get(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	doc.EnsureChecked()

	// Convert position
	pos := ast.Pos{
		Line:   int(params.Position.Line + 1),
		Column: int(params.Position.Character),
	}

	// GetLeaf node under cursor
	for _, decl := range doc.Decls {
		node := ast.GetLeaf(decl, pos)

		if expr, ok := node.(ast.Expr); ok {
			// ast.Expt
			text := expr.Type().String()

			// Ignore literal expressions
			if _, ok := expr.(*ast.Literal); ok {
				text = ""
			}

			// Return
			if text != "" {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.PlainText,
						Value: text,
					},
					Range: convertRangePtr(expr.Range()),
				}, nil
			}
		} else if variable, ok := node.(*ast.Variable); ok {
			// ast.Variable
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.PlainText,
					Value: variable.Type.String(),
				},
				Range: convertRangePtr(ast.TokenToRange(variable.Name)),
			}, nil
		}
	}

	// Return nil
	return nil, nil
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

	// GetLeaf document
	doc, err := h.docs.Get(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	doc.EnsureParsed()

	// GetLeaf semantic tokens
	data := highlight(doc.Decls)

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

func (h *handler) InlayHint(ctx context.Context, params *protocol.InlayHintParams) (result []protocol.InlayHint, err error) {
	h.logger.Debug("handle InlayHint")

	// GetLeaf document
	doc, err := h.docs.Get(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	doc.EnsureParsed()

	// GetLeaf hints
	hints := make([]protocol.InlayHint, 0, 8)

	for _, decl := range doc.Decls {
		ast.VisitStmts[*ast.Variable](decl, func(stmt *ast.Variable) {
			if stmt.InferType {
				hints = append(hints, protocol.InlayHint{
					Position: convertPos(ast.TokenToPos(stmt.Name, true)),
					Label:    " " + stmt.Type.String(),
					Kind:     protocol.InlayHintKindType,
				})
			}
		})
	}

	return hints, nil
}

func (h *handler) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	return nil, errors.New("not implemented")
}

// Utils

func convertRange(r ast.Range) protocol.Range {
	return protocol.Range{
		Start: convertPos(r.Start),
		End:   convertPos(r.End),
	}
}

func convertRangePtr(r ast.Range) *protocol.Range {
	return &protocol.Range{
		Start: convertPos(r.Start),
		End:   convertPos(r.End),
	}
}

func convertPos(p ast.Pos) protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Line - 1),
		Character: uint32(p.Column),
	}
}
