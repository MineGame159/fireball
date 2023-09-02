package checker

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/types"
)

func (c *checker) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(c)

	// Check fields
	fields := core.NewSet[string]()

	for _, field := range decl.Fields {
		// Check name collision
		if !fields.Add(field.Name.Lexeme) {
			c.errorToken(field.Name, "Field with the name '%s' already exists.", field.Name)
		}

		// Check void type
		if types.IsPrimitive(field.Type, types.Void) {
			c.errorToken(field.Name, "Field cannot be of type 'void'.")
		}
	}
}

func (c *checker) VisitEnum(decl *ast.Enum) {
	decl.AcceptChildren(c)

	// Check type
	if decl.Type != nil {
		if v, ok := decl.Type.(*types.PrimitiveType); !ok || !types.IsInteger(v.Kind) {
			c.errorRange(decl.Type.Range(), "Invalid type '%s', can only be a signed or unsigned integer.", decl.Type)
		} else {
			// Check if all cases fit inside the type
			min_, max_ := types.GetRangeTrunc(v.Kind)

			for _, case_ := range decl.Cases {
				if int64(case_.Value) < min_ || int64(case_.Value) > max_ {
					c.errorToken(case_.Name, "Value '%d' does not fit inside the range of '%s'.", case_.Value, decl.Type)
				}
			}
		}
	}
}

func (c *checker) VisitFunc(decl *ast.Func) {
	// Push scope
	c.function = decl
	c.pushScope()

	// Params
	for _, param := range decl.Params {
		c.addVariable(param.Name, param.Type).param = true
	}

	// Body
	for _, stmt := range decl.Body {
		c.AcceptStmt(stmt)
	}

	// Pop scope
	c.popScope()
	c.function = nil

	// Check name collision
	if !c.functions.Add(decl.Name.Lexeme) {
		c.errorToken(decl.Name, "Function with the name '%s' already exists.", decl.Name)
	}

	// Check parameter void type
	for _, param := range decl.Params {
		if types.IsPrimitive(param.Type, types.Void) {
			c.errorToken(param.Name, "Parameter cannot be of type 'void'.")
		}
	}

	// Check last return
	if !decl.Extern && !types.IsPrimitive(decl.Returns, types.Void) {
		valid := len(decl.Body) > 0

		if valid {
			if _, ok := decl.Body[len(decl.Body)-1].(*ast.Return); !ok {
				valid = false
			}
		}

		if !valid {
			c.errorToken(decl.Name, "Function needs to return a '%s' value.", decl.Returns)
		}
	}
}
