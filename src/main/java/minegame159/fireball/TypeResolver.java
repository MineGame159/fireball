package minegame159.fireball;

import minegame159.fireball.types.Type;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Stack;

public class TypeResolver extends AstPass {
    public List<Error> errors;

    private final Context context;

    private final Stack<Map<String, Type>> scopes = new Stack<>();

    public TypeResolver(Context context) {
        this.context = context;
    }

    public void resolve(List<Stmt> stmts) {
        errors = context.resolveTypes();
        acceptS(stmts);
    }

    // Statements

    @Override
    public void visitExpressionStmt(Stmt.Expression stmt) {
        acceptE(stmt.expression);
    }

    @Override
    public void visitBlockStmt(Stmt.Block stmt) {
        scopes.push(new HashMap<>());
        acceptS(stmt.statements);
        scopes.pop();
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        if (stmt.type.type() == TokenType.Var) throw new RuntimeException("Automatic type detection for variables is not supported."); // TODO

        scopes.peek().put(stmt.name.lexeme(), context.getType(stmt.type));

        acceptE(stmt.initializer);
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
        expr.type = resolveType(expr.name);
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        expr.type = resolveType(expr.name);

        acceptE(expr.value);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        acceptE(expr.callee);
        acceptE(expr.arguments);
    }

    // Utils

    private Type resolveType(Token name) {
        Type type = getLocalType(name);

        if (type == null) {
            Function function = context.getFunction(name);
            if (function != null) type = function.returnType();
        }

        return type;
    }

    private Type getLocalType(Token name) {
        for (int i = scopes.size() - 1; i >= 0; i--) {
            Type type = scopes.get(i).get(name.lexeme());
            if (type != null) return type;
        }

        return null;
    }
}
