package lsp

import (
	"context"
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
	"fireball/core/utils"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type handler struct {
	logger *zap.Logger
	client protocol.Client

	projects  []*workspace.Project
	documents map[uri.URI]*workspace.File
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

func newHandler() *handler {
	return &handler{
		projects:  make([]*workspace.Project, 0),
		documents: make(map[uri.URI]*workspace.File),
	}
}

func (h *handler) Initialize(_ context.Context, params *protocol.InitializeParams) (result *protocol.InitializeResult, err error) {
	defer stop(start(h, "Initialize"))

	// Create projects
	for _, folder := range params.WorkspaceFolders {
		// Parse URI
		uri_, err := uri.Parse(folder.URI)
		if err != nil {
			continue
		}

		// Create project
		project, err := workspace.NewProject(uri_.Filename())
		if err != nil {
			continue
		}

		// Load files
		err = project.LoadFiles()
		if err != nil {
			continue
		}

		// Add project to handler's list
		h.projects = append(h.projects, project)
	}

	// Return server capabilities
	toWatch := &protocol.FileOperationRegistrationOptions{
		Filters: make([]protocol.FileOperationFilter, len(h.projects)),
	}

	for i, project := range h.projects {
		toWatch.Filters[i] = protocol.FileOperationFilter{
			Scheme: "file",
			Pattern: protocol.FileOperationPattern{
				Glob:    filepath.Join(project.Path, project.Config.Src) + "/**/*.fb",
				Matches: protocol.FileOperationPatternKindFile,
				Options: protocol.FileOperationPatternOptions{},
			},
		}
	}

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
						protocol.SemanticTokenType,
						protocol.SemanticTokenClass,
						protocol.SemanticTokenEnum,
						protocol.SemanticTokenProperty,
						protocol.SemanticTokenEnumMember,
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

			Workspace: &protocol.ServerCapabilitiesWorkspace{
				FileOperations: &protocol.ServerCapabilitiesWorkspaceFileOperations{
					DidCreate: toWatch,
					DidDelete: toWatch,
				},
			},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "fireball",
			Version: "0.1.0",
		},
	}, nil
}

