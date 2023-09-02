package checker

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"math"
)

type checker struct {
	scopes    []scope
	variables []variable

	types   map[string]types.Type
	structs map[string]*types.StructType
	enums   map[string]*types.EnumType

	functions core.Set[string]
	function  *ast.Func

	loopDepth int

	reporter core.Reporter
	decls    []ast.Decl
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  scanner.Token
	type_ types.Type

	param bool
	used  bool
}

func Check(reporter core.Reporter, decls []ast.Decl) {
	c := &checker{
		types:   make(map[string]types.Type),
		structs: make(map[string]*types.StructType),
		enums:   make(map[string]*types.EnumType),

		functions: core.NewSet[string](),
		reporter:  reporter,
		decls:     decls,
	}

	// Collect types
	for _, decl := range decls {
		if s, ok := decl.(*ast.Struct); ok {
			// Create struct type
			fields := make([]types.Field, len(s.Fields))

			for i, field := range s.Fields {
				fields[i] = types.Field{
					Name: field.Name.Lexeme,
					Type: field.Type,
				}
			}

			type_ := types.Struct(s.Name.Lexeme, fields, core.Range{})
			s.Type = type_

			// Save in map and check name collision
			if _, ok := c.types[s.Name.Lexeme]; ok {
				c.errorToken(s.Name, "Type with the name '%s' already exists.", s.Name)
			}

			c.types[s.Name.Lexeme] = type_
			c.structs[s.Name.Lexeme] = type_
		} else if s, ok := decl.(*ast.Enum); ok {
			// Select smallest possible enum type
			if s.Type == nil {
				minValue := math.MaxInt
				maxValue := math.MinInt

				for _, case_ := range s.Cases {
					minValue = min(minValue, case_.Value)
					maxValue = max(maxValue, case_.Value)
				}

				var kind types.PrimitiveKind

				if minValue >= 0 {
					// Unsigned
					if maxValue <= math.MaxUint8 {
						kind = types.U8
					} else if maxValue <= math.MaxUint16 {
						kind = types.U16
					} else if maxValue <= math.MaxUint32 {
						kind = types.U32
					} else {
						kind = types.U64
					}
				} else {
					// Signed
					if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
						kind = types.I8
					} else if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
						kind = types.I16
					} else if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
						kind = types.I32
					} else {
						kind = types.I64
					}
				}

				s.Type = types.Primitive(kind, core.Range{})
			}

			// Create type
			cases := make([]types.EnumCase, len(s.Cases))

			for i, case_ := range s.Cases {
				cases[i] = types.EnumCase{
					Name:  case_.Name.Lexeme,
					Value: case_.Value,
				}
			}

			type_ := types.Enum(s.Name.Lexeme, s.Type, cases, core.Range{})

			// Save in map and check name collision
			if _, ok := c.types[s.Name.Lexeme]; ok {
				c.errorToken(s.Name, "Type with the name '%s' already exists.", s.Name)
			}

			c.types[s.Name.Lexeme] = type_
			c.enums[s.Name.Lexeme] = type_
		}
	}

	for _, decl := range decls {
		c.AcceptDecl(decl)
	}
}

func (c *checker) getFunction(name scanner.Token) *ast.Func {
	for _, decl := range c.decls {
		if function, ok := decl.(*ast.Func); ok {
			if function.Name.Lexeme == name.Lexeme {
				return function
			}
		}
	}

	return nil
}

// Scope / Variables

func (c *checker) getVariable(name scanner.Token) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.Lexeme == name.Lexeme {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *checker) addVariable(name scanner.Token, type_ types.Type) *variable {
	c.variables = append(c.variables, variable{
		name:  name,
		type_: type_,
	})

	c.peekScope().variableCount++
	return &c.variables[len(c.variables)-1]
}

func (c *checker) pushScope() {
	c.scopes = append(c.scopes, scope{
		variableI:     len(c.variables),
		variableCount: 0,
	})
}

func (c *checker) popScope() {
	// Check unused variables
	for i := len(c.variables) - 1; i >= c.peekScope().variableI; i-- {
		v := c.variables[i]

		if !v.used && v.name.Lexeme[0] != '_' && (!c.function.Extern || !v.param) {
			c.warningToken(v.name, "Unused variable '%s'. Prefix with '_' to ignore.", v.name)
		}
	}

	// Pop scope
	c.variables = c.variables[:c.peekScope().variableI]
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *checker) peekScope() *scope {
	return &c.scopes[len(c.scopes)-1]
}

// types.PtrVisitor

func (c *checker) VisitType(type_ *types.Type) {
	if v, ok := (*type_).(*types.UnresolvedType); ok {
		if t, ok := c.structs[v.Identifier.Lexeme]; ok {
			*type_ = types.Struct(t.Name, t.Fields, v.Range())
		} else if t, ok := c.enums[v.Identifier.Lexeme]; ok {
			*type_ = types.Enum(t.Name, t.Type, t.Cases, v.Range())
		} else {
			c.errorRange(v.Range(), "Unknown type '%s'.", v)
			*type_ = types.Primitive(types.Void, v.Range())
		}
	}

	if *type_ != nil {
		(*type_).AcceptChildrenPtr(c)
	}
}

// ast.Acceptor

func (c *checker) AcceptDecl(decl ast.Decl) {
	decl.Accept(c)
	decl.AcceptTypesPtr(c)
}

func (c *checker) AcceptStmt(stmt ast.Stmt) {
	stmt.Accept(c)
	stmt.AcceptTypesPtr(c)
}

func (c *checker) AcceptExpr(expr ast.Expr) {
	expr.Accept(c)
	expr.AcceptTypesPtr(c)
}

// Diagnostics

func (c *checker) errorRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) errorToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningRange(range_ core.Range, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   range_,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) warningToken(token scanner.Token, format string, args ...any) {
	c.reporter.Report(core.Diagnostic{
		Kind:    core.WarningKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}
