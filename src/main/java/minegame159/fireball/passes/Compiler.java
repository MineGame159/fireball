package minegame159.fireball.passes;

import minegame159.fireball.context.Context;
import minegame159.fireball.context.Function;
import minegame159.fireball.context.Method;
import minegame159.fireball.context.Struct;
import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Parser;
import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.prototypes.ProtoFunction;
import minegame159.fireball.parser.prototypes.ProtoMethod;
import minegame159.fireball.parser.prototypes.ProtoParameter;
import minegame159.fireball.parser.prototypes.ProtoStruct;
import minegame159.fireball.types.StructType;
import minegame159.fireball.types.Type;

import java.io.Writer;
import java.util.List;

public class Compiler extends AstPass {
    private final Context context;
    private final CompilerWriter w;

    private Compiler(Context context, Writer writer) {
        this.context = context;
        this.w = new CompilerWriter(writer);
    }

    public static void compile(Parser.Result result, Context context, Writer writer) {
        new Compiler(context, writer).compile(result);
    }

    private void compile(Parser.Result result) {
        // Standard library
        w.writeln("\n// Standard library\n");
        w.writeln("#include <stdbool.h>");
        w.writeln("#include <stdint.h>");
        w.writeln("#include <stdio.h>\n");

        w.writeln("typedef uint8_t u8;");
        w.writeln("typedef uint16_t u16;");
        w.writeln("typedef uint32_t u32;");
        w.writeln("typedef uint64_t u64;");
        w.writeln("typedef int8_t i8;");
        w.writeln("typedef int16_t i16;");
        w.writeln("typedef int32_t i32;");
        w.writeln("typedef int64_t i64;");
        w.writeln("typedef float f32;");
        w.writeln("typedef double f64;");

        // Forward declarations
        w.writeln("\n// Forward declarations\n");

        //     Structs
        for (Struct struct : context.getStructs()) {
            w.write("typedef struct ").write(struct.name().lexeme()).write(' ').write(struct.name().lexeme()).writeSemicolon();

            for (Method method : struct.methods()) writeFunctionTypedef(method);
        }
        if (context.getStructs().size() > 0) w.write('\n');

        //     Functions
        for (Function function : context.getFunctions()) writeFunctionTypedef(function);

        // User code
        w.writeln("\n// User code\n");

        //     Structs
        for (ProtoStruct struct : result.structs) {
            w.write("struct ").write(struct.name().lexeme()).write(" {\n");
            w.indentUp();

            for (ProtoParameter field : struct.fields()) {
                w.indent().write(field.type().toString()).write(' ').write(field.name().lexeme()).writeSemicolon();
            }

            w.indentDown();
            w.indent().write("}").writeSemicolon();

            w.write('\n');

            for (ProtoMethod method : struct.methods()) {
                writeFunction(method, "_" + struct.name() + "__" + method.name);
            }
        }

        //     Functions
        for (ProtoFunction function : result.functions) {
            writeFunction(function, function.name.lexeme());
        }

        w.close();
    }

    private void writeFunctionTypedef(Function function) {
        w.write(function.returnType.name).write(' ').write(function.getOutputName()).write('(');

        for (int i = 0; i < function.params.size(); i++) {
            Function.Param param = function.params.get(i);

            if (i > 0) w.write(", ");
            w.write(param.type().name).write(' ').write(param.name().lexeme());
        }

        w.writeln(");");
    }

    private void writeFunction(ProtoFunction function, String name) {
        w.write(function.returnType.toString()).write(' ').write(name).write('(');

        for (int i = 0; i < function.params.size(); i++) {
            ProtoParameter param = function.params.get(i);

            if (i > 0) w.write(", ");
            w.write(param.type().toString()).write(' ').write(param.name().lexeme());
        }

        w.write(") ");
        function.accept(this);
        w.write('\n');
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
        w.write(stmt.getType(context).name).write(' ').write(stmt.name.lexeme());

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
        w.write('"').write(expr.value.replace("\n", "\\n")).write('"');
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
        if (expr.callee instanceof Expr.Variable) {
            acceptE(expr.callee);
            w.write('(');

            for (int i = 0; i < expr.arguments.size(); i++) {
                if (i > 0) w.write(", ");
                acceptE(expr.arguments.get(i));
            }

            w.write(')');
        }
        else {
            Type type = ((Expr.Get) expr.callee).object.getType();
            Struct struct = ((StructType) type).struct;
            Method method = struct.getMethod(((Expr.Get) expr.callee).name);

            w.write(method.getOutputName());
            w.write('(');

            w.write('&');
            acceptE(((Expr.Get) expr.callee).object);

            for (int i = 0; i < expr.arguments.size(); i++) {
                w.write(", ");
                acceptE(expr.arguments.get(i));
            }

            w.write(')');
        }
    }

    @Override
    public void visitGetExpr(Expr.Get expr) {
        acceptE(expr.object);
        w.write(expr.object.getType().isPointer() ? "->" : ".").write(expr.name.lexeme());
    }

    @Override
    public void visitSetExpr(Expr.Set expr) {
        acceptE(expr.object);
        w.write(expr.object.getType().isPointer() ? "->" : ".").write(expr.name.lexeme()).write(" = ");
        acceptE(expr.value);
    }

    // Accept


    @Override
    protected void acceptS(Stmt stmt) {
        w.indent();
        super.acceptS(stmt);
    }

    @Override
    protected void acceptS(List<Stmt> stmts) {
        for (Stmt stmt : stmts) {
            w.indent();
            stmt.accept(this);
        }
    }
}
