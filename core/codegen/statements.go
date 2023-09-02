package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (c *codegen) VisitBlock(stmt *ast.Block) {
	c.pushScope()
	c.debug.pushScope(c.debug.lexicalBlock(stmt.Token()))

	for _, s := range stmt.Stmts {
		c.acceptStmt(s)
	}

	c.debug.popScope()
	c.popScope()
}

func (c *codegen) VisitExpression(stmt *ast.Expression) {
	c.acceptExpr(stmt.Expr)
}

func (c *codegen) VisitVariable(stmt *ast.Variable) {
	// Variable
	val := c.locals.named(stmt.Name.Lexeme, stmt.Type)
	c.addVariable(stmt.Name, val)
	c.writeFmt("%s = alloca %s\n", val, c.getType(stmt.Type))

	// Initializer
	if stmt.Initializer != nil {
		init := c.load(c.acceptExpr(stmt.Initializer), stmt.Initializer.Type())
		c.writeFmt("store %s %s, ptr %s\n", c.getType(stmt.Initializer.Type()), init, val)
	}

	// Debug
	c.variableDebug(stmt.Name, val, stmt.Type, 0)
}

func (c *codegen) variableDebug(name scanner.Token, ptr value, type_ types.Type, arg int) {
	dbg := c.debug.localVariable(name.Lexeme, c.getDbgType(type_), arg, name.Line)
	loc := c.debug.location(name)
	c.writeFmt("call void @llvm.dbg.declare(metadata ptr %s, metadata %s, metadata !DIExpression()), !dbg %s\n", ptr, dbg, loc)
}

func (c *codegen) VisitIf(stmt *ast.If) {
	// GetLeaf basic block names
	then := c.blocks.unnamedRaw()
	else_ := ""
	end := ""

	if stmt.Else != nil {
		else_ = c.blocks.unnamedRaw()
		end = c.blocks.unnamedRaw()
	} else {
		else_ = c.blocks.unnamedRaw()
		end = else_
	}

	// Condition
	condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Type())
	loc := c.debug.location(stmt.Token())

	c.writeFmt("br i1 %s, label %%%s, label %%%s, !dbg %s\n", condition, then, else_, loc)

	// Then
	c.writeBlock(then)
	c.acceptStmt(stmt.Then)
	c.writeFmt("br label %%%s\n", end)

	// Else
	if stmt.Else != nil {
		c.writeBlock(else_)
		c.acceptStmt(stmt.Else)
		c.writeFmt("br label %%%s\n", end)
	}

	// End
	c.writeBlock(end)
}

func (c *codegen) VisitFor(stmt *ast.For) {
	// GetLeaf basic block names
	c.loopStart = c.blocks.unnamedRaw()
	body := c.loopStart
	c.loopEnd = ""

	c.writeFmt("br label %%%s\n", c.loopStart)

	// Condition
	c.writeBlock(c.loopStart)

	if stmt.Condition != nil {
		body = c.blocks.unnamedRaw()
		c.loopEnd = c.blocks.unnamedRaw()

		condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Type())
		loc := c.debug.location(stmt.Token())

		c.writeFmt("br i1 %s, label %%%s, label %%%s, !dbg %s\n", condition, body, c.loopEnd, loc)
	} else {
		c.loopEnd = c.blocks.unnamedRaw()
	}

	// Body
	if c.loopStart != body {
		c.writeBlock(body)
	}

	c.acceptStmt(stmt.Body)
	c.writeFmt("br label %%%s\n", c.loopStart)

	// End
	c.writeBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = ""
	c.loopEnd = ""
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	loc := c.debug.location(stmt.Token())

	if stmt.Expr == nil {
		// Void
		c.writeFmt("ret void, !dbg %s\n", loc)
	} else {
		// Other
		val := c.load(c.acceptExpr(stmt.Expr), stmt.Expr.Type())
		c.writeFmt("ret %s %s, !dbg %s\n", c.getType(stmt.Expr.Type()), val, loc)
	}
}

func (c *codegen) VisitBreak(stmt *ast.Break) {
	loc := c.debug.location(stmt.Token())
	c.writeFmt("br label %%%s, !dbg %s\n", c.loopEnd, loc)
}

func (c *codegen) VisitContinue(stmt *ast.Continue) {
	loc := c.debug.location(stmt.Token())
	c.writeFmt("br label %%%s, !dbg %s\n", c.loopStart, loc)
}
