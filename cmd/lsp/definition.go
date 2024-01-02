package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/workspace"
	"github.com/MineGame159/protocol"
	"go.lsp.dev/uri"
	"path/filepath"
)

func getDefinition(file *workspace.File, pos core.Pos) []protocol.Location {
	for _, decl := range file.Ast.Decls {
		// Get leaf
		node := ast.GetLeaf(decl, pos)
		if node == nil {
			continue
		}

		if identifier, ok := node.(*ast.Identifier); ok {
			// ast.Identifier
			name := identifier.String()

			switch identifier.Kind {
			case ast.FunctionKind:
				type_, path := file.Project.GetFunction(name)
				if type_ == nil {
					return nil
				}

				if type_.Cst() != nil {
					return []protocol.Location{{
						URI:   uri.New(filepath.Join(file.Project.Path, path)),
						Range: convertRange(type_.Cst().Range),
					}}
				}

			case ast.StructKind, ast.EnumKind:
				type_, path := file.Project.GetType(name)
				if type_ == nil {
					return nil
				}

				if type_.Cst() != nil {
					return []protocol.Location{{
						URI:   uri.New(filepath.Join(file.Project.Path, path)),
						Range: convertRange(type_.Cst().Range),
					}}
				}

			case ast.ParameterKind:
				if function, ok := decl.(*ast.Func); ok {
					for _, param := range function.Params {
						if param.Name.String() == name && param.Name.Cst() != nil {
							return []protocol.Location{{
								URI:   uri.New(filepath.Join(file.Project.Path, file.Path)),
								Range: convertRange(param.Name.Cst().Range),
							}}
						}
					}
				}

			case ast.VariableKind:
				resolver := variableResolver{
					name:   name,
					target: node,
				}

				resolver.VisitNode(decl)

				if resolver.variable != nil && resolver.variable.Cst() != nil {
					return []protocol.Location{{
						URI:   uri.New(filepath.Join(file.Project.Path, file.Path)),
						Range: convertRange(resolver.variable.Cst().Range),
					}}
				}
			}
		} else if member, ok := node.(*ast.Member); ok {
			type_ := member.Value.Result().Type

			if pointer, ok := ast.As[*ast.Pointer](type_); ok {
				type_ = pointer.Pointee
			}

			switch member.Result().Kind {
			case ast.ValueResultKind:
				switch type_.(type) {
				case *ast.Struct:
					if t, path := file.Project.GetType(type_.String()); t != nil {
						_, field := t.(*ast.Struct).GetField(member.Name.String())

						if field != nil && field.Name.Cst() != nil {
							return []protocol.Location{{
								URI:   uri.New(filepath.Join(file.Project.Path, path)),
								Range: convertRange(field.Name.Cst().Range),
							}}
						}
					}

				case *ast.Enum:
					if t, path := file.Project.GetType(type_.String()); t != nil {
						case_ := t.(*ast.Enum).GetCase(member.Name.String())

						if case_ != nil && case_.Name.Cst() != nil {
							return []protocol.Location{{
								URI:   uri.New(filepath.Join(file.Project.Path, path)),
								Range: convertRange(case_.Name.Cst().Range),
							}}
						}
					}
				}

			case ast.FunctionResultKind:
				function, path := file.Project.GetMethod(type_, member.Name.String(), false)

				if function == nil {
					function, path = file.Project.GetMethod(type_, member.Name.String(), true)

					if function == nil {
						return nil
					}
				}

				if function.Cst() != nil {
					return []protocol.Location{{
						URI:   uri.New(filepath.Join(file.Project.Path, path)),
						Range: convertRange(function.Cst().Range),
					}}
				}
			}
		} else if initializer, ok := node.(*ast.StructInitializer); ok {
			if struct_, ok := initializer.Type.(*ast.Struct); ok {
				if t, path := file.Project.GetType(struct_.String()); t != nil {
					for _, field := range initializer.Fields {
						if field.Name.Cst() != nil && field.Name.Cst().Range.Contains(pos) {
							_, field := t.(*ast.Struct).GetField(field.Name.String())

							if field != nil && field.Name.Cst() != nil {
								return []protocol.Location{{
									URI:   uri.New(filepath.Join(file.Project.Path, path)),
									Range: convertRange(field.Name.Cst().Range),
								}}
							}
						}
					}
				}
			}
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
	} else if variable, ok := node.(*ast.Var); ok {
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

		if variable.Name.String() == v.name {
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
