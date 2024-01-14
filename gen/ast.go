package main

var groups = []Group{
	types,
	declarations,
	statements,
	expressions,
	other,
}

// Types

var types = Group{
	file: "types.go",
	name: "Type",

	nodes: []Node{
		node(
			"Primitive",
			field("kind", type_("PrimitiveKind")),
			field("token", type_("scanner.Token")),
		),
		node(
			"Pointer",
			field("pointee", type_("Type")),
		),
		node(
			"Array",
			field("base", type_("Type")),
			field("count", type_("uint32")),
		),
		node(
			"Resolvable",
			field("parts", array("Token")),
			field("Type", type_("Type")),
		),
	},

	additionalVisitors: []string{
		"Struct",
		"Enum",
		"Func",
	},
}

// Declarations

var declarations = Group{
	file: "declarations.go",
	name: "Decl",

	nodes: []Node{
		node(
			"Namespace",
			field("name", type_("NamespaceName")),
		),
		node(
			"Using",
			field("name", type_("NamespaceName")),
		),
		node(
			"Struct",
			field("name", type_("Token")),
			field("fields", array("Field")),
			field("staticFields", array("Field")),
		),
		node(
			"Enum",
			field("name", type_("Token")),
			field("type", type_("Type")),
			field("ActualType", type_("Type")),
			field("cases", array("EnumCase")),
		),
		node(
			"Impl",
			field("struct", type_("Token")),
			field("Type", type_("Type")),
			field("methods", array("Func")),
		),
		node(
			"Func",
			field("attributes", array("Attribute")),
			field("flags", type_("FuncFlags")),
			field("name", type_("Token")),
			field("params", array("Param")),
			field("returns", type_("Type")),
			field("body", array("Stmt")),
		),
		node(
			"GlobalVar",
			field("name", type_("Token")),
			field("type", type_("Type")),
		),
	},
}

// Statements

var statements = Group{
	file: "statements.go",
	name: "Stmt",

	nodes: []Node{
		node(
			"Expression",
			field("expr", type_("Expr")),
		),
		node(
			"Block",
			field("stmts", array("Stmt")),
		),
		node(
			"Var",
			field("name", type_("Token")),
			field("type", type_("Type")),
			field("ActualType", type_("Type")),
			field("value", type_("Expr")),
		),
		node(
			"If",
			field("condition", type_("Expr")),
			field("then", type_("Stmt")),
			field("else", type_("Stmt")),
		),
		node(
			"While",
			field("condition", type_("Expr")),
			field("body", type_("Stmt")),
		),
		node(
			"For",
			field("initializer", type_("Stmt")),
			field("condition", type_("Expr")),
			field("increment", type_("Expr")),
			field("body", type_("Stmt")),
		),
		node(
			"Return",
			field("value", type_("Expr")),
		),
		node("Break"),
		node("Continue"),
	},
}

// Expressions

var expressions = Group{
	file: "expressions.go",
	name: "Expr",

	nodes: []Node{
		node(
			"Paren",
			field("expr", type_("Expr")),
		),
		node(
			"Unary",
			field("prefix", type_("bool")),
			field("operator", type_("Token")),
			field("value", type_("Expr")),
		),
		node(
			"Binary",
			field("left", type_("Expr")),
			field("operator", type_("Token")),
			field("right", type_("Expr")),
		),
		node(
			"Logical",
			field("left", type_("Expr")),
			field("operator", type_("Token")),
			field("right", type_("Expr")),
		),
		node(
			"Assignment",
			field("assignee", type_("Expr")),
			field("operator", type_("Token")),
			field("value", type_("Expr")),
		),
		node(
			"Member",
			field("value", type_("Expr")),
			field("name", type_("Token")),
		),
		node(
			"Index",
			field("value", type_("Expr")),
			field("index", type_("Expr")),
		),
		node(
			"Cast",
			field("value", type_("Expr")),
			field("target", type_("Type")),
		),
		node(
			"Call",
			field("callee", type_("Expr")),
			field("args", array("Expr")),
		),
		node(
			"TypeCall",
			field("callee", type_("Token")),
			field("arg", type_("Type")),
		),
		node(
			"StructInitializer",
			field("new", type_("bool")),
			field("type", type_("Type")),
			field("fields", array("InitField")),
		),
		node(
			"ArrayInitializer",
			field("values", array("Expr")),
		),
		node(
			"AllocateArray",
			field("type", type_("Type")),
			field("count", type_("Expr")),
		),
		node(
			"Identifier",
			field("name", type_("scanner.Token")),
		),
		node(
			"Literal",
			field("token", type_("scanner.Token")),
		),
	},
}

// Other

var other = Group{
	file: "other.go",
	name: "",

	nodes: []Node{
		node(
			"File",
			field("path", type_("string")),
			field("namespace", type_("Namespace")),
			field("decls", array("Decl")),
		),
		node(
			"NamespaceName",
			field("parts", array("Token")),
		),
		node(
			"Field",
			field("name", type_("Token")),
			field("type", type_("Type")),
		),
		node(
			"InitField",
			field("name", type_("Token")),
			field("value", type_("Expr")),
		),
		node(
			"EnumCase",
			field("name", type_("Token")),
			field("value", type_("Token")),
			field("ActualValue", type_("int64")),
		),
		node(
			"Param",
			field("name", type_("Token")),
			field("type", type_("Type")),
		),
		node(
			"Attribute",
			field("name", type_("Token")),
			field("args", array("Token")),
		),
		node(
			"Token",
			field("token", type_("scanner.Token")),
		),
	},
}
