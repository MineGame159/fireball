package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *codegen) VisitFunc(decl *ast.Func) {
	c.pushScope()
	c.locals.reset()

	c.writeFmt("define %s @%s(", c.getType(decl.Returns), decl.Name)

	for i, param := range decl.Params {
		if i > 0 {
			c.writeStr(", ")
		}

		c.writeFmt("%s %%%s", c.getType(param.Type), param.Name)
	}

	c.writeStr(") {\n")
	c.writeRaw("entry:\n")

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
	c.popScope()
}
