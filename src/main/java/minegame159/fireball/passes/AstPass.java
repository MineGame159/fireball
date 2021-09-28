package minegame159.fireball.passes;

import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.prototypes.ProtoFunction;

import java.util.List;

public abstract class AstPass implements Stmt.Visitor, Expr.Visitor {
    @Override
    public void visitFunctionStart(ProtoFunction proto) {}

    @Override
    public void visitFunctionEnd(ProtoFunction proto) {}

    // Accept

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
