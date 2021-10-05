package minegame159.fireball.passes;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.context.Context;
import minegame159.fireball.context.Function;
import minegame159.fireball.context.Method;
import minegame159.fireball.context.Struct;
import minegame159.fireball.parser.*;
import minegame159.fireball.parser.prototypes.ProtoFunction;
import minegame159.fireball.parser.prototypes.ProtoParameter;
import minegame159.fireball.types.PrimitiveTypes;
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
    private boolean hasTopLevelReturn;

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
        // Begin function
        beginScope();
        skipBlockScopes = proto.body instanceof Stmt.Block;
        currentFunction = context.getFunction(proto.name);
        hasTopLevelReturn = false;

        // Declare and define parameters
        for (ProtoParameter param : proto.params) {
            declare(param.name(), context.getType(param.type()));
            define(param.name());
        }
    }

    @Override
    public void visitFunctionEnd(ProtoFunction proto) {
        // Check for return statement if function returns something
        if (!hasTopLevelReturn && context.getType(proto.returnType) != PrimitiveTypes.Void.type) errors.add(Errors.missingReturn(proto.name));

        // End function
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
        // Begin scope
        if (!skipBlockScopes) beginScope();

        // Check statements inside the block
        acceptS(stmt.statements);

        // End scope
        if (!skipBlockScopes) endScope();
    }

    @Override
    public void visitVariableStmt(Stmt.Variable stmt) {
        Type type = stmt.getType(context);

        // Declare variable
        declare(stmt.name, type);

        // Check initializer
        acceptE(stmt.initializer);

        // Define variable
        if (stmt.initializer != null || context.getType(stmt.type) instanceof StructType) define(stmt.name);

        // Check expected initializer type
        if (stmt.initializer != null) {
            Type valueType = stmt.initializer.getType();
            if (!type.equals(valueType)) errors.add(Errors.mismatchedType(stmt.name, type, valueType));
        }
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
        // Check expected return type
        Type valueType = stmt.value.getType();
        if (!currentFunction.returnType.equals(valueType)) errors.add(Errors.mismatchedType(stmt.token, currentFunction.returnType, valueType));

        // Check function body
        acceptE(stmt.value);

        if (scopes.size() <= 1) hasTopLevelReturn = true;
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

        // Can only apply binary operations to numbers
        if (!expr.left.getType().isNumber() || !expr.right.getType().isNumber()) errors.add(Errors.wrongOperands(expr.operator, "binary", "number", true));
    }

    @Override
    public void visitUnaryExpr(Expr.Unary expr) {
        acceptE(expr.right);

        // Can only invert booleans
        if (expr.operator.type() == TokenType.Bang) {
            if (!expr.right.getType().isBool()) errors.add(Errors.wrongOperands(expr.operator, "invert", "boolean", false));
        }
        // Can only negate numbers
        else if (expr.operator.type() == TokenType.Minus) {
            if (!expr.right.getType().isNumber()) errors.add(Errors.wrongOperands(expr.operator, "negate", "number", false));
        }
        // Check invalid pointer target
        else if (expr.operator.type() == TokenType.Ampersand) {
            // Can only take address of a variable or a field
            if (!(expr.right instanceof Expr.Variable) && !(expr.right instanceof Expr.Get)) errors.add(Errors.invalidPointerTarget(expr.operator));
                // Cannot take address of a pointer
            else if (expr.right.getType().isPointer()) errors.add(Errors.invalidPointerTarget(expr.operator));
        }
    }

    @Override
    public void visitLogicalExpr(Expr.Logical expr) {
        acceptE(expr.left);
        acceptE(expr.right);

        // Can only apply logical operations to booleans
        if (!expr.left.getType().isBool() || !expr.right.getType().isBool()) errors.add(Errors.wrongOperands(expr.operator, "logical", "boolean", true));
    }

    @Override
    public void visitVariableExpr(Expr.Variable expr) {
        // If variable is local variable check if it's defined
        Variable var = getLocal(expr.name);
        if (var != null && !var.defined) errors.add(Errors.undefined(expr.name));
    }

    @Override
    public void visitAssignExpr(Expr.Assign expr) {
        // Check value
        acceptE(expr.value);

        // Check expected type
        Variable var = getLocal(expr.name);

        if (var == null) errors.add(Errors.undeclared(expr.name));
        else if (!var.type.equals(expr.value.getType())) errors.add(Errors.mismatchedType(expr.name, var.type, expr.value.getType()));

        // Define variable
        define(expr.name);
    }

    @Override
    public void visitCallExpr(Expr.Call expr) {
        acceptE(expr.callee);
        acceptE(expr.arguments);

        // This is horrible but i cba to clean it up now
        if (!(expr.callee instanceof Expr.Variable || expr.callee instanceof Expr.Get)) errors.add(Errors.invalidCallTarget(expr.token));
        else {
            Function function = null;
            boolean check = true;

            if (expr.callee instanceof Expr.Variable) function = context.getFunction(((Expr.Variable) expr.callee).name);
            else {
                Type type = ((Expr.Get) expr.callee).object.getType();

                if (!(type instanceof StructType)) {
                    errors.add(Errors.invalidCallTarget(expr.token));
                    check = false;
                }
                else {
                    Struct struct = ((StructType) type).struct;
                    function = struct.getMethod(((Expr.Get) expr.callee).name);
                }
            }

            if (function != null && check) {
                boolean isMethod = function instanceof Method;
                int argCount = expr.arguments.size() + (isMethod ? 1 : 0);

                if (function.params.size() != argCount) errors.add(Errors.wrongArgumentCount(expr.token, function.params.size(), argCount));
                else {
                    for (int i = 0; i < function.params.size(); i++) {
                        Function.Param param = function.params.get(i);

                        Type argType;
                        if (isMethod) {
                            if (i == 0) argType = ((Expr.Get) expr.callee).object.getType().pointer();
                            else argType = expr.arguments.get(i - 1).getType();
                        }
                        else argType = expr.arguments.get(i).getType();

                        if (!param.type().equals(argType)) errors.add(Errors.mismatchedType(expr.token, param.type(), argType));
                    }
                }
            }
        }
    }

    @Override
    public void visitGetExpr(Expr.Get expr) {
        acceptE(expr.object);
    }

    @Override
    public void visitSetExpr(Expr.Set expr) {
        acceptE(expr.object);
        acceptE(expr.value);

        if (expr.getType() != null) {
            // Check if expression can be assigned to the field
            if (!expr.getType().equals(expr.value.getType())) errors.add(Errors.mismatchedType(expr.name, expr.getType(), expr.value.getType()));
        }
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
        Variable var = getLocal(name);
        if (var != null) var.defined = true;
    }

    private Variable getLocal(Token name) {
        for (int i = scopes.size() - 1; i >= 0; i--) {
            Variable var = scopes.get(i).get(name.lexeme());
            if (var != null) return var;
        }

        return null;
    }
}
