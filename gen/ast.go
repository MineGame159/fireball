package main

var groups = []Group{
	types,
	declarations,
	statements,
	expressions,
	other,
}

var externalNodes = []string{
	"SpecializedField",
	"SpecializedStruct",
	"SpecializedFunc",
	"PartiallySpecializedFunc",
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
			field("genericArgs", array("Type")),
			field("Type", type_("Type")),
		),
		nodeSkipResolved(
			"Generic",
			field("name", type_("scanner.Token")),
			field("Type", type_("Type")),
		),
	},

	additionalVisitors: []string{
		"Struct",
		"Enum",
		"Interface",
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
			field("attributes", array("Attribute")),
			field("name", type_("Token")),
			field("genericParams", array("Generic")),
			field("fields", array("Field")),
			field("staticFields", array("Field")),

			field("Specializations", array("*SpecializedStruct")),
			field("Type", type_("Type")),
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
			field("implements", type_("Type")),
			field("methods", array("Func")),
		),
		node(
			"Interface",
			field("name", type_("Token")),
			field("methods", array("Func")),
		),
		node(
			"Func",
			field("attributes", array("Attribute")),
			field("flags", type_("FuncFlags")),
			field("name", type_("Token")),
			field("genericParams", array("Generic")),
			field("params", array("Param")),
			field("returns_", type_("Type")),
			field("body", array("Stmt")),

			field("Specializations", array("*SpecializedFunc")),
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
		nodeAllowEmpty(
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
			"Literal",
			field("token", type_("scanner.Token")),
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
			field("genericArgs", array("Type")),
		),
		node(
			"Index",
			field("value", type_("Expr")),
			field("index", type_("Expr")),
		),
		node(
			"Cast",
			field("value", type_("Expr")),
			field("operator", type_("Token")),
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
			"Typeof",
			field("callee", type_("Token")),
			field("arg", type_("Expr")),
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
			field("name", type_("Token")),
			field("genericArgs", array("Type")),
		),
	},
}

// Other

var other = Group{
	file: "other.go",
	name: "",

	nodes: []Node{
		nodeAllowEmpty(
			"File",
			field("path", type_("string")),
			field("namespace", type_("Namespace")),
			field("Resolver", type_("Resolver")),
			field("decls", array("Decl")),
		),
		node(
			"NamespaceName",
			field("parts", array("Token")),
		),
		node(
			"Field",
			field("name_", type_("Token")),
			field("type_", type_("Type")),
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