func (h *handler) Initialized(ctx context.Context, _ *protocol.InitializedParams) (err error) {
	defer stop(start(h, "Initialized"))

	// Publish diagnotics for all projects
	for _, project := range h.projects {
		err := h.publishDiagnostics(ctx, project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) Shutdown(_ context.Context) (err error) {
	defer stop(start(h, "Shutdown"))

	return nil
}

func (h *handler) Exit(_ context.Context) (err error) {
	defer stop(start(h, "Exit"))

	return nil
}

func (h *handler) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (err error) {
	defer stop(start(h, "DidCreateFiles"))

	for _, createdFile := range params.Files {
		// Parse URI
		uri_, err := uri.Parse(createdFile.URI)
		if err != nil {
			continue
		}

		// Get project
		path := uri_.Filename()

		var project *workspace.Project
		relative := ""

		for _, proj := range h.projects {
			rel, err := filepath.Rel(proj.Path, path)
			if err != nil {
				continue
			}

			project = proj
			relative = rel

			break
		}

		if project == nil {
			continue
		}

		// Read file
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Create file and parse
		file := project.GetOrCreateFile(relative)
		file.SetText(string(contents), true)

		// Publish diagnostics
		err = h.publishDiagnostics(ctx, project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (err error) {
	defer stop(start(h, "DidDeleteFiles"))

	for _, createdFile := range params.Files {
		// Parse URI
		uri_, err := uri.Parse(createdFile.URI)
		if err != nil {
			continue
		}

		// Delete file
		path := uri_.Filename()

		for _, project := range h.projects {
			if project.RemoveFileAbs(path) {
				// Publish diagnostics
				return h.publishDiagnostics(ctx, project)
			}
		}
	}

	return nil
}

func (h *handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	defer stop(start(h, "DidOpen"))

	// Check if file belongs to some open project
	path := params.TextDocument.URI.Filename()

	for _, project := range h.projects {
		file := project.GetFileAbs(path)

		if file != nil {
			h.documents[params.TextDocument.URI] = file
			return nil
		}
	}

	// Create an empty project for this one file
	name := filepath.Base(path)
	project := workspace.NewEmptyProject(filepath.Dir(path), name[:strings.LastIndexByte(name, '.')])

	relative, err := filepath.Rel(project.Path, path)
	if err != nil {
		return err
	}

	file := project.GetOrCreateFile(relative)
	file.SetText(params.TextDocument.Text, true)

	h.documents[params.TextDocument.URI] = file

	// Publish diagnostics
	return h.publishDiagnostics(ctx, project)
}

func (h *handler) DidClose(_ context.Context, params *protocol.DidCloseTextDocumentParams) (err error) {
	defer stop(start(h, "DidClose"))

	delete(h.documents, params.TextDocument.URI)

	return nil
}

func (h *handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	defer stop(start(h, "DidChange"))

	// Get file
	file := h.getFile(params.TextDocument.URI)
	if file == nil {
		return nil
	}

	// Mark document as dirty
	doc := getDocument(file)
	doc.setDirty()

	// Set text and parse
	file.SetText(params.ContentChanges[0].Text, true)

	// Publish diagnostics
	return h.publishDiagnostics(ctx, file.Project)
}

func (h *handler) SemanticTokensFull(_ context.Context, params *protocol.SemanticTokensParams) (result *protocol.SemanticTokens, err error) {
	defer stop(start(h, "SemanticTokensFull"))

	// Get document
	file := h.getFile(params.TextDocument.URI)
	if file == nil {
		return nil, nil
	}

	file.EnsureChecked()

	// Check cached
	doc := getDocument(file)

	if doc.semanticTokens != nil {
		return doc.semanticTokens, nil
	}

	// Get semantic tokens
	data := highlight(file.Decls)

	tokens := &protocol.SemanticTokens{
		Data: data,
	}

	doc.semanticTokens = tokens
	return tokens, nil
}

func (h *handler) DocumentSymbol(_ context.Context, params *protocol.DocumentSymbolParams) (result []interface{}, err error) {
	defer stop(start(h, "DocumentSymbol"))

	// Get file
	file := h.getFile(params.TextDocument.URI)
	if file == nil {
		return nil, nil
	}

	file.EnsureParsed()

	// GetLeaf symbols
	symbols := make([]interface{}, 0, 8)

	for _, decl := range file.Decls {
		if v, ok := decl.(*ast.Struct); ok {
			// Struct
			symbols = append(symbols, &protocol.DocumentSymbol{
				Name:           v.Name.Lexeme,
				Kind:           protocol.SymbolKindStruct,
				Range:          convertRange(v.Range()),
				SelectionRange: convertRange(core.TokenToRange(v.Name)),
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
				SelectionRange: convertRange(core.TokenToRange(v.Name)),
			})
		}
	}

	return symbols, nil
}

func (h *handler) InlayHint(_ context.Context, params *protocol.InlayHintParams) (result []protocol.InlayHint, err error) {
	defer stop(start(h, "InlayHint"))

	// Get document
	file := h.getFile(params.TextDocument.URI)
	if file == nil {
		return nil, nil
	}

	file.EnsureChecked()

	// Check cache
	doc := getDocument(file)

	if doc.inlayHints != nil {
		return doc.inlayHints, nil
	}

	// Get hints
	hints := annotate(file.Decls)

	doc.inlayHints = hints
	return hints, nil
}

func (h *handler) Hover(_ context.Context, params *protocol.HoverParams) (result *protocol.Hover, err error) {
	defer stop(start(h, "Hover"))

	// Get document
	file := h.getFile(params.TextDocument.URI)
	if file == nil {
		return nil, nil
	}

	file.EnsureChecked()

	// Convert position
	pos := core.Pos{
		Line:   int(params.Position.Line + 1),
		Column: int(params.Position.Character),
	}

	// Get hover
	return getHover(file.Decls, pos), nil
}

// Utils

type request struct {
	h     *handler
	name  string
	start time.Time
}

func start(h *handler, name string) request {
	return request{
		h:     h,
		name:  name,
		start: time.Now(),
	}
}

func stop(req request) {
	duration := time.Now().Sub(req.start)
	req.h.logger.Debug(req.name, zap.Duration("duration", duration))
}

func (h *handler) publishDiagnostics(ctx context.Context, project *workspace.Project) error {
	// Loop all files
	for _, file := range project.Files {
		// Flush
		fbDiagnostics := file.FlushDiagnostics()

		// Check if diagnostics need to be sent
		doc := getDocument(file)

		if len(fbDiagnostics) == 0 && !doc.hasDiagnostics {
			continue
		}

		// Convert
		lspDiagnostics := make([]protocol.Diagnostic, len(fbDiagnostics))

		for i, diagnostic := range fbDiagnostics {
			severity := protocol.DiagnosticSeverityError

			if diagnostic.Kind == utils.WarningKind {
				severity = protocol.DiagnosticSeverityWarning
			}

			lspDiagnostics[i] = protocol.Diagnostic{
				Range:    convertRange(diagnostic.Range),
				Severity: severity,
				Message:  diagnostic.Message,
			}
		}

		// Send
		err := h.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         uri.New(filepath.Join(project.Path, file.Path)),
			Diagnostics: lspDiagnostics,
		})

		if err != nil {
			return err
		}

		doc.hasDiagnostics = len(lspDiagnostics) > 0
	}

	return nil
}

func getDocument(file *workspace.File) *document {
	// Check data field on file
	if doc, ok := file.Data.(*document); ok {
		return doc
	}

	// Create and set new document on file
	doc := &document{}

	file.Data = doc
	return doc
}

func (h *handler) getFile(uri_ uri.URI) *workspace.File {
	if file, ok := h.documents[uri_]; ok {
		return file
	}

	return nil
}

// Conversions

func convertRange(r core.Range) protocol.Range {
	return protocol.Range{
		Start: convertPos(r.Start),
		End:   convertPos(r.End),
	}
}

func convertRangePtr(r core.Range) *protocol.Range {
	return &protocol.Range{
		Start: convertPos(r.Start),
		End:   convertPos(r.End),
	}
}

func convertPos(p core.Pos) protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Line - 1),
		Character: uint32(p.Column),
	}
}
