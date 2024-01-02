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

	name   string
	detail string

	range_         core.Range
	selectionRange core.Range
}

type symbolConsumer interface {
	add(symbol symbol, childrenCount int) int
	addChild(parent int, child symbol)

	supportsDetail() bool
}

func getSymbols(symbols symbolConsumer, files []*workspace.File) {
	// Find method count per struct
	methodCount := make(map[ast.Type]int)

	for _, file := range files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil {
				methodCount[impl.Type] += len(impl.Methods)
			}
		}
	}

	// Structs
	structs := make(map[ast.Type]int)

	for _, file := range files {
		for _, decl := range file.Ast.Decls {
			if struct_, ok := decl.(*ast.Struct); ok && struct_.Cst() != nil && struct_.Name.Cst() != nil {
				id := symbols.add(symbol{
					kind:           protocol.SymbolKindStruct,
					name:           struct_.Name.String(),
					range_:         struct_.Cst().Range,
					selectionRange: struct_.Name.Cst().Range,
					file:           file,
				}, len(struct_.StaticFields)+len(struct_.Fields)+methodCount[struct_])

				for _, field := range struct_.StaticFields {
					if field.Name.Cst() != nil {
						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindField,
							name:           field.Name.String(),
							detail:         ast.PrintType(field.Type),
							range_:         field.Name.Cst().Range,
							selectionRange: field.Name.Cst().Range,
						})
					}
				}

				for _, field := range struct_.Fields {
					if field.Cst() != nil {
						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindField,
							name:           field.Name.String(),
							detail:         ast.PrintType(field.Type),
							range_:         field.Name.Cst().Range,
							selectionRange: field.Name.Cst().Range,
						})
					}
				}

				structs[struct_] = id
			}
		}
	}

	// Rest
	for _, file := range files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil {
				// Methods
				id, ok := structs[impl.Type]

				if !ok {
					id = symbols.add(symbol{
						kind: protocol.SymbolKindStruct,
						name: impl.Type.(*ast.Struct).Name.String(),
					}, methodCount[impl.Type])

					structs[impl.Type] = id
				}

				for _, function := range impl.Methods {
					if function.Cst() != nil && function.Name.Cst() != nil {
						detail := ""

						if symbols.supportsDetail() {
							detail = function.Signature(true)
						}

						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindMethod,
							name:           function.Name.String(),
							detail:         detail,
							range_:         function.Cst().Range,
							selectionRange: function.Name.Cst().Range,
						})
					}
				}
			} else if enum, ok := decl.(*ast.Enum); ok && enum.Cst() != nil && enum.Name.Cst() != nil {
				// Enum
				id := symbols.add(symbol{
					kind:           protocol.SymbolKindEnum,
					name:           enum.Name.String(),
					range_:         enum.Cst().Range,
					selectionRange: enum.Name.Cst().Range,
					file:           file,
				}, len(enum.Cases))

				for _, case_ := range enum.Cases {
					if case_.Name.Cst() != nil {
						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindEnumMember,
							name:           case_.Name.String(),
							detail:         strconv.FormatInt(case_.ActualValue, 10),
							range_:         case_.Name.Cst().Range,
							selectionRange: case_.Name.Cst().Range,
						})
					}
				}
			} else if function, ok := decl.(*ast.Func); ok && function.Cst() != nil && function.Name.Cst() != nil {
				// Function
				detail := ""

				if symbols.supportsDetail() {
					detail = function.Signature(true)
				}

				symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindFunction,
					name:           function.Name.String(),
					detail:         detail,
					range_:         function.Cst().Range,
					selectionRange: function.Name.Cst().Range,
				}, 0)
			}
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
	containerName := ""

	if parent >= 0 {
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
