package minegame159.fireball.parser;

import minegame159.fireball.TokenPair;
import minegame159.fireball.context.Context;
import minegame159.fireball.types.Type;

import java.util.List;

public abstract class Stmt {
    public interface Visitor {
        void visitExpressionStmt(Expression stmt);
        void visitBlockStmt(Block stmt);
        void visitVariableStmt(Variable stmt);
        void visitIfStmt(If stmt);
        void visitWhileStmt(While stmt);
        void visitForStmt(For stmt);
        void visitFunctionStmt(Function stmt);
        void visitReturnStmt(Return stmt);
        void visitCBlockStmt(CBlock stmt);
    }

    public static class Expression extends Stmt {
        public final Expr expression;

        public Expression(Expr expression) {
            this.expression = expression;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitExpressionStmt(this);
        }
    }

    public static class Block extends Stmt {
        public final List<Stmt> statements;

        public Block(List<Stmt> statements) {
            this.statements = statements;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitBlockStmt(this);
        }
    }

    public static class Variable extends Stmt {
        public final Token type;
        public final Token name;
        public final Expr initializer;

        public Variable(Token type, Token name, Expr initializer) {
            this.type = type;
            this.name = name;
            this.initializer = initializer;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitVariableStmt(this);
        }

        public Type getType(Context context) {
            return type.type() == TokenType.Var ? initializer.getType() : context.getType(type);
        }
    }

    public static class If extends Stmt {
        public final Expr condition;
        public final Stmt thenBranch;
        public final Stmt elseBranch;

        public If(Expr condition, Stmt thenBranch, Stmt elseBranch) {
            this.condition = condition;
            this.thenBranch = thenBranch;
            this.elseBranch = elseBranch;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitIfStmt(this);
        }
    }

    public static class While extends Stmt {
        public final Expr condition;
        public final Stmt body;

        public While(Expr condition, Stmt body) {
            this.condition = condition;
            this.body = body;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitWhileStmt(this);
        }
    }

    public static class For extends Stmt {
        public final Stmt initializer;
        public final Expr condition;
        public final Expr increment;
        public final Stmt body;

        public For(Stmt initializer, Expr condition, Expr increment, Stmt body) {
            this.initializer = initializer;
            this.condition = condition;
            this.increment = increment;
            this.body = body;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitForStmt(this);
        }
    }

    public static class Function extends Stmt {
        public final Token returnType;
        public final Token name;
        public final List<TokenPair> params;
        public final Stmt body;

        public Function(Token returnType, Token name, List<TokenPair> params, Stmt body) {
            this.returnType = returnType;
            this.name = name;
            this.params = params;
            this.body = body;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitFunctionStmt(this);
        }
    }

    public static class Return extends Stmt {
        public final Token token;
        public final Expr value;

        public Return(Token token, Expr value) {
            this.token = token;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitReturnStmt(this);
        }
    }

    public static class CBlock extends Stmt {
        public final String code;

        public CBlock(String code) {
            this.code = code;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitCBlockStmt(this);
        }
    }

    public abstract void accept(Visitor visitor);
}
