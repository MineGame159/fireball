package typeresolver

import (
	"fireball/core/ast"
)

func (r *resolver) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(r)
}

func (r *resolver) VisitFunc(decl *ast.Func) {
	decl.AcceptChildren(r)
}
