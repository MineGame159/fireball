package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
)

func parentWantsFunction(node ast.Node) bool {
	if call, ok := node.Parent().(*ast.Call); ok {
		return call.Callee == node
	}

	if unary, ok := node.Parent().(*ast.Unary); ok {
		return unary.Prefix && unary.Operator.Token().Kind == scanner.FuncPtr
	}

	return false
}

func asThroughPointer[T ast.Type](type_ ast.Type) (T, bool) {
	if pointer, ok := ast.As[*ast.Pointer](type_); ok {
		return ast.As[T](pointer.Pointee)
	}

	return ast.As[T](type_)
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
