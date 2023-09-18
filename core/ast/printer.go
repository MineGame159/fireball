package ast

import (
	"fmt"
	"io"
	"strings"
)

type printer struct {
	writer io.Writer
	depth  int
}

func Print(node Node, writer io.Writer) {
	p := &printer{
		writer: writer,
		depth:  0,
	}

	if decl, ok := node.(Decl); ok {
		decl.Accept(p)
	} else if stmt, ok := node.(Stmt); ok {
		stmt.Accept(p)
	} else if expr, ok := node.(Expr); ok {
		expr.Accept(p)
	}
}

// Declarations

func (p *printer) VisitStruct(decl *Struct) {
	p.print("struct %s", decl.Name)

	for _, field := range decl.Fields {
		p.print("%s %s", field.Name, field.Type)
	}
}

func (p *printer) VisitImpl(decl *Impl) {
	p.print("impl %s", decl.Struct)

	for _, function := range decl.Functions {
		p.AcceptDecl(function)
	}
}

func (p *printer) VisitEnum(decl *Enum) {
	p.print("enum %s %s", decl.Name, decl.Type)
	p.depth++

	for _, case_ := range decl.Cases {
		p.print("%s = %d", case_.Name, case_.Value)
	}

	p.depth--
}

func (p *printer) VisitFunc(decl *Func) {
	str := strings.Builder{}

	str.WriteString(decl.Name.Lexeme)
	str.WriteRune('(')

	for i, param := range decl.Params {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(param.Type.String())
		str.WriteRune(' ')
		str.WriteString(param.Name.Lexeme)
	}

	str.WriteString(") ")

	if decl.Returns != nil {
		str.WriteString(decl.Returns.String())
	}

	p.print(str.String())

	for _, stmt := range decl.Body {
		p.AcceptStmt(stmt)
	}
}

// Statements

func (p *printer) VisitBlock(stmt *Block) {
	p.print("{}")

	for _, s := range stmt.Stmts {
		p.AcceptStmt(s)
	}
}

func (p *printer) VisitExpression(stmt *Expression) {
	p.print("expr")
	p.AcceptExpr(stmt.Expr)
}

func (p *printer) VisitVariable(stmt *Variable) {
	p.print("%s %s", stmt.Type.String(), stmt.Name.Lexeme)
	p.AcceptExpr(stmt.Initializer)
}

func (p *printer) VisitIf(stmt *If) {
	p.print("if")
	p.AcceptExpr(stmt.Condition)

	p.AcceptStmt(stmt.Then)
	p.AcceptStmt(stmt.Else)
}

func (p *printer) VisitFor(stmt *For) {
	p.print("for")

	p.AcceptExpr(stmt.Condition)
	p.AcceptStmt(stmt.Body)
}

func (p *printer) VisitReturn(stmt *Return) {
	p.print("return")
	p.AcceptExpr(stmt.Expr)
}

func (p *printer) VisitBreak(stmt *Break) {
	p.print("break")
}

func (p *printer) VisitContinue(stmt *Continue) {
	p.print("continue")
}

// Expressions

func (p *printer) VisitGroup(expr *Group) {
	p.print("()")
	p.AcceptExpr(expr.Expr)
}

func (p *printer) VisitLiteral(expr *Literal) {
	p.print(expr.Value.Lexeme)
}

func (p *printer) VisitStructInitializer(expr *StructInitializer) {
	p.print("%s {}", expr.Name)
	p.depth++

	for _, field := range expr.Fields {
		p.print("%s:", field.Name)
		p.AcceptExpr(field.Value)
	}

	p.depth--
}

func (p *printer) VisitArrayInitializer(expr *ArrayInitializer) {
	p.print("[]")
	p.depth++

	for _, value := range expr.Values {
		p.AcceptExpr(value)
	}

	p.depth--
}

func (p *printer) VisitUnary(expr *Unary) {
	p.print(expr.Op.Lexeme)
	p.AcceptExpr(expr.Value)
}

func (p *printer) VisitBinary(expr *Binary) {
	p.print(expr.Op.Lexeme)
	p.AcceptExpr(expr.Left)
	p.AcceptExpr(expr.Right)
}

func (p *printer) VisitLogical(expr *Logical) {
	p.print(expr.Op.Lexeme)
	p.AcceptExpr(expr.Left)
	p.AcceptExpr(expr.Right)
}

func (p *printer) VisitIdentifier(expr *Identifier) {
	p.print(expr.Identifier.Lexeme)
}

func (p *printer) VisitAssignment(expr *Assignment) {
	p.print(expr.Op.Lexeme)
	p.AcceptExpr(expr.Assignee)
	p.AcceptExpr(expr.Value)
}

func (p *printer) VisitCast(expr *Cast) {
	p.print("as %s", expr.Result().Type)
	p.AcceptExpr(expr.Expr)
}

func (p *printer) VisitSizeof(expr *Sizeof) {
	p.print("sizeof %s", expr.Target)
}

func (p *printer) VisitCall(expr *Call) {
	p.print("call")
	p.AcceptExpr(expr.Callee)

	for _, arg := range expr.Args {
		p.AcceptExpr(arg)
	}
}

func (p *printer) VisitIndex(expr *Index) {
	p.print("[]")

	p.AcceptExpr(expr.Value)
	p.AcceptExpr(expr.Index)
}

func (p *printer) VisitMember(expr *Member) {
	p.print(".%s", expr.Name)
	p.AcceptExpr(expr.Value)
}

// ast.Acceptor

func (p *printer) AcceptDecl(decl Decl) {
	if decl != nil {
		p.depth++
		decl.Accept(p)
		p.depth--
	}
}

func (p *printer) AcceptStmt(stmt Stmt) {
	if stmt != nil {
		p.depth++
		stmt.Accept(p)
		p.depth--
	}
}

func (p *printer) AcceptExpr(expr Expr) {
	if expr != nil {
		p.depth++
		expr.Accept(p)
		p.depth--
	}
}

// Utils

func (p *printer) print(format string, args ...any) {
	for i := 0; i < p.depth; i++ {
		_, _ = fmt.Fprint(p.writer, "  ")
	}

	_, _ = fmt.Fprintf(p.writer, format, args...)
	_, _ = fmt.Fprintln(p.writer)
}
