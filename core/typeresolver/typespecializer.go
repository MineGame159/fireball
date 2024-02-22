package typeresolver

import (
	"fireball/core/ast"
	"fireball/core/utils"
)

type typeSpecializer struct {
	reporter utils.Reporter
}

func Specialize(reporter utils.Reporter, file *ast.File) {
	s := typeSpecializer{reporter: reporter}

	s.VisitNode(file)
}

func (t *typeSpecializer) visitResolvable(resolvable *ast.Resolvable) {
	// Specialize struct if needed
	if s, ok := ast.As[*ast.Struct](resolvable.Type); ok {
		if len(resolvable.GenericArgs) != len(s.GenericParams) {
			if resolvable.GenericArgs != nil {
				errorSlice(t.reporter, resolvable.GenericArgs, "Got '%d' generic arguments but struct takes '%d'", len(resolvable.GenericArgs), len(s.GenericParams))
			} else {
				errorNode(t.reporter, resolvable.Parts[len(resolvable.Parts)-1], "Got '%d' generic arguments but struct takes '%d'", len(resolvable.GenericArgs), len(s.GenericParams))
			}
		}

		if len(resolvable.GenericArgs) != 0 {
			resolvable.Type = s.Specialize(resolvable.GenericArgs)
		}
	} else if len(resolvable.GenericArgs) != 0 {
		errorSlice(t.reporter, resolvable.GenericArgs, "This type doesn't have any generic parameters")
	}
}

// ast.Visitor

func (t *typeSpecializer) VisitNode(node ast.Node) {
	if resolvable, ok := node.(*ast.Resolvable); ok {
		t.visitResolvable(resolvable)
	}

	node.AcceptChildren(t)
}
