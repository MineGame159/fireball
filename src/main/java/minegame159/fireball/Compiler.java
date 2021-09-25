package minegame159.fireball;

import java.io.Writer;
import java.util.List;

public class Compiler implements Stmt.Visitor, Expr.Visitor {
    private final CompilerWriter w;

    public Compiler(Writer writer) {
        this.w = new CompilerWriter(writer);
    }

    public void compile(List<Stmt> stmts) {
        w.writeln("\n// Standard library\n");
        w.writeln("#include <stdbool.h>");
        w.writeln("#include <stdint.h>");
        w.writeln("#include <stdio.h>\n");

        w.writeln("typedef int8_t i8;");
        w.writeln("typedef int16_t i16;");
        w.writeln("typedef int32_t i32;");
        w.writeln("typedef int64_t i64;");
        w.writeln("typedef int8_t u8;");
        w.writeln("typedef uint16_t u16;");
        w.writeln("typedef uint32_t u32;");
        w.writeln("typedef uint64_t u64;");
        w.writeln("typedef float f32;");
        w.writeln("typedef double f64;");

        w.writeln("\n// User code\n");
        acceptS(stmts);

        w.close();
    }

    // Statements

    @Override
    public void visitExpressionStmt(Stmt.Expression stmt) {
        acceptE(stmt.expression);
        w.writeSemicolon();
    }

    @Override
    public void visitBlockStmt(Stmt.Block stmt) {
        if (stmt.statements.isEmpty()) {
            w.writeln("{}");
            return;
        }

        w.writeln('{');
        w.indentUp();

        acceptS(stmt.statements);

        w.indentDown();
        w.indent().writeln('}');
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        w.write(stmt.type.lexeme()).write(' ').write(stmt.name.lexeme());

        if (stmt.initializer != null) {
            w.write(" = ");
            acceptE(stmt.initializer);
        }

        w.writeSemicolon();
    }

    @Override
    public void visitIfStmt(Stmt.If stmt) {
        w.write("if (");
        acceptE(stmt.condition);
        w.write(") ");

        if (stmt.thenBranch != null) {
            w.skipIndent();
            acceptS(stmt.thenBranch);
        }

        if (stmt.elseBranch != null) {
            w.indent().write("else ").skipIndent();
            acceptS(stmt.elseBranch);
        }
    }

    @Override
    public void visitWhileStmt(Stmt.While stmt) {
        w.write("while (");
        acceptE(stmt.condition);
        w.write(") ");

        w.skipIndent();
        acceptS(stmt.body);
    }

    @Override
    public void visitForStmt(Stmt.For stmt) {
        w.write("for (");

        if (stmt.initializer != null) {
            w.skipIndent().skipNewLine();
            acceptS(stmt.initializer);
            w.write(' ');
        }
        else w.write("; ");

        if (stmt.condition != null) acceptE(stmt.condition);
        w.write("; ");

        if (stmt.increment != null) acceptE(stmt.increment);
        w.write(") ");

        w.skipIndent();
        acceptS(stmt.body);
    }

    @Override
    public void visitFunctionStmt(Stmt.Function stmt) {
        w.write(stmt.returnType.lexeme()).write(' ').write(stmt.name.lexeme()).write('(');

        for (int i = 0; i < stmt.params.size(); i++) {
            TokenPair param = stmt.params.get(i);

            if (i > 0) w.write(", ");
            w.write(param.first().lexeme()).write(' ').write(param.second().lexeme());
        }

        w.write(") ");
        acceptS(stmt.body);
        w.write('\n');
    }

    @Override
    public void visitReturnStmt(Stmt.Return stmt) {
        w.write("return");

        if (stmt.value != null) {
            w.write(' ');
            acceptE(stmt.value);
        }

        w.writeSemicolon();
    }

    @Override
    public void visitCBlockStmt(Stmt.CBlock stmt) {
        w.write(stmt.code).write('\n');
    }

    // Expressions

    @Override
    public void visitNullExpr(Expr.Null expr) {
        w.write("NULL");
    }

    @Override
    public void visitBoolExpr(Expr.Bool expr) {
        w.write(expr.value ? "true" : "false");
    }

    @Override
    public void visitUnsignedIntExpr(Expr.UnsignedInt expr) {
        w.write(Integer.toUnsignedString((int) expr.value));
    }

    @Override
    public void visitIntExpr(Expr.Int expr) {
        w.write(Integer.toString((int) expr.value));
    }

    @Override
    public void visitFloatExpr(Expr.Float expr) {
        w.write(Double.toString(expr.value));
    }

    @Override
    public void visitStringExpr(Expr.String expr) {
        w.write('"').write(expr.value).write('"');
    }

    @Override
    public void visitGroupingExpr(Expr.Grouping expr) {
        w.write('(');
        acceptE(expr.expression);
        w.write(')');
    }

    @Override
    public void visitBinaryExpr(Expr.Binary expr) {
        acceptE(expr.left);
        w.write(' ').write(expr.operator.lexeme()).write(' ');
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        w.write(expr.operator.lexeme());
        acceptE(expr.right);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        acceptE(expr.left);
        w.write(' ').write(expr.operator.lexeme()).write(' ');
        acceptE(expr.right);
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        w.write(expr.name.lexeme());
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        w.write(expr.name.lexeme()).write(" = ");
        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        acceptE(expr.callee);
        w.write('(');

        for (int i = 0; i < expr.arguments.size(); i++) {
            if (i > 0) w.write(", ");
            acceptE(expr.arguments.get(i));
        }

        w.write(')');
    }


    // Utils

    // Accept

    private void acceptS(Stmt stmt) {
        w.indent();
        stmt.accept(this);
    }

    private void acceptS(List<Stmt> stmts) {
        for (Stmt stmt : stmts) {
            w.indent();
            stmt.accept(this);
        }
    }

    private void acceptE(Expr expr) {
        expr.accept(this);
    }

    private void acceptE(List<Expr> exprs) {
        for (Expr expr : exprs) expr.accept(this);
    }
}
