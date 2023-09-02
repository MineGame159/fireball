package main

import (
	"fmt"
	"os"
	"strings"
)

type item struct {
	name   string
	fields []field
	token  string
	ast    bool
}

type field struct {
	name  string
	type_ string
}

var decls = []item{
	{
		name: "Struct",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Fields", type_: "[]Field"},
			{name: "Type", type_: "Type"},
		},
		token: "Name",
		ast:   true,
	},
	{
		name: "Field",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Type", type_: "Type"},
		},
		ast: false,
	},
	{
		name: "Enum",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Type", type_: "Type"},
			{name: "InferType", type_: "bool"},
			{name: "Cases", type_: "[]EnumCase"},
		},
		token: "Name",
		ast:   true,
	},
	{
		name: "EnumCase",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Value", type_: "int"},
			{name: "InferValue", type_: "bool"},
		},
		ast: false,
	},
	{
		name: "Func",
		fields: []field{
			{name: "Extern", type_: "bool"},
			{name: "Name", type_: "Token"},
			{name: "Params", type_: "[]Param"},
			{name: "Variadic", type_: "bool"},
			{name: "Returns", type_: "Type"},
			{name: "Body", type_: "[]Stmt"},
		},
		token: "Name",
		ast:   true,
	},
	{
		name: "Param",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Type", type_: "Type"},
		},
		ast: false,
	},
}

