package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"strconv"
)

func getCompletions(project *workspace.Project, file *ast.File, pos core.Pos) *protocol.CompletionList {
	baseResolver := project.GetResolverFile(file)
	if baseResolver == nil {
		return nil
	}

	c := completions{}
	resolver := ast.NewCombinedResolver(baseResolver)

	for _, decl := range file.Decls {
		if using, ok := decl.(*ast.Using); ok {
			if r := baseResolver.GetResolver(using.Name); r != nil {
				resolver.Add(r)
			}
		}
	}

	// Leaf
	leaf := ast.GetLeaf(file, pos)

	if leaf != nil {
		if isInFunctionBody(pos, leaf) {
			switch parent := leaf.Parent().(type) {
			case *ast.Resolvable:
				getResolvableCompletions(resolver, &c, pos, parent)

			case *ast.Member:
				if isAfterNode(pos, parent.Value) {
					getMemberCompletions(resolver, &c, parent)
				} else {
					getIdentifierCompletions(resolver, &c, pos, leaf)
				}

			default:
				getStmtCompletions(&c, parent)
				getIdentifierCompletions(resolver, &c, pos, leaf)
			}
		} else {
			getTypeCompletions(project.GetResolverRoot(), resolver, &c, pos, leaf.Parent())
		}
	} else {
		// Non leaf
		node := ast.Get(file, pos)

		if isInFunctionBody(pos, node) {
			switch node := node.(type) {
			case *ast.Resolvable:
				getResolvableCompletions(resolver, &c, pos, node)

			case *ast.Var:
				if isAfterNode(pos, node.Name) && isBeforeCst(pos, node, scanner.Equal) {
					getGlobalCompletions(resolver, &c, true)
				} else if isAfterCst(pos, node, scanner.Equal, false) {
					getIdentifierCompletions(resolver, &c, pos, node)
				}

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
				getStmtCompletions(&c, node)
				getIdentifierCompletions(resolver, &c, pos, node)
			}
		} else {
			getTypeCompletions(project.GetResolverRoot(), resolver, &c, pos, node)
		}
	}

	// Return
	return c.get()
}

func getResolvableCompletions(resolver ast.Resolver, c *completions, pos core.Pos, resolvable *ast.Resolvable) {
	if resolvable.Cst() == nil || !resolvable.Cst().Contains(scanner.Dot) {
		getGlobalCompletions(resolver, c, true)
		return
	}

	for _, part := range resolvable.Parts {
		if part.Cst() != nil && pos.IsAfter(part.Cst().Range.End) && resolver != nil {
			resolver = resolver.GetChild(part.String())
		}
	}

	if resolver != nil {
		getResolverCompletions(c, resolver, true)
	}
}

func getTypeCompletions(root ast.RootResolver, resolver ast.Resolver, c *completions, pos core.Pos, node ast.Node) {
	switch node := node.(type) {
	case *ast.NamespaceName:
		getNamespaceCompletions(root, c, pos, node)

	case *ast.Struct:
		for _, field := range node.Fields {
			if field.Type == nil && isAfterNode(pos, field.Name) {
				getGlobalCompletions(resolver, c, true)
			}
		}

	case *ast.Field:
		if isAfterNode(pos, node.Name) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.Impl:
		if isAfterCst(pos, node, scanner.Impl, true) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.Enum:
		if isAfterCst(pos, node, scanner.Colon, true) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.Func:
		for _, param := range node.Params {
			if param.Type == nil && isAfterNode(pos, param.Name) {
				getGlobalCompletions(resolver, c, true)
			}
		}

		if isAfterCst(pos, node, scanner.RightParen, true) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.Param:
		if isAfterNode(pos, node.Name) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.GlobalVar:
		if isAfterNode(pos, node.Name) {
			getGlobalCompletions(resolver, c, true)
		}

	case *ast.Resolvable:
		getResolvableCompletions(resolver, c, pos, node)

	case ast.Type:
		if !isComplexType(node) {
			getGlobalCompletions(resolver, c, true)
		}
	}
}

func getNamespaceCompletions(root ast.RootResolver, c *completions, pos core.Pos, node *ast.NamespaceName) {
	var resolver ast.Resolver = root

	for _, part := range node.Parts {
		if part.Cst().Range.Contains(pos) {
			break
		}

		resolver = root.GetChild(part.String())
	}

	if resolver != nil {
		for _, child := range resolver.GetChildren() {
			c.add(protocol.CompletionItemKindModule, child, "")
		}
	}
}

