package minegame159.fireball;

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

    public final List<Error> errors = new ArrayList<>();

    private final Context context;

    private final Stack<Map<String, Variable>> scopes = new Stack<>();
    private Function currentFunction;

    public Checker(Context context) {
        this.context = context;
    }

    public void check(List<Stmt> stmts) {
        acceptS(stmts);
    }

    // Statements

    @Override
    public void visitExpressionStmt(Stmt.Expression stmt) {
        acceptE(stmt.expression);
    }

    @Override
    public void visitBlockStmt(Stmt.Block stmt) {
        beginScope();
        acceptS(stmt.statements);
        endScope();
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        Type type = context.getType(stmt.type);
        if (type == null) errors.add(new Error(stmt.type, "Unknown type '" + stmt.type.lexeme() + "'."));
        else {
            Type valueType = stmt.initializer.getType();
            if (!type.equals(valueType)) errors.add(new Error(stmt.name, "Mismatched type, expected '" + type + "' but got '" + valueType + "'."));
        }

        declare(stmt.name, type);
        acceptE(stmt.initializer);
        define(stmt.name);
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
    public void visitFunctionStmt(Stmt.Function stmt) {
        currentFunction = context.getFunction(stmt.name);
        acceptS(stmt.body);
        currentFunction = null;
    }

    @Override
    public void visitReturnStmt(Stmt.Return stmt) {
        Type valueType = stmt.value.getType();
        if (!currentFunction.returnType().equals(valueType)) errors.add(new Error(stmt.token, "Mismatched type, expected '" + currentFunction.returnType() + "' but got '" + valueType + "'."));

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
        if (!expr.left.getType().isNumber() || !expr.right.getType().isNumber()) errors.add(new Error(expr.operator, "Operands of binary operations must be numbers."));

        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        if (expr.operator.type() == TokenType.Bang) {
            if (!expr.right.getType().isBool()) errors.add(new Error(expr.operator, "Operand of invert operation must be a boolean."));
        }
        else if (expr.operator.type() == TokenType.Minus) {
            if (!expr.right.getType().isNumber()) errors.add(new Error(expr.operator, "Operand of negate operation must be a number."));
        }

        acceptE(expr.right);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        if (!expr.left.getType().isBool() || !expr.right.getType().isBool()) errors.add(new Error(expr.operator, "Operands of logical operations must be booleans."));

        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        Variable var = getLocal(expr.name);

        if (var == null) {
            if (context.getFunction(expr.name) == null) errors.add(new Error(expr.name, "Undeclared identifier '" + expr.name.lexeme() + "'."));
        }
        else if (!var.defined) errors.add(new Error(expr.name, "Undefined variable '" + expr.name.lexeme() + "'."));
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        Variable var = getLocal(expr.name);

        if (var == null) errors.add(new Error(expr.name, "Undeclared identifier '" + expr.name.lexeme() + "'."));
        else if (!var.type.equals(expr.value.getType())) errors.add(new Error(expr.name, "Mismatched type, expected '" + var.type + "' but got '" + expr.value.getType() + "'."));

        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        if (!(expr.callee instanceof Expr.Variable)) errors.add(new Error(expr.token, "Invalid call target."));
        else {
            Function function = context.getFunction(((Expr.Variable) expr.callee).name);

            if (function != null) {
                if (function.params().size() != expr.arguments.size()) errors.add(new Error(expr.token, "Invalid number of arguments, expected " + function.params().size() + " but got " + expr.arguments.size() + "."));
                else {
                    for (int i = 0; i < function.params().size(); i++) {
                        Function.Param param = function.params().get(i);

                        Type argType = expr.arguments.get(i).getType();
                        if (!param.type().equals(argType)) errors.add(new Error(expr.token, "Mismatched type, expected '" + param.type() + "' but got '" + argType + "'."));
                    }
                }
            }
        }

        acceptE(expr.callee);
        acceptE(expr.arguments);
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