var stmts = []item{
	{
		name: "Block",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Stmts", type_: "[]Stmt"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Expression",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Expr", type_: "Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Variable",
		fields: []field{
			{name: "Type", type_: "Type"},
			{name: "Name", type_: "Token"},
			{name: "Initializer", type_: "Expr"},
			{name: "InferType", type_: "bool"},
		},
		token: "Name",
		ast:   true,
	},
	{
		name: "If",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Condition", type_: "Expr"},
			{name: "Then", type_: "Stmt"},
			{name: "Else", type_: "Stmt"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "For",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Condition", type_: "Expr"},
			{name: "Body", type_: "Stmt"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Return",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Expr", type_: "Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Break",
		fields: []field{
			{name: "Token_", type_: "Token"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Continue",
		fields: []field{
			{name: "Token_", type_: "Token"},
		},
		token: "Token_",
		ast:   true,
	},
}

var exprs = []item{
	{
		name: "Group",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Expr", type_: "Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Literal",
		fields: []field{
			{name: "Value", type_: "Token"},
		},
		token: "Value",
		ast:   true,
	},
	{
		name: "Initializer",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Fields", type_: "[]InitField"},
		},
		token: "Name",
		ast:   true,
	},
	{
		name: "InitField",
		fields: []field{
			{name: "Name", type_: "Token"},
			{name: "Value", type_: "Expr"},
		},
		ast: false,
	},
	{
		name: "Unary",
		fields: []field{
			{name: "Op", type_: "Token"},
			{name: "Right", type_: "Expr"},
		},
		token: "Op",
		ast:   true,
	},
	{
		name: "Binary",
		fields: []field{
			{name: "Left", type_: "Expr"},
			{name: "Op", type_: "Token"},
			{name: "Right", type_: "Expr"},
		},
		token: "Op",
		ast:   true,
	},
	{
		name: "Logical",
		fields: []field{
			{name: "Left", type_: "Expr"},
			{name: "Op", type_: "Token"},
			{name: "Right", type_: "Expr"},
		},
		token: "Op",
		ast:   true,
	},
	{
		name: "Identifier",
		fields: []field{
			{name: "Identifier", type_: "Token"},
		},
		token: "Identifier",
		ast:   true,
	},
	{
		name: "Assignment",
		fields: []field{
			{name: "Assignee", type_: "Expr"},
			{name: "Op", type_: "Token"},
			{name: "Value", type_: "Expr"},
		},
		token: "Op",
		ast:   true,
	},
	{
		name: "Cast",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Expr", type_: "Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Call",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Callee", type_: "Expr"},
			{name: "Args", type_: "[]Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Index",
		fields: []field{
			{name: "Token_", type_: "Token"},
			{name: "Value", type_: "Expr"},
			{name: "Index", type_: "Expr"},
		},
		token: "Token_",
		ast:   true,
	},
	{
		name: "Member",
		fields: []field{
			{name: "Value", type_: "Expr"},
			{name: "Name", type_: "Token"},
		},
		token: "Name",
		ast:   true,
	},
}

func main() {
	file := os.Getenv("GOFILE")
	w := newWriter()

	switch file {
	case "declarations.go":
		generate(w, "Decl", decls)
	case "statements.go":
		generate(w, "Stmt", stmts)
	case "expressions.go":
		generate(w, "Expr", exprs)
	}

	w.flush(file)
}

func generate(w *writer, kind string, items []item) {
	w.write("package ast")
	w.write("")
	w.write("import \"fireball/core\"")
	w.write("import \"fireball/core/types\"")
	w.write("import \"fireball/core/scanner\"")

	w.write("")
	w.write("//go:generate go run ../../gen/ast.go")
	w.write("")

	// Visitor
	w.write("type %sVisitor interface {", kind)

	for _, item := range items {
		if item.ast {
			w.write("Visit%s(%s *%s)", item.name, strings.ToLower(kind), item.name)
		}
	}

	w.write("}")
	w.write("")

	// Base
	w.write("type %s interface {", kind)

	w.write("Node")
	w.write("")
	w.write("Accept(visitor %sVisitor)", kind)

	if kind == "Expr" {
		w.write("")
		w.write("Type() types.Type")
		w.write("SetType(type_ types.Type)")
	}

	w.write("}")
	w.write("")

	// Items
	for _, item := range items {
		// Comment
		w.write("// %s", item.name)
		w.write("")

		// Struct
		w.write("type %s struct {", item.name)

		if item.ast {
			w.write("range_ core.Range")

			if kind == "Expr" {
				w.write("type_ types.Type")
			}

			w.write("")
		}

		for _, field := range item.fields {
			type_ := field.type_

			if type_ == "Token" {
				type_ = "scanner.Token"
			} else if type_ == "Type" {
				type_ = "types.Type"
			}

			w.write("%s %s", field.name, type_)
		}

		w.write("}")
		w.write("")

		// Node
		if item.ast {
			short := strings.ToLower(item.name)[0]
			method := fmt.Sprintf("func (%c *%s)", short, item.name)

			// Token
			w.write("%s Token() scanner.Token {", method)
			w.write("return %c.%s", short, item.token)
			w.write("}")
			w.write("")

			// Range
			w.write("%s Range() core.Range {", method)
			w.write("return %c.range_", short)
			w.write("}")
			w.write("")

			// SetRangeToken
			w.write("%s SetRangeToken(start, end scanner.Token) {", method)
			w.write("%c.range_ = core.Range{", short)
			w.write("Start: core.TokenToPos(start, false),")
			w.write("End: core.TokenToPos(end, true),")
			w.write("}")
			w.write("}")
			w.write("")

			// SetRangePos
			w.write("%s SetRangePos(start, end core.Pos) {", method)
			w.write("%c.range_ = core.Range{", short)
			w.write("Start: start,")
			w.write("End: end,")
			w.write("}")
			w.write("}")
			w.write("")

			// SetRangeNode
			w.write("%s SetRangeNode(start, end Node) {", method)
			w.write("%c.range_ = core.Range{", short)
			w.write("Start: start.Range().Start,")
			w.write("End: end.Range().End,")
			w.write("}")
			w.write("}")
			w.write("")

			// Accept
			w.write("%s Accept(visitor %sVisitor) {", method, kind)
			w.write("visitor.Visit%s(%c)", item.name, short)
			w.write("}")
			w.write("")

			// AcceptChildren
			leaf := !genVisitor(w, kind, items, item, short, method, "AcceptChildren", false, "Acceptor", "Accept?", func(s string) bool {
				return s == "Decl" || s == "Stmt" || s == "Expr"
			})

			// AcceptTypes
			genVisitor(w, kind, items, item, short, method, "AcceptTypes", false, "types.Visitor", "VisitType", func(s string) bool {
				return s == "Type"
			})

			// AcceptTypesPtr
			genVisitor(w, kind, items, item, short, method, "AcceptTypesPtr", true, "types.PtrVisitor", "VisitType", func(s string) bool {
				return s == "Type"
			})

			// Leaf
			w.write("%s Leaf() bool {", method)
			w.write("return %t", leaf)
			w.write("}")
			w.write("")

			// Expr
			if kind == "Expr" {
				// Type
				w.write("%s Type() types.Type {", method)
				w.write("return %c.type_", short)
				w.write("}")
				w.write("")

				// SetType
				w.write("%s SetType(type_ types.Type) {", method)
				w.write("%c.type_ = type_", short)
				w.write("}")
				w.write("")
			}
		}
	}
}

func genVisitor(w *writer, kind string, items []item, item item, short uint8, method string, name string, ptr bool, visitor string, visitFormat string, target func(string) bool) bool {
	w.write("%s %s(visitor %s) {", method, name, visitor)

	hasChildren := false
	ptrStr := ""

	if ptr {
		ptrStr = "&"
	}

	if kind == "Expr" && target("Type") {
		visit(w, ptr, "VisitType", fmt.Sprintf("%s%c.type_", ptrStr, short))
	}

	for _, f := range item.fields {
		if strings.HasPrefix(f.type_, "[]") {
			type_ := f.type_[2:]

			if target(type_) {
				w.write("for i_ := range %c.%s {", short, f.name)
				visit(w, ptr, formatVisit(visitFormat, type_), fmt.Sprintf("%s%c.%s[i_]", ptrStr, short, f.name))
				w.write("}")

				hasChildren = true
			} else {
				fi := getItem(items, type_)

				if fi != nil && fi.hasFieldWithType(target) {
					w.write("for i_ := range %c.%s {", short, f.name)

					for _, fif := range fi.fields {
						if target(fif.type_) {
							visit(w, ptr, formatVisit(visitFormat, fif.type_), fmt.Sprintf("%s%c.%s[i_].%s", ptrStr, short, f.name, fif.name))
						}
					}

					w.write("}")

					hasChildren = true
				}
			}
		} else if target(f.type_) {
			visit(w, ptr, formatVisit(visitFormat, f.type_), fmt.Sprintf("%s%c.%s", ptrStr, short, f.name))
			hasChildren = true
		} else {
			fi := getItem(items, f.type_)

			if fi != nil && fi.hasFieldWithType(target) {
				for _, fif := range fi.fields {
					if target(fif.type_) {
						visit(w, ptr, formatVisit(visitFormat, fif.type_), fmt.Sprintf("%s%c.%s.%s", ptrStr, short, f.name, fif.name))
					}
				}

				hasChildren = true
			}
		}
	}

	w.write("}")
	w.write("")

	return hasChildren
}

func visit(w *writer, ptr bool, visit string, arg string) {
	if !ptr {
		w.write("if %s != nil {", arg)
	}

	w.write("visitor.%s(%s)", visit, arg)

	if !ptr {
		w.write("}")
	}
}

func formatVisit(format, kind string) string {
	if strings.HasSuffix(format, "?") {
		format = format[:len(format)-1] + kind
	}

	return format
}

func getItem(items []item, name string) *item {
	for i := range items {
		if items[i].name == name {
			return &items[i]
		}
	}

	return nil
}

func (i *item) hasFieldWithType(target func(string) bool) bool {
	for _, f := range i.fields {
		if target(f.type_) {
			return true
		}
	}

	return false
}

type writer struct {
	str   strings.Builder
	depth int
}

func newWriter() *writer {
	return &writer{
		str:   strings.Builder{},
		depth: 0,
	}
}

func (w *writer) flush(file string) {
	_ = os.WriteFile(file, []byte(w.str.String()), 0666)
}

func (w *writer) write(format string, args ...any) {
	str := fmt.Sprintf(format, args...)

	if strings.HasPrefix(str, "}") {
		w.depth--
	}

	for i := 0; i < w.depth; i++ {
		w.str.WriteRune('\t')
	}

	w.str.WriteString(str)
	w.str.WriteRune('\n')

	if strings.HasSuffix(str, "{") {
		w.depth++
	}
}
