// This file is automatically generated, do not edit it manually

package minegame159.fireball;

import java.util.List;

public abstract class Stmt {
    interface Visitor {
        void visitExpressionStmt(Expression stmt);
    }

    public static class Expression extends Stmt {
        public final Expr expression;

        Expression(Expr expression) {
            this.expression = expression;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitExpressionStmt(this);
        }
    }

    public abstract void accept(Visitor visitor);
}