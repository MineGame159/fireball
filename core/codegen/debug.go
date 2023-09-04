package codegen

import (
	"fireball/core/scanner"
	"fmt"
	"path/filepath"
	"strings"
)

type debug struct {
	count    int
	metadata []string

	fileName        string
	compileUnitName string

	scope []string
}

func (d *debug) write(c *codegen) {
	// Debug intrinsics
	c.writeRaw("declare void @llvm.dbg.declare(metadata, metadata, metadata)\n")
	c.writeRaw("\n")

	// Module metadata
	dwarfVersion := d.node("!{i32 7, !\"Dwarf Version\", i32 4}")
	dwarfInfoVersion := d.node("!{i32 2, !\"Debug Info Version\", i32 3}")
	wcharSize := d.node("!{i32 1, !\"wchar_size\", i32 4}")
	picLevel := d.node("!{i32 8, !\"PIC Level\", i32 2}")
	pieLevel := d.node("!{i32 7, !\"PIE Level\", i32 2}")
	uwTable := d.node("!{i32 7, !\"uwtable\", i32 2}")
	framePointer := d.node("!{i32 7, !\"frame-pointer\", i32 2}")

	c.writeFmt("!llvm.dbg.cu = !{%s}\n", d.compileUnitName)
	c.writeFmt("!llvm.module.flags = !{%s, %s, %s, %s, %s, %s, %s}\n", dwarfVersion, dwarfInfoVersion, wcharSize, picLevel, pieLevel, uwTable, framePointer)
	c.writeFmt("!llvm.ident = !{%s}\n", d.node("!{!\"fireball version 0.1.0\"}"))
	c.writeRaw("\n")

	// Metadata
	for _, metadata := range d.metadata {
		c.writeRaw(metadata)
		c.writeRaw("\n")
	}
}

// Metadata

func (d *debug) location(token scanner.Token) string {
	return d.node("!DILocation(scope: %s, line: %d, column: %d)", d.getScope(), token.Line, token.Column)
}

func (d *debug) localVariable(name string, type_ string, arg int, line int) string {
	return d.node("!DILocalVariable(name: \"%s\", type: %s, arg: %d, file: %s, scope: %s, line: %d)", name, type_, arg, d.fileName, d.getScope(), line)
}

func (d *debug) lexicalBlock(token scanner.Token) string {
	return d.node("distinct !DILexicalBlock(scope: %s, file: %s, line: %d, column: %d)", d.getScope(), d.fileName, token.Line, token.Column)
}

func (d *debug) subprogram(name string, type_ string, line int) string {
	return d.node("distinct !DISubprogram(name: \"%s\", type: %s, file: %s, scope: %s, line: %d, scopeLine: %d, spFlags: DISPFlagDefinition, unit: %s)", name, type_, d.fileName, d.getScope(), line, line, d.compileUnitName)
}

func (d *debug) compileUnit(file string) string {
	name := d.node("distinct !DICompileUnit(language: DW_LANG_C, file: %s, producer: \"fireball version 0.1.0\", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug, splitDebugInlining: false, nameTableKind: None)", file)

	d.compileUnitName = name
	return name
}

func (d *debug) file(path string) string {
	name := d.node("!DIFile(filename: \"%s\", directory: \"%s\")", filepath.Base(path), filepath.Dir(path))

	d.fileName = name
	return name
}

func (d *debug) tuple(nodes []string) string {
	str := strings.Builder{}
	str.WriteString("!{")

	for i, node := range nodes {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(node)
	}

	str.WriteRune('}')
	return d.node(str.String())
}

// Types

type Encoding string

const (
	AddressEncoding      Encoding = "DW_ATE_encoding"
	BooleanEncoding      Encoding = "DW_ATE_boolean"
	FloatEncoding        Encoding = "DW_ATE_float"
	SignedEncoding       Encoding = "DW_ATE_signed"
	SignedCharEncoding   Encoding = "DW_ATE_signed_char"
	UnsignedEncoding     Encoding = "DW_ATE_unsigned"
	UnsignedCharEncoding Encoding = "DW_ATE_unsigned_char"
)

type DTag string

const (
	MemberDTag          DTag = "DW_TAG_member"
	PointerTypeDTag     DTag = "DW_TAG_pointer_type"
	ReferenceTypeDTag   DTag = "DW_TAG_reference_type"
	TypedefDTag         DTag = "DW_TAG_typedef"
	InheritanceDTag     DTag = "DW_TAG_inheritance"
	PtrToMemberTypeDTag DTag = "DW_TAG_ptr_to_member_type"
	ConstTypeDTag       DTag = "DW_TAG_const_tag"
	FriendDTag          DTag = "DW_TAG_friend"
	VolatileTypeDTag    DTag = "DW_TAG_volatile_type"
)

type CTag string

const (
	ArrayTypeCTag       CTag = "DW_TAG_array_type"
	ClassTypeCTag       CTag = "DW_TAG_class_type"
	EnumerationTypeCTag CTag = "DW_TAG_enumeration_type"
	StructureTypeCTag   CTag = "DW_TAG_structure_type"
	UnionTypeCTag       CTag = "DW_TAG_union_type"
)

func (d *debug) basicType(name string, size int, encoding Encoding) string {
	return d.node("!DIBasicType(name: \"%s\", size: %d, encoding: %s)", name, size, encoding)
}

func (d *debug) subroutineType(types string) string {
	return d.node("!DISubroutineType(types: %s)", types)
}

func (d *debug) derivedType(tag DTag, name string, baseType string, size int, offset int) string {
	if name == "" {
		if offset == 0 {
			return d.node("!DIDerivedType(tag: %s, baseType: %s, size: %d)", tag, baseType, size)
		}

		return d.node("!DIDerivedType(tag: %s, baseType: %s, size: %d, offset: %d)", tag, baseType, size, offset)
	}

	if offset == 0 {
		return d.node("!DIDerivedType(tag: %s, name: \"%s\", baseType: %s, size: %d)", tag, name, baseType, size)
	}

	return d.node("!DIDerivedType(tag: %s, name: \"%s\", baseType: %s, size: %d, offset: %d)", tag, name, baseType, size, offset)
}

func (d *debug) compositeType(tag CTag, baseType string, line int, size int, elements string) string {
	if baseType == "" {
		return d.node("!DICompositeType(tag: %s, file: %s, line: %d, size: %d, elements: %s)", tag, d.fileName, line, size, elements)
	}

	return d.node("!DICompositeType(tag: %s, baseType: %s, file: %s, line: %d, size: %d, elements: %s)", tag, baseType, d.fileName, line, size, elements)
}

func (d *debug) subrange(count, lowerBound int) string {
	return d.node("!DISubrange(count: %d, lowerBound: %d)", count, lowerBound)
}

// Scope

func (d *debug) pushScope(name string) string {
	d.scope = append(d.scope, name)
	return name
}

func (d *debug) popScope() {
	d.scope = d.scope[:len(d.scope)-1]
}

func (d *debug) getScope() string {
	return d.scope[len(d.scope)-1]
}

// Utils

func (d *debug) node(format string, args ...any) string {
	name := d.name()
	d.metadata = append(d.metadata, name+" = "+fmt.Sprintf(format, args...))

	return name
}

func (d *debug) name() string {
	d.count++
	return fmt.Sprintf("!%d", d.count-1)
}
