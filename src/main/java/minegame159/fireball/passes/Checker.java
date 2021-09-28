package minegame159.fireball.passes;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.context.Context;
import minegame159.fireball.context.Function;
import minegame159.fireball.parser.*;
import minegame159.fireball.parser.prototypes.ProtoFunction;
import minegame159.fireball.parser.prototypes.ProtoParameter;
import minegame159.fireball.types.StructType;
import minegame159.fireball.types.Type;

import java.util.*;

public class Checker extends AstPass {
    private static class Variable {
        public final Type type;
        public boolean defined;

        public Variable(Type type) {
            this.type = type;
        }
    }

    private final List<Error> errors = new ArrayList<>();

    private final Context context;

    private final Stack<Map<String, Variable>> scopes = new Stack<>();
    private boolean skipBlockScopes;
    private Function currentFunction;

    private Checker(Context context) {
        this.context = context;
    }

    public static List<Error> check(Parser.Result result, Context context) {
        Checker checker = new Checker(context);
        result.accept(checker);
        return checker.errors;
    }

    @Override
    public void visitFunctionStart(ProtoFunction proto) {
        beginScope();
        skipBlockScopes = proto.body() instanceof Stmt.Block;
        currentFunction = context.getFunction(proto.name());

        for (ProtoParameter param : proto.params()) {
            declare(param.name(), context.getType(param.type()));
            define(param.name());
        }
    }

    @Override
    public void visitFunctionEnd(ProtoFunction proto) {
        endScope();
        skipBlockScopes = false;
        currentFunction = null;
    }

    // Statements

    @Override
    public void visitExpressionStmt(Stmt.Expression stmt) {
        acceptE(stmt.expression);
    }

    @Override
    public void visitBlockStmt(Stmt.Block stmt) {
        if (!skipBlockScopes) beginScope();
        acceptS(stmt.statements);
        if (!skipBlockScopes) endScope();
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        Type type = stmt.getType(context);

        if (stmt.initializer != null) {
            Type valueType = stmt.initializer.getType();
            if (!type.equals(valueType)) errors.add(Errors.mismatchedType(stmt.name, type, valueType));
        }

        declare(stmt.name, type);
        acceptE(stmt.initializer);
        if (stmt.initializer != null || context.getType(stmt.type) instanceof StructType) define(stmt.name);
    }

    @Override
    public void visitIfStmt(Stmt.If stmt) {
        acceptE(stmt.condition);
        acceptS(stmt.thenBranch);
        acceptS(stmt.elseBranch);
    }

    @Override
    public void visitWhileStmt(Stmt.While stmt) {
        acceptE(stmt.condition);
        acceptS(stmt.body);
    }

    @Override
    public void visitForStmt(Stmt.For stmt) {
        acceptS(stmt.initializer);
        acceptE(stmt.condition);
        acceptE(stmt.increment);
        acceptS(stmt.body);
    }

    @Override
    public void visitReturnStmt(Stmt.Return stmt) {
        Type valueType = stmt.value.getType();
        if (!currentFunction.returnType().equals(valueType)) errors.add(Errors.mismatchedType(stmt.token, currentFunction.returnType(), valueType));

        acceptE(stmt.value);
    }

    @Override
    public void visitCBlockStmt(Stmt.CBlock stmt) {}

    // Expressions

    @Override
    public void visitNullExpr(Expr.Null expr) {}

    @Override
    public void visitBoolExpr(Expr.Bool expr) {}

    @Override
    public void visitUnsignedIntExpr(Expr.UnsignedInt expr) {}

    @Override
    public void visitIntExpr(Expr.Int expr) {}

    @Override
    public void visitFloatExpr(Expr.Float expr) {}

    @Override
    public void visitStringExpr(Expr.String expr) {}

    @Override
    public void visitGroupingExpr(Expr.Grouping expr) {
        acceptE(expr.expression);
    }

    @Override
    public void visitBinaryExpr(Expr.Binary expr) {
        if (!expr.left.getType().isNumber() || !expr.right.getType().isNumber()) errors.add(Errors.wrongOperands(expr.operator, "binary", "number", true));

        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        if (expr.operator.type() == TokenType.Bang) {
            if (!expr.right.getType().isBool()) errors.add(Errors.wrongOperands(expr.operator, "invert", "boolean", false));
        }
        else if (expr.operator.type() == TokenType.Minus) {
            if (!expr.right.getType().isNumber()) errors.add(Errors.wrongOperands(expr.operator, "negate", "number", false));
        }

        acceptE(expr.right);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        if (!expr.left.getType().isBool() || !expr.right.getType().isBool()) errors.add(Errors.wrongOperands(expr.operator, "logical", "boolean", true));

        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        Variable var = getLocal(expr.name);

        if (var == null) {
            if (context.getFunction(expr.name) == null) errors.add(Errors.undeclared(expr.name));
        }
        else if (!var.defined) errors.add(Errors.undefined(expr.name));
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        Variable var = getLocal(expr.name);

        if (var == null) errors.add(Errors.undeclared(expr.name));
        else if (!var.type.equals(expr.value.getType())) errors.add(Errors.mismatchedType(expr.name, var.type, expr.value.getType()));

        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        if (!(expr.callee instanceof Expr.Variable)) errors.add(Errors.invalidCallTarget(expr.token));
        else {
            Function function = context.getFunction(((Expr.Variable) expr.callee).name);

            if (function != null) {
                if (function.params().size() != expr.arguments.size()) errors.add(Errors.wrongArgumentCount(expr.token, function.params().size(), expr.arguments.size()));
                else {
                    for (int i = 0; i < function.params().size(); i++) {
                        Function.Param param = function.params().get(i);

                        Type argType = expr.arguments.get(i).getType();
                        if (!param.type().equals(argType)) errors.add(Errors.mismatchedType(expr.token, param.type(), argType));
                    }
                }
            }
        }

        acceptE(expr.callee);
        acceptE(expr.arguments);
    }

    @Override
    public void visitGetExpr(Expr.Get expr) {
        acceptE(expr.object);
    }

    @Override
    public void visitSetExpr(Expr.Set expr) {
        acceptE(expr.object);
        acceptE(expr.value);

        if (!expr.getType().equals(expr.value.getType())) errors.add(Errors.mismatchedType(expr.name, expr.getType(), expr.value.getType()));
    }

    // Scope

    private void beginScope() {
        scopes.push(new HashMap<>());
    }

    private void endScope() {
        scopes.pop();
    }

    private void declare(Token name, Type type) {
        scopes.peek().put(name.lexeme(), new Variable(type));
    }

    private void define(Token name) {
        scopes.peek().get(name.lexeme()).defined = true;
    }

    private Variable getLocal(Token name) {
        for (int i = scopes.size() - 1; i >= 0; i--) {
            Variable var = scopes.get(i).get(name.lexeme());
            if (var != null) return var;
        }

        return null;
    }
}
