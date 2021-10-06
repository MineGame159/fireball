// This file is automatically generated, do not edit it manually

package minegame159.fireball.parser;

import minegame159.fireball.parser.prototypes.ProtoType;
import minegame159.fireball.types.PrimitiveTypes;
import minegame159.fireball.types.Type;

import java.util.List;

public abstract class Expr {
    public interface Visitor {
        void visitNullExpr(Null expr);
        void visitBoolExpr(Bool expr);
        void visitUnsignedIntExpr(UnsignedInt expr);
        void visitIntExpr(Int expr);
        void visitFloatExpr(Float expr);
        void visitStringExpr(String expr);
        void visitGroupingExpr(Grouping expr);
        void visitBinaryExpr(Binary expr);
        void visitCastExpr(Cast expr);
        void visitUnaryExpr(Unary expr);
        void visitUnaryPostExpr(UnaryPost expr);
        void visitLogicalExpr(Logical expr);
        void visitVariableExpr(Variable expr);
        void visitAssignExpr(Assign expr);
        void visitCallExpr(Call expr);
        void visitGetExpr(Get expr);
        void visitSetExpr(Set expr);
    }

    public static class Null extends Expr {
        @Override
        public void accept(Visitor visitor) {
            visitor.visitNullExpr(this);
        }

        @Override
        public Type getType() {
            throw new RuntimeException("No type for null."); // TODO
        }
    }

    public static class Bool extends Expr {
        public final boolean value;

        public Bool(boolean value) {
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitBoolExpr(this);
        }

        @Override
        public Type getType() {
            return PrimitiveTypes.Bool.type;
        }
    }

    public static class UnsignedInt extends Expr {
        public final int bytes;
        public final long value;

        public UnsignedInt(int bytes, long value) {
            this.bytes = bytes;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitUnsignedIntExpr(this);
        }

        @Override
        public Type getType() {
            return switch (bytes) {
                case 1 -> PrimitiveTypes.U8.type;
                case 2 -> PrimitiveTypes.U16.type;
                case 4 -> PrimitiveTypes.U32.type;
                case 8 -> PrimitiveTypes.U64.type;
                default -> throw new RuntimeException("Unknown unsigned integer size.");
            };
        }
    }

    public static class Int extends Expr {
        public final int bytes;
        public final long value;

        public Int(int bytes, long value) {
            this.bytes = bytes;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitIntExpr(this);
        }

        @Override
        public Type getType() {
            return switch (bytes) {
                case 1 -> PrimitiveTypes.I8.type;
                case 2 -> PrimitiveTypes.I16.type;
                case 4 -> PrimitiveTypes.I32.type;
                case 8 -> PrimitiveTypes.I64.type;
                default -> throw new RuntimeException("Unknown integer size.");
            };
        }
    }

    public static class Float extends Expr {
        public final boolean is64bit;
        public final double value;

        public Float(boolean is64bit, double value) {
            this.is64bit = is64bit;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitFloatExpr(this);
        }

        @Override
        public Type getType() {
            return is64bit ? PrimitiveTypes.F64.type : PrimitiveTypes.F32.type;
        }
    }

    public static class String extends Expr {
        public final java.lang.String value;

        public String(java.lang.String value) {
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitStringExpr(this);
        }

        @Override
        public Type getType() {
            throw new RuntimeException("No type for string."); // TODO
        }
    }

    public static class Grouping extends Expr {
        public final Expr expression;

        public Grouping(Expr expression) {
            this.expression = expression;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitGroupingExpr(this);
        }

        @Override
        public Type getType() {
            return expression.getType();
        }
    }

    public static class Binary extends Expr {
        public final Expr left;
        public final Token operator;
        public final Expr right;

        public Binary(Expr left, Token operator, Expr right) {
            this.left = left;
            this.operator = operator;
            this.right = right;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitBinaryExpr(this);
        }

        @Override
        public Type getType() {
            return left.getType();
        }
    }

    public static class Cast extends Expr {
        public final Expr expr;
        public final ProtoType type;

        public Type _type;

        public Cast(Expr expr, ProtoType type) {
            this.expr = expr;
            this.type = type;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitCastExpr(this);
        }

        @Override
        public Type getType() {
            return _type;
        }
    }

    public static class Unary extends Expr {
        public final Token operator;
        public final Expr right;

        public Unary(Token operator, Expr right) {
            this.operator = operator;
            this.right = right;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitUnaryExpr(this);
        }

        @Override
        public Type getType() {
            Type type = right.getType();
            return operator.type() == TokenType.Ampersand ? type.pointer() : type;
        }
    }

    public static class UnaryPost extends Expr {
        public final Token operator;
        public final Expr left;

        public UnaryPost(Token operator, Expr left) {
            this.operator = operator;
            this.left = left;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitUnaryPostExpr(this);
        }

        @Override
        public Type getType() {
            return left.getType();
        }
    }

    public static class Logical extends Expr {
        public final Expr left;
        public final Token operator;
        public final Expr right;

        public Logical(Expr left, Token operator, Expr right) {
            this.left = left;
            this.operator = operator;
            this.right = right;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitLogicalExpr(this);
        }

        @Override
        public Type getType() {
            return left.getType();
        }
    }

    public static class Variable extends Expr {
        public final Token name;
        public Type type;

        public Variable(Token name) {
            this.name = name;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitVariableExpr(this);
        }

        @Override
        public Type getType() {
            return type;
        }
    }

    public static class Assign extends Expr {
        public final Token name;
        public final Expr value;
        public Type type;

        public Assign(Token name, Expr value) {
            this.name = name;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitAssignExpr(this);
        }

        @Override
        public Type getType() {
            return type;
        }
    }

    public static class Call extends Expr {
        public final Token token;
        public final Expr callee;
        public final List<Expr> arguments;

        public Call(Token token, Expr callee, List<Expr> arguments) {
            this.token = token;
            this.callee = callee;
            this.arguments = arguments;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitCallExpr(this);
        }

        @Override
        public Type getType() {
            return callee.getType();
        }
    }

    public static class Get extends Expr {
        public final Expr object;
        public final Token name;
        public Type type;

        public Get(Expr object, Token name) {
            this.object = object;
            this.name = name;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitGetExpr(this);
        }

        @Override
        public Type getType() {
            return type;
        }
    }

    public static class Set extends Expr {
        public final Expr object;
        public final Token name;
        public final Expr value;
        public Type type;

        public Set(Expr object, Token name, Expr value) {
            this.object = object;
            this.name = name;
            this.value = value;
        }

        @Override
        public void accept(Visitor visitor) {
            visitor.visitSetExpr(this);
        }

        @Override
        public Type getType() {
            return type;
        }
    }

    public abstract void accept(Visitor visitor);

    public abstract Type getType();
}
