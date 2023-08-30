package typeresolver

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
	"fmt"
)

type resolver struct {
	structs map[string]*types.StructType

	reporter core.Reporter
}

func Resolve(reporter core.Reporter, decls []ast.Decl) {
	r := &resolver{
		structs:  make(map[string]*types.StructType),
		reporter: reporter,
	}

	// Collect structs
	for _, decl := range decls {
		if s, ok := decl.(*ast.Struct); ok {
			fields := make([]types.Field, len(s.Fields))

			for i, field := range s.Fields {
				fields[i] = types.Field{
					Name: field.Name.Lexeme,
					Type: field.Type,
				}
			}

			type_ := &types.StructType{
				Name:   s.Name.Lexeme,
				Fields: fields,
			}

			s.Type = type_
			r.structs[s.Name.Lexeme] = type_
		}
	}

	// Resolve
	for _, decl := range decls {
		r.AcceptDecl(decl)
	}
}

// types.Visitor

func (r *resolver) VisitType(type_ *types.Type) {
	if v, ok := (*type_).(*types.UnresolvedType); ok {
		t, ok := r.structs[v.Identifier.Lexeme]

		if !ok {
			r.error("Unknown type '%s'.", v)
			*type_ = types.Primitive(types.Void)
		} else {
			*type_ = t
		}
	}

	if *type_ != nil {
		(*type_).AcceptTypes(r)
	}
}

// ast.Acceptor

func (r *resolver) AcceptDecl(decl ast.Decl) {
	decl.AcceptTypes(r)
	decl.Accept(r)
}

func (r *resolver) AcceptStmt(stmt ast.Stmt) {
	stmt.AcceptTypes(r)
	stmt.Accept(r)
}

func (r *resolver) AcceptExpr(expr ast.Expr) {
	expr.AcceptTypes(r)
	expr.Accept(r)
}

// Diagnostic

func (r *resolver) error(format string, args ...any) {
	// TODO
	r.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Message: fmt.Sprintf(format, args...),
	})
}
