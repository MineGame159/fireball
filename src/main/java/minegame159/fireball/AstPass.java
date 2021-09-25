package minegame159.fireball;

import java.util.List;

public abstract class AstPass implements Stmt.Visitor, Expr.Visitor {
    protected void acceptS(Stmt stmt) {
        if (stmt != null) stmt.accept(this);
    }

    protected void acceptS(List<Stmt> stmts) {
        for (Stmt stmt : stmts) {
            stmt.accept(this);
        }
    }

    protected void acceptE(Expr expr) {
        if (expr != null) expr.accept(this);
    }

    protected void acceptE(List<Expr> exprs) {
        for (Expr expr : exprs) expr.accept(this);
    }
}
