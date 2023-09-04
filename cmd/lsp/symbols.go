package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
	"path/filepath"
	"strconv"
	"strings"
)

type symbol struct {
	file *workspace.File
	kind protocol.SymbolKind

	name   string
	detail string

	range_         core.Range
	selectionRange core.Range
}

type symbolConsumer interface {
	add(symbol symbol, children []symbol)
}

func getSymbols(symbols symbolConsumer, file *workspace.File) {
	for _, decl := range file.Decls {
		if v, ok := decl.(*ast.Struct); ok {
			// Struct
			children := make([]symbol, len(v.Fields))

			for i, field := range v.Fields {
				range_ := core.TokenToRange(field.Name)

				children[i] = symbol{
					file:           file,
					kind:           protocol.SymbolKindField,
					name:           field.Name.Lexeme,
					detail:         field.Type.String(),
					range_:         range_,
					selectionRange: range_,
				}
			}

			symbols.add(symbol{
				kind:           protocol.SymbolKindStruct,
				name:           v.Name.Lexeme,
				detail:         "",
				range_:         v.Range(),
				selectionRange: core.TokenToRange(v.Name),
				file:           file,
			}, children)
		} else if v, ok := decl.(*ast.Enum); ok {
			// Enum
			children := make([]symbol, len(v.Cases))

			for i, case_ := range v.Cases {
				range_ := core.TokenToRange(case_.Name)

				children[i] = symbol{
					file:           file,
					kind:           protocol.SymbolKindEnumMember,
					name:           case_.Name.Lexeme,
					detail:         strconv.Itoa(case_.Value),
					range_:         range_,
					selectionRange: range_,
				}
			}

			symbols.add(symbol{
				kind:           protocol.SymbolKindEnum,
				name:           v.Name.Lexeme,
				detail:         "",
				range_:         v.Range(),
				selectionRange: core.TokenToRange(v.Name),
				file:           file,
			}, children)
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

			symbols.add(symbol{
				file:           file,
				kind:           protocol.SymbolKindFunction,
				name:           v.Name.Lexeme,
				detail:         signature.String(),
				range_:         v.Range(),
				selectionRange: core.TokenToRange(v.Name),
			}, nil)
		}
	}
}

// Document symbols

type documentSymbolConsumer struct {
	symbols []any
}

func (d *documentSymbolConsumer) add(symbol symbol, children []symbol) {
	lspSymbol := d.convert(symbol)

	if children != nil {
		lspSymbol.Children = make([]protocol.DocumentSymbol, len(children))

		for i, child := range children {
			lspSymbol.Children[i] = d.convert(child)
		}
	}

	d.symbols = append(d.symbols, lspSymbol)
}

func (d *documentSymbolConsumer) convert(symbol symbol) protocol.DocumentSymbol {
	return protocol.DocumentSymbol{
		Name:           symbol.name,
		Detail:         symbol.detail,
		Kind:           symbol.kind,
		Range:          convertRange(symbol.range_),
		SelectionRange: convertRange(symbol.selectionRange),
	}
}

// Workspace symbols

type workspaceSymbolConsumer struct {
	symbols []protocol.SymbolInformation
}

func (d *workspaceSymbolConsumer) add(symbol symbol, children []symbol) {
	d.symbols = append(d.symbols, d.convert(symbol, ""))

	for _, child := range children {
		d.symbols = append(d.symbols, d.convert(child, symbol.name))
	}
}

func (d *workspaceSymbolConsumer) convert(symbol symbol, parent string) protocol.SymbolInformation {
	return protocol.SymbolInformation{
		Name:       symbol.name,
		Kind:       symbol.kind,
		Tags:       nil,
		Deprecated: false,
		Location: protocol.Location{
			URI:   uri.New(filepath.Join(symbol.file.Project.Path, symbol.file.Path)),
			Range: convertRange(symbol.range_),
		},
		ContainerName: parent,
	}
}
