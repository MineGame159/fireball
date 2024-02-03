package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
	"fireball/core/scanner"
)

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  ast.Node
	value exprValue
}

type scopes struct {
	c *codegen

	scopes    []scope
	variables []variable

	file ir.MetaID

	unitMeta *ir.CompileUnitMeta
	unitId   ir.MetaID

	scopesMeta []ir.MetaID

	declare ir.Value
}

func (s *scopes) pushFile(path string) {
	const producer = "fireball compiler 0.1.0"

	// File
	s.file = s.c.module.Meta(&ir.FileMeta{Path: path})

	s.push(s.file)

	// Unit
	s.unitMeta = &ir.CompileUnitMeta{
		File:      s.file,
		Producer:  producer,
		Emission:  ir.FullDebug,
		NameTable: ir.None,
	}

	m := s.c.module
	s.unitId = m.Meta(s.unitMeta)

	m.NamedMeta("llvm.dbg.cu", []ir.MetaID{s.unitId})

	// Identifier
	m.NamedMeta("llvm.ident", []ir.MetaID{m.Meta(&ir.GroupMeta{
		Metadata: []ir.MetaID{m.Meta(&ir.StringMeta{Value: producer})},
	})})

	// Flags
	m.NamedMeta("llvm.module.flags", []ir.MetaID{
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 7}),
			m.Meta(&ir.StringMeta{Value: "Dwarf Version"}),
			m.Meta(&ir.IntMeta{Value: 4}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 2}),
			m.Meta(&ir.StringMeta{Value: "Debug Info Version"}),
			m.Meta(&ir.IntMeta{Value: 3}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 1}),
			m.Meta(&ir.StringMeta{Value: "wchar_size"}),
			m.Meta(&ir.IntMeta{Value: 4}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 8}),
			m.Meta(&ir.StringMeta{Value: "PIC Level"}),
			m.Meta(&ir.IntMeta{Value: 2}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 7}),
			m.Meta(&ir.StringMeta{Value: "PIE Level"}),
			m.Meta(&ir.IntMeta{Value: 2}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 7}),
			m.Meta(&ir.StringMeta{Value: "uwtable"}),
			m.Meta(&ir.IntMeta{Value: 2}),
		}}),
		m.Meta(&ir.GroupMeta{Metadata: []ir.MetaID{
			m.Meta(&ir.IntMeta{Value: 7}),
			m.Meta(&ir.StringMeta{Value: "frame-pointer"}),
			m.Meta(&ir.IntMeta{Value: 2}),
		}}),
	})
}

func (s *scopes) pushBlock(node ast.Node) {
	block := &ir.LexicalBlockMeta{
		Scope: s.getMeta(),
		File:  s.file,
	}

	if node.Cst() != nil {
		block.Line = uint32(node.Cst().Range.Start.Line)
	}

	s.push(s.c.module.Meta(block))
}

func (s *scopes) push(id ir.MetaID) {
	s.scopes = append(s.scopes, scope{
		variableI:     len(s.variables),
		variableCount: 0,
	})

	s.scopesMeta = append(s.scopesMeta, id)
}

func (s *scopes) pop() {
	s.variables = s.variables[:s.get().variableI]
	s.scopes = s.scopes[:len(s.scopes)-1]

	s.scopesMeta = s.scopesMeta[:len(s.scopesMeta)-0]
}

func (s *scopes) get() *scope {
	return &s.scopes[len(s.scopes)-1]
}

func (s *scopes) getMeta() ir.MetaID {
	return s.scopesMeta[len(s.scopesMeta)-1]
}

// Variables

func (s *scopes) getVariable(name scanner.Token) *variable {
	for i := len(s.variables) - 1; i >= 0; i-- {
		if s.variables[i].name.String() == name.Lexeme {
			return &s.variables[i]
		}
	}

	return nil
}

func (s *scopes) addVariable(name ast.Node, type_ ast.Type, value ir.Value, arg uint32) *variable {
	if s.declare == nil {
		s.declare = s.c.module.Declare(
			"llvm.dbg.declare",
			&ir.FuncType{Params: []*ir.Param{
				{Typ: ir.MetaT},
				{Typ: ir.MetaT},
				{Typ: ir.MetaT},
			}})
	}

	if _, ok := ast.As[*ast.Func](type_); ok {
		type_ = &ast.Pointer{Pointee: type_}
	}

	meta := &ir.LocalVarMeta{
		Name:  name.String(),
		Type:  s.c.types.getMeta(type_),
		Arg:   arg,
		Scope: s.getMeta(),
		File:  s.file,
	}

	if name.Cst() != nil {
		meta.Line = uint32(name.Cst().Range.Start.Line)
	}

	declare := s.c.block.Add(&ir.CallInst{
		Callee: s.declare,
		Args: []ir.Value{
			value,
			s.c.module.Meta(meta),
			s.c.module.Meta(&ir.ExpressionMeta{}),
		},
	})

	s.c.setLocationMeta(declare, name)

	s.variables = append(s.variables, variable{
		name: name,
		value: exprValue{
			v:           value,
			addressable: true,
		},
	})

	s.get().variableCount++
	return &s.variables[len(s.variables)-1]
}
