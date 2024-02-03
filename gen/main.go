package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, group := range groups {
		if group.file == "" {
			continue
		}

		w := newWriter()

		w.write("package ast")
		w.write("")

		w.write("import (")
		w.depth++
		w.write("\"fireball/core/cst\"")
		w.write("\"fireball/core/scanner\"")
		w.depth--
		w.write(")")
		w.write("")

		genGroup(w, group)

		w.flush("../core/ast/" + group.file)
	}
}

func genGroup(w *Writer, group Group) {
	if group.name != "" {
		genVisitor(w, group)
	}

	for _, node := range group.nodes {
		genNode(w, node, group.name)
	}

	if group.name == "" {
		genIsNil(w)
	}
}

func genVisitor(w *Writer, group Group) {
	w.write("// Visitor")
	w.write("")

	// Visitor

	w.write("type %sVisitor interface {", group.name)

	for _, node := range group.nodes {
		w.write("Visit%s(%s *%s)", node.name, name(strings.ToLower(group.name)), node.name)
	}

	for _, visitor := range group.additionalVisitors {
		w.write("Visit%s(%s *%s)", visitor, name(strings.ToLower(group.name)), visitor)
	}

	w.write("}")
	w.write("")

	// Interface

	w.write("type %s interface {", group.name)
	w.write("Node")
	w.write("")

	if group.name == "Type" {
		w.write("Equals(other Type) bool")
		w.write("")
		w.write("Resolved() Type")
		w.write("")
	} else if group.name == "Expr" {
		w.write("Result() *ExprResult")
		w.write("")
	}

	w.write("Accept%s(visitor %sVisitor)", group.name, group.name)
	w.write("}")
	w.write("")
}

func genNode(w *Writer, node Node, visitor string) {
	w.write("// %s", node.name)
	w.write("")

	// Struct

	w.write("type %s struct {", node.name)

	w.write("cst cst.Node")
	w.write("parent Node")
	w.write("")

	for _, field := range node.fields {
		w.write("%s %s", name(field.name), field.type_)
	}

	if visitor == "Expr" {
		w.write("")
		w.write("result ExprResult")
	}

	w.write("}")
	w.write("")

	this := strings.ToLower(node.name)[0:1]
	method := fmt.Sprintf("func (%s *%s)", this, node.name)

	// Constructor

	genConstructor(w, node, this)

	// Cst()

	w.write("%s Cst() *cst.Node {", method)
	w.write("if %s.cst.Kind == cst.UnknownNode {", this)
	w.write("return nil")
	w.write("}")
	w.write("")
	w.write("return &%s.cst", this)
	w.write("}")
	w.write("")

	// Token()

	w.write("%s Token() scanner.Token {", method)
	tokenField := node.tokenField()

	if tokenField == nil {
		w.write("return scanner.Token{}")
	} else {
		w.write("return %s.%s", this, name(tokenField.name))
	}

	w.write("}")
	w.write("")

	// Parent()

	w.write("%s Parent() Node {", method)
	w.write("return %s.parent", this)
	w.write("}")
	w.write("")

	// SetParent()

	w.write("%s SetParent(parent Node) {", method)
	w.write("if parent != nil && %s.parent != nil {", this)
	w.write("panic(\"ast.%s.SetParent() - Parent is already set\")", node.name)
	w.write("}")
	w.write("")
	w.write("%s.parent = parent", this)
	w.write("}")
	w.write("")

	// AcceptChildren()

	w.write("%s AcceptChildren(visitor Visitor) {", method)

	for _, field := range node.fields {
		if !field.type_.node() || field.public {
			continue
		}

		if field.type_.array {
			w.write("for _, child := range %s.%s {", this, name(field.name))
			w.write("visitor.VisitNode(child)")
			w.write("}")
		} else {
			w.write("if %s.%s != nil {", this, name(field.name))
			w.write("visitor.VisitNode(%s.%s)", this, name(field.name))
			w.write("}")
		}
	}

	w.write("}")
	w.write("")

	// Clone()

	w.write("%s Clone() Node {", method)

	w.write("%s2 := &%s{", this, node.name)
	w.write("cst: %s.cst,", this)

	for _, field := range node.fields {
		if !field.type_.node() || field.public {
			w.write("%s: %s.%s,", name(field.name), this, name(field.name))
		}
	}

	w.write("}")
	w.write("")

	setParents := false

	for _, field := range node.fields {
		if !field.type_.node() {
			continue
		}

		fieldName := name(field.name)

		castTo := field.type_.name
		if field.type_.concreteNode() {
			castTo = "*" + castTo
		}

		if field.type_.array {
			w.write("%s2.%s = make(%s, len(%s.%s))", this, fieldName, field.type_, this, fieldName)
			w.write("for i, child := range %s2.%s {", this, fieldName)
			w.write("%s2.%s[i] = child.Clone().(%s)", this, fieldName, castTo)
			w.write("%s2.%s[i].SetParent(%s2)", this, fieldName, this)
			w.write("}")
		} else {
			w.write("if %s.%s != nil {", this, fieldName)
			w.write("%s2.%s = %s.%s.Clone().(%s)", this, fieldName, this, fieldName, castTo)
			w.write("%s2.%s.SetParent(%s2)", this, fieldName, this)
			w.write("}")
		}

		setParents = true
	}

	if setParents {
		w.write("")
	}

	w.write("return %s2", this)
	w.write("}")
	w.write("")

	// String()

	w.write("%s String() string {", method)

	if tokenField == nil {
		w.write("return \"\"")
	} else {
		w.write("return %s.%s.String()", this, name(tokenField.name))
	}

	w.write("}")
	w.write("")

	// Accept()

	if visitor != "" {
		w.write("%s Accept%s(visitor %sVisitor) {", method, visitor, visitor)
		w.write("visitor.Visit%s(%s)", node.name, this)
		w.write("}")
		w.write("")
	}

	// Resolved()

	if visitor == "Type" {
		w.write("%s Resolved() Type {", method)

		var resolved *Field

		for i := 0; i < len(node.fields); i++ {
			field := &node.fields[i]

			if field.name == "Type" {
				resolved = field
			}
		}

		if resolved == nil {
			w.write("return %s", this)
		} else {
			w.write("return %s.%s", this, name(resolved.name))
		}

		w.write("}")
		w.write("")
	}

	// Result()

	if visitor == "Expr" {
		w.write("%s Result() *ExprResult {", method)
		w.write("return &%s.result", this)
		w.write("}")
		w.write("")
	}
}

