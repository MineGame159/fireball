package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
	"path/filepath"
)

func getDefinition(file *workspace.File, pos core.Pos) []protocol.Location {
	for _, decl := range file.Decls {
		// Get leaf
		node := ast.GetLeaf(decl, pos)
		if node == nil {
			continue
		}

		if identifier, ok := node.(*ast.Identifier); ok {
			// ast.Identifier
			name := identifier.Identifier.Lexeme

			switch identifier.Kind {
			case ast.FunctionKind:
				type_, path := file.Project.GetFunction(name)
				if type_ == nil {
					return nil
				}

				return []protocol.Location{{
					URI:   uri.New(filepath.Join(file.Project.Path, path)),
					Range: convertRange(type_.Range()),
				}}

			case ast.EnumKind:
				type_, path := file.Project.GetType(name)
				if type_ == nil {
					return nil
				}

				return []protocol.Location{{
					URI:   uri.New(filepath.Join(file.Project.Path, path)),
					Range: convertRange(type_.Range()),
				}}

			case ast.ParameterKind:
				if function, ok := decl.(*ast.Func); ok {
					for _, param := range function.Params {
						if param.Name.Lexeme == name {
							return []protocol.Location{{
								URI:   uri.New(filepath.Join(file.Project.Path, file.Path)),
								Range: convertRange(core.TokenToRange(param.Name)),
							}}
						}
					}
				}

			case ast.VariableKind:
				resolver := variableResolver{
					name:   name,
					target: node,
				}

				resolver.Accept(decl)

				if resolver.variable != nil {
					return []protocol.Location{{
						URI:   uri.New(filepath.Join(file.Project.Path, file.Path)),
						Range: convertRange(resolver.variable.Range()),
					}}
				}
			}
		} else if member, ok := node.(*ast.Member); ok && member.Result().Kind == ast.FunctionResultKind {
			type_ := member.Value.Result().Type

			if pointer, ok := type_.(*types.PointerType); ok {
				type_ = pointer.Pointee
			}

			function, path := file.Project.GetMethod(type_, member.Name.Lexeme)
			if function == nil {
				return nil
			}

			return []protocol.Location{{
				URI:   uri.New(filepath.Join(file.Project.Path, path)),
				Range: convertRange(function.Range()),
			}}
		}
	}

	return nil
}

type scope struct {
	variableI     int
	variableCount int
}

type variableResolver struct {
	name   string
	target ast.Node

	scopes    []scope
	variables []*ast.Variable

	done     bool
	variable *ast.Variable
}

func (v *variableResolver) Accept(node ast.Node) {
	// Propagate
	if v.done {
		return
	}

	// Check target
	if node == v.target {
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
	} else if variable, ok := node.(*ast.Variable); ok {
		v.scopes[len(v.scopes)-1].variableCount++
		v.variables = append(v.variables, variable)
	}

	// Propagate or visit children
	if v.done {
		return
	}

	node.AcceptChildren(v)

	if pop {
		v.popScope()
	}
}

func (v *variableResolver) checkScope() {
	for i := len(v.variables) - 1; i >= 0; i-- {
		variable := v.variables[i]

		if variable.Name.Lexeme == v.name {
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

func (v *variableResolver) AcceptDecl(decl ast.Decl) {
	v.Accept(decl)
}

func (v *variableResolver) AcceptStmt(stmt ast.Stmt) {
	v.Accept(stmt)
}

func (v *variableResolver) AcceptExpr(expr ast.Expr) {
	v.Accept(expr)
}
