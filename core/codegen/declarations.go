package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *codegen) VisitStruct(decl *ast.Struct) {

}

func (c *codegen) VisitFunc(decl *ast.Func) {
	// Setup state
	c.pushScope()
	c.locals.reset()

	// Signature
	c.writeFmt("%s %s @%s(", ternary(decl.Extern, "declare", "define"), c.getType(decl.Returns), decl.Name)

	for i, param := range decl.Params {
		if i > 0 {
			c.writeStr(", ")
		}

		c.writeFmt("%s %%%s", c.getType(param.Type), param.Name)
	}

	if decl.Variadic {
		if len(decl.Params) > 0 {
			c.writeStr(", ")
		}

		c.writeStr("...")
	}

	// Body
	if decl.Extern {
		c.writeStr(")\n\n")
	} else {
		c.writeStr(") {\n")
		c.writeBlock(c.blocks.unnamedRaw())

		for _, param := range decl.Params {
			val := c.locals.named(param.Name.Lexeme+".var", param.Type)
			type_ := c.getType(param.Type)

			c.writeFmt("%s = alloca %s\n", val, type_)
			c.writeFmt("store %s %%%s, ptr %s\n", type_, param.Name, val)

			c.addVariable(param.Name, val)
		}

		for _, stmt := range decl.Body {
			c.acceptStmt(stmt)
		}

		if types.IsPrimitive(decl.Returns, types.Void) {
			c.writeStr("ret void\n")
		}

		c.writeStr("}\n\n")
	}

	// Restore state
	c.popScope()
}
