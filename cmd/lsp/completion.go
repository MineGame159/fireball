package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/fuckoff"
	"fireball/core/scanner"
	"fireball/core/utils"
	"github.com/MineGame159/protocol"
	"strconv"
)

func getCompletions(resolver fuckoff.Resolver, node ast.Node, pos core.Pos) *protocol.CompletionList {
	c := completions{}

	// Leaf
	leaf := ast.GetLeaf(node, pos)

	if leaf != nil {
		if isInFunctionBody(pos, leaf) {
			switch parent := leaf.Parent().(type) {
			case *ast.Member:
				if isAfterNode(pos, parent.Value) {
					getMemberCompletions(resolver, &c, parent)
				} else {
					getIdentifierCompletions(resolver, &c, pos, leaf)
				}

			default:
				getIdentifierCompletions(resolver, &c, pos, leaf)
			}
		} else {
			getTypeCompletions(resolver, &c, pos, leaf.Parent())
		}
	} else {
		// Non leaf
		node = ast.Get(node, pos)

		if isInFunctionBody(pos, node) {
			switch node := node.(type) {
			case *ast.Member:
				if isAfterNode(pos, node.Value) {
					getMemberCompletions(resolver, &c, node)
				} else {
					getIdentifierCompletions(resolver, &c, pos, leaf)
				}

			case *ast.StructInitializer:
				if isAfterCst(pos, node, scanner.LeftBrace, false) {
					if s, ok := ast.As[*ast.Struct](node.Type); ok {
						for _, field := range s.Fields {
							c.addNode(protocol.CompletionItemKindField, field.Name, printType(field.Type))
						}
					}
				}

			default:
				getIdentifierCompletions(resolver, &c, pos, node)
			}
		} else {
			getTypeCompletions(resolver, &c, pos, node)
		}
	}

	// Return
	return c.get()
}

func getTypeCompletions(resolver fuckoff.Resolver, c *completions, pos core.Pos, node ast.Node) {
	switch node := node.(type) {
	case *ast.Struct:
		for _, field := range node.Fields {
			if field.Type == nil && isAfterNode(pos, field.Name) {
				getGlobalCompletions(resolver, c, false)
			}
		}

	case *ast.Impl:
		if isAfterCst(pos, node, scanner.Impl, true) {
			getGlobalCompletions(resolver, c, false)
		}

	case *ast.Enum:
		if isAfterCst(pos, node, scanner.Colon, true) {
			getGlobalCompletions(resolver, c, false)
		}

	case *ast.Func:
		for _, param := range node.Params {
			if param.Type == nil && isAfterNode(pos, param.Name) {
				getGlobalCompletions(resolver, c, false)
			}
		}

		if isAfterCst(pos, node, scanner.RightParen, true) {
			getGlobalCompletions(resolver, c, false)
		}

	case *ast.Field:
		if isAfterNode(pos, node.Name) {
			getGlobalCompletions(resolver, c, false)
		}

	case *ast.Param:
		if isAfterNode(pos, node.Name) {
			getGlobalCompletions(resolver, c, false)
		}

	case ast.Type:
		if !isComplexType(node) {
			getGlobalCompletions(resolver, c, false)
		}
	}
}

func getMemberCompletions(resolver fuckoff.Resolver, c *completions, member *ast.Member) {
	if s, ok := asThroughPointer[*ast.Struct](member.Value.Result().Type); ok {
		fields := s.Fields
		static := false

		if member.Value.Result().Kind == ast.TypeResultKind {
			fields = s.StaticFields
			static = true
		}

		for _, field := range fields {
			c.addNode(protocol.CompletionItemKindField, field.Name, printType(field.Type))
		}

		for _, method := range resolver.GetMethods(s, static) {
			c.addNode(protocol.CompletionItemKindMethod, method.Name, printType(method))
		}
	} else if e, ok := asThroughPointer[*ast.Enum](member.Value.Result().Type); ok {
		for _, case_ := range e.Cases {
			c.addNode(protocol.CompletionItemKindEnumMember, case_.Name, strconv.FormatInt(case_.ActualValue, 10))
		}
	}
}

