package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
	"path/filepath"
	"strconv"
)

type symbol struct {
	file *workspace.File
	kind protocol.SymbolKind

	name          string
	containerName string
	detail        string

	range_         core.Range
	selectionRange core.Range
}

type symbolConsumer interface {
	add(symbol symbol, childrenCount int) int
	addChild(parent int, child symbol)

	supportsDetail() bool
}

func getSymbols(symbols symbolConsumer, file *workspace.File) {
	for _, decl := range file.Decls {
		if struct_, ok := decl.(*ast.Struct); ok {
			// Struct
			parent := symbols.add(symbol{
				kind:           protocol.SymbolKindStruct,
				name:           struct_.Name.Lexeme,
				detail:         "",
				range_:         struct_.Range(),
				selectionRange: core.TokenToRange(struct_.Name),
				file:           file,
			}, len(struct_.Fields))

			for _, field := range struct_.Fields {
				range_ := core.TokenToRange(field.Name)

				symbols.addChild(parent, symbol{
					file:           file,
					kind:           protocol.SymbolKindField,
					name:           field.Name.Lexeme,
					detail:         field.Type.String(),
					range_:         range_,
					selectionRange: range_,
				})
			}
		} else if impl, ok := decl.(*ast.Impl); ok {
			for _, f := range impl.Functions {
				function := f.(*ast.Func)
				detail := ""

				if symbols.supportsDetail() {
					detail = function.Signature(true)
				}

				symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindMethod,
					name:           function.Name.Lexeme,
					containerName:  impl.Struct.Lexeme,
					detail:         detail,
					range_:         function.Range(),
					selectionRange: core.TokenToRange(function.Name),
				}, 0)
			}
		} else if enum, ok := decl.(*ast.Enum); ok {
			// Enum
			parent := symbols.add(symbol{
				kind:           protocol.SymbolKindEnum,
				name:           enum.Name.Lexeme,
				detail:         "",
				range_:         enum.Range(),
				selectionRange: core.TokenToRange(enum.Name),
				file:           file,
			}, len(enum.Cases))

			for _, case_ := range enum.Cases {
				range_ := core.TokenToRange(case_.Name)

				symbols.addChild(parent, symbol{
					file:           file,
					kind:           protocol.SymbolKindEnumMember,
					name:           case_.Name.Lexeme,
					detail:         strconv.Itoa(case_.Value),
					range_:         range_,
					selectionRange: range_,
				})
			}
		} else if function, ok := decl.(*ast.Func); ok {
			// Function
			detail := ""

			if symbols.supportsDetail() {
				detail = function.Signature(true)
			}

			symbols.add(symbol{
				file:           file,
				kind:           protocol.SymbolKindFunction,
				name:           function.Name.Lexeme,
				detail:         detail,
				range_:         function.Range(),
				selectionRange: core.TokenToRange(function.Name),
			}, 0)
		}
	}
}

// Document symbols

type documentSymbolConsumer struct {
	symbols []any
}

func (d *documentSymbolConsumer) add(symbol symbol, childrenCount int) int {
	d.symbols = append(d.symbols, d.convert(symbol, childrenCount))
	return len(d.symbols) - 1
}

func (d *documentSymbolConsumer) addChild(parent int, child symbol) {
	symbol := d.symbols[parent].(protocol.DocumentSymbol)
	symbol.Children = append(symbol.Children, d.convert(child, 0))

	d.symbols[parent] = symbol
}

func (d *documentSymbolConsumer) supportsDetail() bool {
	return true
}

func (d *documentSymbolConsumer) convert(symbol symbol, childrenCount int) protocol.DocumentSymbol {
	var children []protocol.DocumentSymbol

	if childrenCount > 0 {
		children = make([]protocol.DocumentSymbol, 0, childrenCount)
	}

	return protocol.DocumentSymbol{
		Name:           symbol.name,
		Detail:         symbol.detail,
		Kind:           symbol.kind,
		Range:          convertRange(symbol.range_),
		SelectionRange: convertRange(symbol.selectionRange),
		Children:       children,
	}
}

// Workspace symbols

type workspaceSymbolConsumer struct {
	symbols []protocol.SymbolInformation
}

func (w *workspaceSymbolConsumer) add(symbol symbol, _ int) int {
	w.symbols = append(w.symbols, w.convert(symbol, -1))
	return len(w.symbols) - 1
}

func (w *workspaceSymbolConsumer) addChild(parent int, child symbol) {
	w.symbols = append(w.symbols, w.convert(child, parent))
}

func (w *workspaceSymbolConsumer) supportsDetail() bool {
	return false
}

func (w *workspaceSymbolConsumer) convert(symbol symbol, parent int) protocol.SymbolInformation {
	containerName := symbol.containerName

	if parent >= 0 && containerName == "" {
		containerName = w.symbols[parent].Name
	}

	return protocol.SymbolInformation{
		Name:       symbol.name,
		Kind:       symbol.kind,
		Tags:       nil,
		Deprecated: false,
		Location: protocol.Location{
			URI:   uri.New(filepath.Join(symbol.file.Project.Path, symbol.file.Path)),
			Range: convertRange(symbol.range_),
		},
		ContainerName: containerName,
	}
}
