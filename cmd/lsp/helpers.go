package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
)

func buildResolverStack(resolver *ast.Resolver, node ast.Node) {
	if !ast.IsNil(node.Parent()) {
		buildResolverStack(resolver, node.Parent())
	}

	switch node := node.(type) {
	case *ast.Struct:
		*resolver = ast.NewGenericResolver(*resolver, node.GenericParams)

	case *ast.Impl:
		if s, ok := ast.As[*ast.Struct](node.Type); ok && len(s.GenericParams) > 0 {
			*resolver = ast.NewGenericResolver(*resolver, s.GenericParams)
		}

	case *ast.Func:
		*resolver = ast.NewGenericResolver(*resolver, node.GenericParams)
	}
}

func printType(type_ ast.Type) string {
	return ast.PrintTypeOptions(type_, ast.TypePrintOptions{ParamNames: true})
}

func asThroughPointer[T ast.Type](type_ ast.Type) (T, bool) {
	if pointer, ok := ast.As[*ast.Pointer](type_); ok {
		return ast.As[T](pointer.Pointee)
	}

	return ast.As[T](type_)
}

func getNodeUnderPos(node ast.Node, pos core.Pos) ast.Node {
	leaf := ast.GetLeaf(node, pos)
	if leaf != nil {
		return leaf
	}

	return ast.Get(node, pos)
}

func isBetween(pos core.Pos, node ast.Node, start, end scanner.TokenKind) bool {
	if node.Cst() == nil {
		return false
	}

	left := node.Cst().Get(start)
	if left == nil {
		return false
	}

	right := node.Cst().Get(end)
	if right == nil {
		return pos.IsAfter(left.Range.Start)
	}

	return core.Range{Start: left.Range.End, End: right.Range.Start}.Contains(pos)
}

func nodeCst(node ast.Node) *cst.Node {
	if ast.IsNil(node) {
		return nil
	}

	return node.Cst()
}

// Variable resolver

type scope struct {
	variableI     int
	variableCount int
}

type variableResolver struct {
	target             core.Pos
	targetVariableName ast.Node

	scopes    []scope
	variables []*ast.Var

	done     bool
	variable *ast.Var
}

func (v *variableResolver) VisitNode(node ast.Node) {
	// Propagate
	if v.done {
		return
	}

	// Check target
	if node == v.targetVariableName || (node.Cst() != nil && node.Cst().Range.Start.IsAfter(v.target)) {
		v.done = true
		v.checkScope()
		return
	}

	// Check node
	pop := false

	if _, ok := node.(*ast.Func); ok {
		v.pushScope()
		pop = true
	} else if _, ok := node.(*ast.For); ok {
		v.pushScope()
		pop = true
	} else if _, ok := node.(*ast.Block); ok {
		v.pushScope()
		pop = true
	} else if variable, ok := node.(*ast.Var); ok {
		v.scopes[len(v.scopes)-1].variableCount++
		v.variables = append(v.variables, variable)
	}

	// Propagate or visit children
	if v.done {
		return
	}

	node.AcceptChildren(v)

	if pop && !v.done {
		v.popScope()
	}
}

func (v *variableResolver) checkScope() {
	for i := len(v.variables) - 1; i >= 0; i-- {
		variable := v.variables[i]

		if v.targetVariableName != nil && v.variable == nil && variable.Name.String() == v.targetVariableName.String() {
			v.variable = variable
			break
		}
	}
}

func (v *variableResolver) pushScope() {
	v.scopes = append(v.scopes, scope{
		variableI:     len(v.variables),
		variableCount: 0,
	})
}

func (v *variableResolver) popScope() {
	v.variables = v.variables[:v.scopes[len(v.scopes)-1].variableI]
	v.scopes = v.scopes[:len(v.scopes)-1]
}
