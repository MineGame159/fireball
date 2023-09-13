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
	val := c.locals.unnamed(stmt.Type)
	val.identifier += "." + stmt.Name.Lexeme

	c.addVariable(stmt.Name, val)

	loc := c.debug.location(stmt.Name)
	c.writeFmt("%s = alloca %s, !dbg %s\n", val, c.getType(stmt.Type), loc)

	// Initializer
	if stmt.Initializer != nil {
		init := c.load(c.acceptExpr(stmt.Initializer), stmt.Initializer.Result().Type)
		c.writeFmt("store %s %s, ptr %s, !dbg %s\n", c.getType(stmt.Initializer.Result().Type), init, val, loc)
	}

	// Debug
	c.variableDebug(stmt.Name, val, stmt.Type, 0, loc)
}

func (c *codegen) variableDebug(name scanner.Token, ptr value, type_ types.Type, arg int, loc string) {
	dbg := c.debug.localVariable(name.Lexeme, c.getDbgType(type_), arg, name.Line)
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
	condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Result().Type)
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
	// Get basic block names
	prevLoopStart := c.loopStart
	prevLoopEnd := c.loopEnd

	c.loopStart = c.blocks.unnamedRaw()
	body := c.loopStart
	c.loopEnd = ""

	// Initializer
	c.pushScope()

	c.acceptStmt(stmt.Initializer)
	c.writeFmt("br label %%%s\n", c.loopStart)

	// Condition
	c.writeBlock(c.loopStart)

	if stmt.Condition != nil {
		body = c.blocks.unnamedRaw()
		c.loopEnd = c.blocks.unnamedRaw()

		condition := c.load(c.acceptExpr(stmt.Condition), stmt.Condition.Result().Type)
		loc := c.debug.location(stmt.Token())

		c.writeFmt("br i1 %s, label %%%s, label %%%s, !dbg %s\n", condition, body, c.loopEnd, loc)
	} else {
		c.loopEnd = c.blocks.unnamedRaw()
	}

	// Body and increment
	if c.loopStart != body {
		c.writeBlock(body)
	}

	c.acceptStmt(stmt.Body)
	c.acceptExpr(stmt.Increment)

	c.writeFmt("br label %%%s\n", c.loopStart)

	// End
	c.popScope()
	c.writeBlock(c.loopEnd)

	// Reset basic block names
	c.loopStart = prevLoopStart
	c.loopEnd = prevLoopEnd
}

func (c *codegen) VisitReturn(stmt *ast.Return) {
	loc := c.debug.location(stmt.Token())

	if stmt.Expr == nil {
		// Void
		c.writeFmt("ret void, !dbg %s\n", loc)
	} else {
		// Other
		val := c.load(c.acceptExpr(stmt.Expr), stmt.Expr.Result().Type)
		c.writeFmt("ret %s %s, !dbg %s\n", c.getType(stmt.Expr.Result().Type), val, loc)
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
