package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

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

	// Body
	if decl.Extern {
		c.writeStr(")\n\n")
	} else {
		c.writeStr(") {\n")
		c.writeRaw(c.blocks.unnamedRaw() + ":\n")

		for _, param := range decl.Params {
			c.addVariable(param.Name, c.locals.named(param.Name.Lexeme, param.Type))
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
