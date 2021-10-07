package minegame159.fireball.passes;

import minegame159.fireball.context.*;
import minegame159.fireball.parser.*;
import minegame159.fireball.types.StructType;
import minegame159.fireball.types.Type;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.Writer;
import java.util.List;

public class Compiler extends AstPass {
    private final Context context;
    private final CompilerWriter w;

    private List<Expr> callArguments;

    private Compiler(Context context, Writer writer) {
        this.context = context;
        this.w = new CompilerWriter(writer);
    }

    public static void compile(Parser.Result result, Context context, Writer writer) {
        new Compiler(context, writer).compile(result);
    }

    private void compile(Parser.Result result) {
        // Standard library
        try (BufferedReader br = new BufferedReader(new FileReader("scripts/standard-lib.c"))) {
            String line;
            while ((line = br.readLine()) != null) w.writeln(line);
        }
        catch (Exception e) {
            System.err.println("Error locating standard library, compile failed.");
            e.printStackTrace();
            return;
        }

        // Forward declarations
        w.writeln("\n// Forward declarations\n");

        //     Structs
        for (Struct struct : context.getStructs()) {
            // Struct
            w.write("typedef struct ").write(struct.name).write(' ').write(struct.name).writeSemicolon();

            // Constructors
            for (Constructor constructor : struct.constructors) {
                writeFunctionDefinition(constructor);
                w.writeSemicolon();
            }

            // Destructor
            if (struct.destructor != null) {
                writeFunctionDefinition(struct.destructor);
                w.writeSemicolon();
            }

            // Methods
            for (Method method : struct.methods) {
                writeFunctionDefinition(method);
                w.writeSemicolon();
            }
        }
        if (context.getStructs().size() > 0) w.write('\n');

        //     Functions
        for (Function function : context.getFunctions()) {
            writeFunctionDefinition(function);
            w.writeSemicolon();
        }

        // User code
        w.writeln("\n// User code\n");

        //     Structs
        for (Struct struct : context.getStructs()) {
            w.write("//     ").write(struct.name).write("\n\n");

            w.indent().write("struct ").write(struct.name).write(" {\n");
            w.indentUp();

            // Fields
            for (Field field : struct.fields) {
                w.indent().write(field.type()).write(' ').write(field.name()).writeSemicolon();
            }

            w.indentDown();
            w.indent().write('}').writeSemicolon();

            w.write('\n');

            // Constructors
            for (Constructor constructor : struct.constructors) writeFunction(constructor);

            // Destructor
            if (struct.destructor != null) writeFunction(struct.destructor);

            // Methods
            for (Method method : struct.methods) writeFunction(method);
        }

        //     Functions
        w.write("//     Functions\n\n");
        for (Function function : context.getFunctions()) writeFunction(function);

        w.close();
    }

    private void writeFunctionDefinition(Function function) {
        w.write(function.returnType).write(' ').write(function.getOutputName()).write('(');

        for (int i = 0; i < function.params.size(); i++) {
            Function.Param param = function.params.get(i);

            if (i > 0) w.write(", ");
            w.write(param.type()).write(' ').write(param.name());
        }

        w.write(")");
    }

    private void writeFunction(Function function) {
        writeFunctionDefinition(function);
        w.write(' ');

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
        w.write(stmt.getType(context)).write(' ').write(stmt.name);

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

    @Override
    public void visitDeleteStmt(Stmt.Delete stmt) {
        Struct struct = ((StructType) stmt.expr.getType()).struct;

        // Run destructor
        if (struct.destructor != null) {
            w.write(struct.destructor.getOutputName()).write('(');
            acceptE(stmt.expr);
            w.write(')').writeSemicolon();

            w.indent();
        }

        // Free
        w.write("free(");
        acceptE(stmt.expr);
        w.write(')').writeSemicolon();
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
        w.write(' ').write(expr.operator).write(' ');
        acceptE(expr.right);
    }

    @Override
    public void visitCastExpr(Expr.Cast expr) {
        w.write('(').write(expr.getType()).write(") ");
        acceptE(expr.expr);
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        w.write(expr.operator);
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryPostExpr(Expr.UnaryPost expr) {
        acceptE(expr.left);
        w.write(expr.operator);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        acceptE(expr.left);
        w.write(' ').write(expr.operator).write(' ');
        acceptE(expr.right);
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        String name = expr.name.lexeme();

        if (callArguments != null && expr.getType() instanceof StructType structType) {
            // getConstructor() should never return null because it was validated in previous AST pass
            name = structType.struct.getConstructor(false, callArguments).getOutputName();
        }

        w.write(name);
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        if (expr.getType().isPointer()) w.write('*');
        w.write(expr.name).write(" = ");

        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        if (expr.callee instanceof Expr.Variable) {
            List<Expr> preCallArguments = callArguments;
            callArguments = expr.arguments;
            acceptE(expr.callee);
            callArguments = preCallArguments;

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

            if (!((Expr.Get) expr.callee).object.getType().isPointer()) w.write('&');
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
        w.write(expr.object.getType().isPointer() ? "->" : ".").write(expr.name);
    }

    @Override
    public void visitSetExpr(Expr.Set expr) {
        acceptE(expr.object);
        w.write(expr.object.getType().isPointer() ? "->" : ".").write(expr.name).write(" = ");
        acceptE(expr.value);
    }

    @Override
    public void visitNewExpr(Expr.New expr) {
        Struct struct = ((StructType) expr.getType()).struct;
        Constructor constructor = struct.getConstructor(true, expr.arguments);

        w.write(constructor.getOutputName()).write('(');

        for (int i = 0; i < expr.arguments.size(); i++) {
            if (i > 0) w.write(", ");
            acceptE(expr.arguments.get(i));
        }

        w.write(')');
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