func getIdentifierCompletions(resolver fuckoff.Resolver, c *completions, pos core.Pos, node ast.Node) {
	// Types and global functions
	getGlobalCompletions(resolver, c, true)

	// Variables
	function := ast.GetParent[*ast.Func](node)

	if function != nil {
		names := utils.NewSet[string]()

		// This
		if s := function.Method(); s != nil {
			c.add(protocol.CompletionItemKindVariable, "this", printType(s))
		}

		// Parameters
		for _, param := range function.Params {
			if param.Name != nil && !names.Contains(param.Name.String()) {
				c.addNode(protocol.CompletionItemKindVariable, param.Name, printType(param.Type))
				names.Add(param.Name.String())
			}
		}

		// Variables
		varResolver := variableResolver{target: pos}
		if !node.Token().IsEmpty() {
			varResolver.targetVariableName = node
		}

		varResolver.VisitNode(function)

		for i := len(varResolver.variables) - 1; i >= 0; i-- {
			variable := varResolver.variables[i]
			add := true

			if parent, ok := node.Parent().(*ast.Var); ok {
				add = parent != variable
			}

			if add && variable.Name != nil && !names.Contains(variable.Name.String()) {
				c.addNode(protocol.CompletionItemKindVariable, variable.Name, printType(variable.ActualType))
				names.Add(variable.Name.String())
			}
		}
	}
}

func getGlobalCompletions(resolver fuckoff.Resolver, c *completions, functions bool) {
	// Primitive types
	c.add(protocol.CompletionItemKindStruct, "void", "")
	c.add(protocol.CompletionItemKindStruct, "bool", "")

	c.add(protocol.CompletionItemKindStruct, "u8", "")
	c.add(protocol.CompletionItemKindStruct, "u16", "")
	c.add(protocol.CompletionItemKindStruct, "u32", "")
	c.add(protocol.CompletionItemKindStruct, "u64", "")

	c.add(protocol.CompletionItemKindStruct, "i8", "")
	c.add(protocol.CompletionItemKindStruct, "i16", "")
	c.add(protocol.CompletionItemKindStruct, "i32", "")
	c.add(protocol.CompletionItemKindStruct, "i64", "")

	c.add(protocol.CompletionItemKindStruct, "f32", "")
	c.add(protocol.CompletionItemKindStruct, "f64", "")

	// True, false
	c.add(protocol.CompletionItemKindKeyword, "true", "bool")
	c.add(protocol.CompletionItemKindKeyword, "false", "bool")

	// Language defined types and functions
	for _, file := range resolver.GetFileNodes() {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.Struct:
				c.addNode(protocol.CompletionItemKindStruct, decl.Name, "")

			case *ast.Enum:
				c.addNode(protocol.CompletionItemKindEnum, decl.Name, "")

			case *ast.Func:
				if functions {
					c.addNode(protocol.CompletionItemKindFunction, decl.Name, printType(decl))
				}
			}
		}
	}
}

// Utils

func isComplexType(type_ ast.Type) bool {
	switch type_.(type) {
	case *ast.Struct, *ast.Enum, *ast.Func:
		return true

	default:
		return false
	}
}

func isAfterNode(pos core.Pos, node ast.Node) bool {
	return !ast.IsNil(node) && node.Cst() != nil && pos.Line == node.Cst().Range.End.Line && pos.Column > node.Cst().Range.End.Column
}

func isAfterCst(pos core.Pos, node ast.Node, kind scanner.TokenKind, sameLine bool) bool {
	if node.Cst() == nil {
		return false
	}

	child := node.Cst().Get(kind)
	if child == nil {
		return false
	}

	after := child.Range.End

	if sameLine {
		return pos.Line == after.Line && pos.Column > after.Column
	}

	return pos.IsAfter(after)
}

func isInFunctionBody(pos core.Pos, node ast.Node) bool {
	function := ast.GetParent[*ast.Func](node)
	if function == nil || function.Cst() == nil {
		return false
	}

	left := function.Cst().Get(scanner.LeftBrace)
	if left == nil {
		return false
	}

	right := function.Cst().Get(scanner.RightBrace)
	if right == nil {
		return false
	}

	return core.Range{Start: left.Range.Start, End: right.Range.End}.Contains(pos)
}

func printType(type_ ast.Type) string {
	return ast.PrintTypeOptions(type_, ast.TypePrintOptions{ParamNames: true})
}

// Completions

type completions struct {
	items []protocol.CompletionItem
}

var commitCharacters = []string{".", ";"}
var commitCharactersFunction = []string{".", ";", "("}

func (c *completions) addNode(kind protocol.CompletionItemKind, name ast.Node, detail string) {
	if !ast.IsNil(name) {
		c.add(kind, name.String(), detail)
	}
}

func (c *completions) add(kind protocol.CompletionItemKind, name, detail string) {
	characters := commitCharacters
	if kind == protocol.CompletionItemKindFunction || kind == protocol.CompletionItemKindMethod {
		characters = commitCharactersFunction
	}

	c.items = append(c.items, protocol.CompletionItem{
		Kind:             kind,
		Label:            name,
		Detail:           detail,
		CommitCharacters: characters,
	})
}

func (c *completions) get() *protocol.CompletionList {
	if len(c.items) == 0 {
		return nil
	}

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        c.items,
	}
}
