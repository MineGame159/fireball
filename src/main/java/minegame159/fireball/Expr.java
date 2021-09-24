package minegame159.fireball;

import java.util.List;

public abstract class Expr {
    interface Visitor {
        void visitBinaryExpr(Binary expr);
        void visitGroupingExpr(Grouping expr);
        void visitUnaryExpr(Unary expr);
    }

    public static class Binary extends Expr {
        public final Expr left;
        public final Token operator;
        public final Expr right;

        Binary(Expr left, Token operator, Expr right) {
            this.left = left;
            this.operator = operator;
            this.right = right;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitBinaryExpr(this);
        }
    }

    public static class Grouping extends Expr {
        public final Expr expression;

        Grouping(Expr expression) {
            this.expression = expression;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitGroupingExpr(this);
        }
    }

    public static class Unary extends Expr {
        public final Token operator;
        public final Expr right;

        Unary(Token operator, Expr right) {
            this.operator = operator;
            this.right = right;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitUnaryExpr(this);
        }
    }

    public abstract void accept(Visitor visitor);
}
