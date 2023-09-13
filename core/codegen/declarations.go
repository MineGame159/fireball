package codegen

import (
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *codegen) VisitStruct(_ *ast.Struct) {
}

func (c *codegen) VisitEnum(_ *ast.Enum) {
}

func (c *codegen) VisitFunc(decl *ast.Func) {
	// Setup state
	c.pushScope()
	c.locals.reset()

	// Debug
	var dbg string

	if !decl.Extern {
		dbgTypes := make([]string, len(decl.Params)+1)

		dbgTypes[0] = c.getDbgType(decl.Returns)

		for i, param := range decl.Params {
			dbgTypes[i+1] = c.getDbgType(param.Type)
		}

		type_ := c.debug.subroutineType(c.debug.tuple(dbgTypes))
		dbg = c.debug.subprogram(decl.Name.Lexeme, type_, decl.Name.Line)

		c.debug.pushScope(dbg)
	}

	// Signature
	name := decl.Name.Lexeme
	if !decl.Extern {
		name = mangleName(name)
	}

	c.writeFmt("%s %s @%s(", ternary(decl.Extern, "declare", "define"), c.getType(decl.Returns), name)

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
		c.writeFmt(") !dbg %s {\n", dbg)
		c.writeBlock(c.blocks.unnamedRaw())

		for i, param := range decl.Params {
			val := c.locals.named(param.Name.Lexeme + ".var")
			type_ := c.getType(param.Type)

			loc := c.debug.location(param.Name)
			c.writeFmt("%s = alloca %s, !dbg %s\n", val, type_, loc)
			c.writeFmt("store %s %%%s, ptr %s, !dbg %s\n", type_, param.Name, val, loc)

			c.addVariable(param.Name, val)
			c.variableDebug(param.Name, val, param.Type, i+1, loc)
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
	if !decl.Extern {
		c.debug.popScope()
	}

	c.popScope()
}
