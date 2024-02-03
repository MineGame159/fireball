package main

import "strings"

type Group struct {
	file string
	name string

	nodes              []Node
	additionalVisitors []string
}

type Node struct {
	name       string
	fields     []Field
	allowEmpty bool
}

func node(name string, fields ...Field) Node {
	return Node{name: name, fields: fields}
}

func nodeAllowEmpty(name string, fields ...Field) Node {
	return Node{name: name, fields: fields, allowEmpty: true}
}

func (n *Node) tokenField() *Field {
	for i := 0; i < len(n.fields); i++ {
		field := &n.fields[i]

		if field.type_.name == "scanner.Token" {
			return field
		}
	}

	return nil
}

type Field struct {
	name   string
	public bool

	type_ Type
}

func field(name string, type_ Type) Field {
	return Field{name: strings.ToUpper(name[0:1]) + name[1:], public: name[0] >= 'A' && name[0] <= 'Z', type_: type_}
}

type Type struct {
	name  string
	array bool
}

func type_(name string) Type {
	return Type{name: name}
}

func array(name string) Type {
	return Type{name: name, array: true}
}

func (t Type) concreteNode() bool {
	for _, group := range groups {
		for _, node := range group.nodes {
			if node.name == t.name {
				return true
			}
		}
	}

	return false
}

func (t Type) node() bool {
	for _, group := range groups {
		if group.name == t.name {
			return true
		}
	}

	return t.concreteNode()
}

func (t Type) String() string {
	pointer := t.concreteNode()

	if t.array {
		if pointer {
			return "[]*" + t.name
		}
		return "[]" + t.name
	}

	if pointer {
		return "*" + t.name
	}
	return t.name
}
