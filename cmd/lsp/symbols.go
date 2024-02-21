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
			if struct_, ok := decl.(*ast.Struct); ok {
				if struct_.Cst() == nil || nodeCst(struct_.Name) == nil {
					continue
				}

				id := symbols.add(symbol{
					kind:           protocol.SymbolKindStruct,
					name:           struct_.Name.String(),
					range_:         struct_.Cst().Range,
					selectionRange: nodeCst(struct_.Name).Range,
					file:           file,
				}, len(struct_.StaticFields)+len(struct_.Fields)+methodCount[struct_])

				for _, field := range struct_.StaticFields {
					if nodeCst(field.Name()) == nil {
						continue
					}

					symbols.addChild(id, symbol{
						file:           file,
						kind:           protocol.SymbolKindField,
						name:           field.Name().String(),
						detail:         ast.PrintType(field.Type()),
						range_:         nodeCst(field.Name()).Range,
						selectionRange: nodeCst(field.Name()).Range,
					})
				}

				for _, field := range struct_.Fields {
					if nodeCst(field.Name()) == nil {
						continue
					}

					symbols.addChild(id, symbol{
						file:           file,
						kind:           protocol.SymbolKindField,
						name:           field.Name().String(),
						detail:         ast.PrintType(field.Type()),
						range_:         nodeCst(field.Name()).Range,
						selectionRange: nodeCst(field.Name()).Range,
					})
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
					if nodeCst(function) != nil && nodeCst(function.Name) != nil {
						detail := ""

						if symbols.supportsDetail() {
							detail = ast.Signature(function, true)
						}

						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindMethod,
							name:           function.Name.String(),
							detail:         detail,
							range_:         nodeCst(function).Range,
							selectionRange: nodeCst(function.Name).Range,
						})
					}
				}
			} else if enum, ok := decl.(*ast.Enum); ok && nodeCst(enum) != nil && nodeCst(enum.Name) != nil {
				// Enum
				id := symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindEnum,
					name:           enum.Name.String(),
					range_:         nodeCst(enum).Range,
					selectionRange: nodeCst(enum.Name).Range,
				}, len(enum.Cases))

				for _, case_ := range enum.Cases {
					if nodeCst(case_.Name) != nil {
						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindEnumMember,
							name:           case_.Name.String(),
							detail:         strconv.FormatInt(case_.ActualValue, 10),
							range_:         nodeCst(case_.Name).Range,
							selectionRange: nodeCst(case_.Name).Range,
						})
					}
				}
			} else if inter, ok := decl.(*ast.Interface); ok && nodeCst(inter) != nil && nodeCst(inter.Name) != nil {
				// Interface
				id := symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindInterface,
					name:           inter.Name.String(),
					range_:         nodeCst(inter).Range,
					selectionRange: nodeCst(inter.Name).Range,
				}, len(inter.Methods))

				for _, method := range inter.Methods {
					if nodeCst(method.Name) != nil {
						detail := ""

						if symbols.supportsDetail() {
							detail = ast.Signature(method, true)
						}

						symbols.addChild(id, symbol{
							file:           file,
							kind:           protocol.SymbolKindMethod,
							name:           method.Name.String(),
							detail:         detail,
							range_:         nodeCst(method).Range,
							selectionRange: nodeCst(method.Name).Range,
						})
					}
				}
			} else if function, ok := decl.(*ast.Func); ok && nodeCst(function) != nil && nodeCst(function.Name) != nil {
				// Function
				detail := ""

				if symbols.supportsDetail() {
					detail = ast.Signature(function, true)
				}

				symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindFunction,
					name:           function.Name.String(),
					detail:         detail,
					range_:         nodeCst(function).Range,
					selectionRange: nodeCst(function.Name).Range,
				}, 0)
			} else if variable, ok := decl.(*ast.GlobalVar); ok && nodeCst(variable) != nil && nodeCst(variable.Name) != nil {
				symbols.add(symbol{
					file:           file,
					kind:           protocol.SymbolKindVariable,
					name:           variable.Name.String(),
					detail:         printType(variable.Type),
					range_:         nodeCst(variable).Range,
					selectionRange: nodeCst(variable.Name).Range,
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