func getMemberCompletions(resolver ast.Resolver, c *completions, member *ast.Member) {
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch member.Value.Result().Kind {
	case ast.ResolverResultKind:
		resolver := member.Value.Result().Resolver()
		getResolverCompletions(c, resolver, false)

	case ast.TypeResultKind, ast.ValueResultKind:
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
		} else if e, ok := asThroughPointer[*ast.Enum](member.Value.Result().Type); ok && member.Value.Result().Kind == ast.TypeResultKind {
			for _, case_ := range e.Cases {
				c.addNode(protocol.CompletionItemKindEnumMember, case_.Name, strconv.FormatInt(case_.ActualValue, 10))
			}
		}
	}
}

func getIdentifierCompletions(resolver ast.Resolver, c *completions, pos core.Pos, node ast.Node) {
	// Types and global functions
	getGlobalCompletions(resolver, c, false)

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

func getStmtCompletions(c *completions, node ast.Node) {
	ok := false

	switch node := node.(type) {
	case *ast.Func:
		ok = true

	case *ast.Expression:
		switch node.Parent().(type) {
		case *ast.Func, *ast.Block:
			ok = true
		}
	}

	if ok {
		c.add(protocol.CompletionItemKindKeyword, "var", "")
		c.add(protocol.CompletionItemKindSnippet, "if", "")
		c.add(protocol.CompletionItemKindSnippet, "while", "")
		c.add(protocol.CompletionItemKindSnippet, "for", "")
		c.add(protocol.CompletionItemKindKeyword, "return", "")
		c.add(protocol.CompletionItemKindKeyword, "break", "")
		c.add(protocol.CompletionItemKindKeyword, "continue", "")
	}
}

func getGlobalCompletions(resolver ast.Resolver, c *completions, symbolsOnlyTypes bool) {
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

	if !symbolsOnlyTypes {
		// Builtin identifiers
		c.add(protocol.CompletionItemKindKeyword, "true", "bool")
		c.add(protocol.CompletionItemKindKeyword, "false", "bool")

		c.add(protocol.CompletionItemKindFunction, "sizeof", "(<type>) u32")
		c.add(protocol.CompletionItemKindFunction, "alignof", "(<type>) u32")
	}

	// Language defined types and functions
	getResolverCompletions(c, resolver, symbolsOnlyTypes)
}

func getResolverCompletions(c *completions, resolver ast.Resolver, symbolsOnlyTypes bool) {
	for _, child := range resolver.GetChildren() {
		c.add(protocol.CompletionItemKindModule, child, "")
	}

	c.symbolsOnlyTypes = symbolsOnlyTypes
	resolver.GetSymbols(c)
}

func (c *completions) VisitSymbol(node ast.Node) {
	switch node := node.(type) {
	case *ast.Struct:
		c.addNode(protocol.CompletionItemKindStruct, node.Name, "")

	case *ast.Enum:
		c.addNode(protocol.CompletionItemKindEnum, node.Name, "")

	case *ast.Func:
		if !c.symbolsOnlyTypes {
			c.addNode(protocol.CompletionItemKindFunction, node.Name, printType(node))
		}

	case *ast.GlobalVar:
		if !c.symbolsOnlyTypes {
			c.addNode(protocol.CompletionItemKindVariable, node.Name, printType(node.Type))
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

func isBeforeCst(pos core.Pos, node ast.Node, kind scanner.TokenKind) bool {
	if node.Cst() == nil {
		return false
	}

	child := node.Cst().Get(kind)
	if child == nil {
		return true
	}

	return child.Range.Start.IsAfter(pos)
}

func isInFunctionBody(pos core.Pos, node ast.Node) bool {
	function := ast.GetParent[*ast.Func](node)
	if function == nil {
		return false
	}

	return isBetween(pos, function, scanner.LeftBrace, scanner.RightBrace)
}

// Completions

type completions struct {
	symbolsOnlyTypes bool

	items []protocol.CompletionItem
}

var commitCharacters = []string{".", ";"}

func (c *completions) addNode(kind protocol.CompletionItemKind, name ast.Node, detail string) {
	if !ast.IsNil(name) {
		c.add(kind, name.String(), detail)
	}
}

func (c *completions) add(kind protocol.CompletionItemKind, name, detail string) {
	item := protocol.CompletionItem{
		Kind:             kind,
		Label:            name,
		Detail:           detail,
		CommitCharacters: commitCharacters,
	}

	switch kind {
	case protocol.CompletionItemKindFunction, protocol.CompletionItemKindMethod:
		item.InsertText = name + "($1)"
		item.InsertTextFormat = protocol.InsertTextFormatSnippet

	case protocol.CompletionItemKindSnippet:
		switch name {
		case "if", "while":
			item.InsertText = name + " ($1)"
			item.InsertTextFormat = protocol.InsertTextFormatSnippet

		case "for":
			item.InsertText = "for ($1; $2; $3)"
			item.InsertTextFormat = protocol.InsertTextFormatSnippet
		}
	}

	c.items = append(c.items, item)
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
