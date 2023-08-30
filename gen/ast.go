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
	w.write("import \"fireball/core/scanner\"")
	w.write("import \"fireball/core/types\"")

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
		// Struct
		w.write("type %s struct {", item.name)

		if kind == "Expr" {
			w.write("type_ types.Type")
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

			// Accept
			w.write("%s Accept(visitor %sVisitor) {", method, kind)
			w.write("visitor.Visit%s(%c)", item.name, short)
			w.write("}")
			w.write("")

			// AcceptChildren
			w.write("%s AcceptChildren(acceptor Acceptor) {", method)

			for _, f := range item.fields {
				type_ := f.type_
				array := false

				if strings.HasPrefix(type_, "[]") {
					type_ = type_[2:]
					array = true
				}

				if type_ == "Decl" || type_ == "Stmt" || type_ == "Expr" {
					if array {
						w.write("for _, v := range %c.%s {", short, f.name)
						w.write("acceptor.Accept%s(v)", type_)
						w.write("}")
					} else {
						w.write("if %c.%s != nil {", short, f.name)
						w.write("acceptor.Accept%s(%c.%s)", type_, short, f.name)
						w.write("}")
					}
				}
			}

			w.write("}")
			w.write("")

			// AcceptTypes
			w.write("%s AcceptTypes(visitor types.Visitor) {", method)

			for _, f := range item.fields {
				if strings.HasPrefix(f.type_, "[]") {
					type_ := f.type_[2:]

					if type_ == "Type" {
						w.write("for i := range %c.%s {", short, f.name)
						w.write("visitor.VisitType(&%c.%s[i])", short, f.name)
						w.write("}")
					} else {
						fi := getItem(items, type_)

						if fi != nil && fi.hasTypeField() {
							w.write("for i := range %c.%s {", short, f.name)

							for _, fif := range fi.fields {
								if fif.type_ == "Type" {
									w.write("visitor.VisitType(&%c.%s[i].%s)", short, f.name, fif.name)
								}
							}

							w.write("}")
						}
					}
				} else if f.type_ == "Type" {
					w.write("visitor.VisitType(&%c.%s)", short, f.name)
				} else {
					fi := getItem(items, f.type_)

					if fi != nil && fi.hasTypeField() {
						for _, fif := range fi.fields {
							if fif.type_ == "Type" {
								w.write("visitor.VisitType(&%c.%s.%s)", short, f.name, fif.name)
							}
						}
					}
				}
			}

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

func getItem(items []item, name string) *item {
	for i, _ := range items {
		if items[i].name == name {
			return &items[i]
		}
	}

	return nil
}

func (i *item) hasTypeField() bool {
	for _, f := range i.fields {
		if f.type_ == "Type" {
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
