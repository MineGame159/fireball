package minegame159.fireball;

import java.util.*;

public class Checker extends AstPass {
    private static class Variable {
        public boolean defined;
    }

    public final List<Error> errors = new ArrayList<>();

    private final Context context;

    private final Stack<Map<String, Variable>> scopes = new Stack<>();

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
        declare(stmt.name);
        if (stmt.initializer != null) acceptE(stmt.initializer);
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
        acceptS(stmt.body);
    }

    @Override
    public void visitReturnStmt(Stmt.Return stmt) {
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
        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        acceptE(expr.right);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
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
        if (getLocal(expr.name) == null) errors.add(new Error(expr.name, "Undeclared identifier '" + expr.name.lexeme() + "'."));

        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
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

    private void declare(Token name) {
        scopes.peek().put(name.lexeme(), new Variable());
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

    private Variable getLocalInScope(Token name) {
        return scopes.peek().get(name.lexeme());
    }
}
