package minegame159.fireball.passes;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.context.*;
import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Parser;
import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;
import minegame159.fireball.parser.prototypes.ProtoFunction;
import minegame159.fireball.parser.prototypes.ProtoParameter;
import minegame159.fireball.parser.prototypes.ProtoType;
import minegame159.fireball.types.StructType;
import minegame159.fireball.types.Type;

import java.util.*;

public class TypeResolver extends AstPass {
    private final List<Error> errors = new ArrayList<>();

    private final Context context;

    private final Stack<Map<String, Type>> scopes = new Stack<>();
    private boolean skipBlockScopes;

    private TypeResolver(Context context) {
        this.context = context;
    }

    public static List<Error> resolve(Parser.Result result, Context context) {
        TypeResolver resolver = new TypeResolver(context);
        result.accept(resolver);
        return resolver.errors;
    }

    @Override
    public void visitFunctionStart(ProtoFunction proto) {
        // Begin function
        scopes.push(new HashMap<>());
        skipBlockScopes = proto.body instanceof Stmt.Block;

        // Declare and define parameters
        for (ProtoParameter param : proto.params) {
            declare(param.name(), context.getType(param.type()));
        }
    }

    @Override
    public void visitFunctionEnd(ProtoFunction proto) {
        // End function
        scopes.pop();
        skipBlockScopes = false;
    }

    // Statements

    @Override
    public void visitExpressionStmt(Stmt.Expression stmt) {
        acceptE(stmt.expression);
    }

    @Override
    public void visitBlockStmt(Stmt.Block stmt) {
        // Begin scope
        if (!skipBlockScopes) scopes.push(new HashMap<>());

        // Resolve statements inside the block
        acceptS(stmt.statements);

        // End scope
        if (!skipBlockScopes) scopes.pop();
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        acceptE(stmt.initializer);

        // Declare variable
        Type type = stmt.getType(context);
        if (type == null) errors.add(Errors.unknownType(stmt.type.name(), stmt.type));

        declare(stmt.name, type);
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
    public void visitCastExpr(Expr.Cast expr) {
        acceptE(expr.expr);

        // Resolve target type
        expr._type = context.getType(expr.type);
        if (expr._type == null) errors.add(Errors.unknownType(expr.type.name(), expr.type));
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        acceptE(expr.right);
    }

    @Override
    public void visitUnaryPostExpr(Expr.UnaryPost expr) {
        acceptE(expr.left);
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        acceptE(expr.left);
        acceptE(expr.right);
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        // Resolve variable type
        expr.type = resolveIdentifierType(expr.name);
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        acceptE(expr.value);

        // Resolve variable type
        expr.type = resolveIdentifierType(expr.name);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        acceptE(expr.callee);
        acceptE(expr.arguments);
    }

    @Override
    public void visitGetExpr(Expr.Get expr) {
        acceptE(expr.object);

        // Resolve field type
        expr.type = resolveFieldType(expr.object, expr.name);
    }

    @Override
    public void visitSetExpr(Expr.Set expr) {
        acceptE(expr.object);
        acceptE(expr.value);

        // Resolve field type
        expr.type = resolveFieldType(expr.object, expr.name);
    }

    // Scope

    private void declare(Token name, Type type) {
        scopes.peek().put(name.lexeme(), type);
    }

    private Type getLocal(Token name) {
        for (int i = scopes.size() - 1; i >= 0; i--) {
            Type type = scopes.get(i).get(name.lexeme());
            if (type != null) return type;
        }

        return null;
    }

    // Utils

    private Type resolveFieldType(Expr object, Token name) {
        if (!(object.getType() instanceof StructType)) errors.add(Errors.invalidFieldTarget(name));
        else {
            Struct struct = ((StructType) object.getType()).struct;
            Field field = struct.getField(name);

            if (field == null) {
                Method method = struct.getMethod(name);

                if (method == null) errors.add(Errors.unknownField(struct.name(), name));
                else return method.returnType;
            }
            else return field.type();
        }

        return null;
    }

    private Type resolveIdentifierType(Token name) {
        // Local variable
        Type type = getLocal(name);

        if (type != null) return type;

        // Function
        Function function = context.getFunction(name);
        if (function != null) {
            if (function.returnType == null) errors.add(Errors.couldNotGetType(name));
            return function.returnType;
        }

        // Constructor
        Struct struct = context.getStruct(name);
        if (struct != null) {
            return context.getType(new ProtoType(name));
        }

        errors.add(Errors.couldNotGetType(name));
        return null;
    }
}