func genConstructor(w *Writer, node Node, this string) {
	sb := strings.Builder{}

	// Parameters

	sb.WriteString(fmt.Sprintf("func New%s(node cst.Node", node.name))

	for _, field := range node.fields {
		if !field.public {
			sb.WriteString(fmt.Sprintf(", %s %s", name(strings.ToLower(field.name)), field.type_))
		}
	}

	sb.WriteString(fmt.Sprintf(") *%s {", node.name))
	w.write(sb.String())

	// Check if the node will be empty and return nil

	if !node.allowEmpty {
		sb.Reset()

		i := 0

		for _, field := range node.fields {
			if !field.public && (field.type_.node() || field.type_.name == "scanner.Token" || strings.HasSuffix(field.type_.name, "Kind") || strings.HasSuffix(field.type_.name, "Flags")) {
				i++
			}
		}

		if i > 0 {
			sb.WriteString("if ")
			i := 0

			for _, field := range node.fields {
				if field.public {
					continue
				}

				if field.type_.node() {
					if i > 0 {
						sb.WriteString(" && ")
					}

					sb.WriteString(fmt.Sprintf("%s == nil", name(strings.ToLower(field.name))))
					i++
				} else if field.type_.name == "scanner.Token" {
					if i > 0 {
						sb.WriteString(" && ")
					}

					sb.WriteString(fmt.Sprintf("%s.IsEmpty()", name(strings.ToLower(field.name))))
					i++
				} else if strings.HasSuffix(field.type_.name, "Kind") || strings.HasSuffix(field.type_.name, "Flags") {
					if i > 0 {
						sb.WriteString(" && ")
					}

					sb.WriteString(fmt.Sprintf("%s == 0", name(strings.ToLower(field.name))))
					i++
				}
			}

			sb.WriteString(" {")

			w.write(sb.String())
			w.write("return nil")
			w.write("}")
			w.write("")
		}
	}

	// Create node

	w.write("%s := &%s{", this, node.name)

	w.write("cst: node,")

	for _, field := range node.fields {
		if !field.public {
			w.write("%s: %s,", name(field.name), name(strings.ToLower(field.name)))
		}
	}

	w.write("}")
	w.write("")

	// Set parents

	setParents := false

	for _, field := range node.fields {
		if !field.type_.node() || field.public {
			continue
		}

		param := name(strings.ToLower(field.name))

		if field.type_.array {
			w.write("for _, child := range %s {", param)
			w.write("child.SetParent(%s)", this)
			w.write("}")
		} else {
			w.write("if %s != nil {", param)
			w.write("%s.SetParent(%s)", param, this)
			w.write("}")
		}

		setParents = true
	}

	if setParents {
		w.write("")
	}

	// Return

	w.write("return %s", this)
	w.write("}")
	w.write("")
}

func genIsNil(w *Writer) {
	w.write("func IsNil(node Node) bool {")
	w.write("if node == nil {")
	w.write("return true")
	w.write("}")
	w.write("")
	w.write("switch node := node.(type) {")

	for _, group := range groups {
		for _, node := range group.nodes {
			w.depth--
			w.write("case *%s:", node.name)
			w.depth++

			w.write("return node == nil")
		}
	}

	w.write("")

	w.depth--
	w.write("default:")
	w.depth++
	w.write("panic(\"ast.IsNil() - Not implemented\")")

	w.write("}")
	w.write("}")
	w.write("")
}

func name(name string) string {
	switch name {
	case "struct", "type", "else", "func", "Token":
		return name + "_"

	default:
		return name
	}
}
