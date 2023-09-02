package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"io"
	"log"
	"strconv"
)

type constant struct {
	identifier string
	original   string

	data   []uint8
	length int
}

type typeValuePair struct {
	type_ types.Type
	val   value
}

type typeDbgPair struct {
	type_ types.Type
	name  string
}

type codegen struct {
	debug    debug
	dbgTypes []typeDbgPair

	globals values
	blocks  values
	locals  values

	constants []constant
	types     []typeValuePair
	functions []function

	scopes    []scope
	variables []variable

	block string

	loopStart string
	loopEnd   string

	exprValue value

	writer io.Writer
	depth  int
}

type function struct {
	name scanner.Token
	val  value
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name scanner.Token
	val  value
}

func Emit(filename string, decls []ast.Decl, writer io.Writer) {
	// Init codegen
	c := &codegen{
		globals: values{char: "@"},
		blocks:  values{char: "bb_"},
		locals:  values{char: "%_"},

		types: make([]typeValuePair, 0, 64),

		writer: writer,
		depth:  0,
	}

	// File metadata
	c.writeFmt("source_filename = \"%s\"\n", filename)
	c.writeRaw("\n")

	file := c.debug.pushScope(c.debug.file(filename))
	c.debug.compileUnit(file)

	// Emit types
	for _, decl := range decls {
		if v, ok := decl.(*ast.Struct); ok {
			val := c.globals.constant("%struct."+v.Name.Lexeme, v.Type)

			c.types = append(c.types, typeValuePair{
				type_: v.Type,
				val:   val,
			})

			c.writeRaw(val.identifier)
			c.writeRaw(" = type { ")

			for i, field := range v.Fields {
				if i > 0 {
					c.writeRaw(", ")
				}

				c.writeRaw(c.getType(field.Type).identifier)
			}

			c.writeRaw(" }\n")
		}
	}

	c.writeRaw("\n")

	// Find all functions
	for _, decl := range decls {
		if f, ok := decl.(*ast.Func); ok {
			params := make([]types.Type, len(f.Params))

			for i, param := range f.Params {
				params[i] = param.Type
			}

			c.addFunction(f.Name, &types.FunctionType{
				Params:   params,
				Variadic: f.Variadic,
				Returns:  f.Returns,
			})
		}
	}

	// Emit
	for _, decl := range decls {
		c.acceptDecl(decl)
	}

	// Emit constants
	for _, co := range c.constants {
		c.writeFmt("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", co.identifier, co.length+1, co.data)
	}

	if len(c.constants) > 0 {
		c.writeRaw("\n")
	}

	// Emit debug metadata
	c.debug.popScope()
	c.debug.write(c)
}

// IR

func (c *codegen) load(val value, type_ types.Type) value {
	if val.needsLoading {
		res := c.locals.unnamed(val.type_)
		c.writeFmt("%s = load %s, ptr %s\n", res, c.getType(type_), val)

		return res
	}

	return val
}

func (c *codegen) toPtrOrLoad(val value, type_ types.Type) value {
	if val.needsLoading {
		return value{
			identifier: val.identifier,
			type_:      &types.PointerType{Pointee: val.type_},
		}
	}

	return c.load(val, type_)
}

func (c *codegen) writeBlock(block string) {
	c.writeRaw(block + ":\n")
	c.block = block
}

// Constants

func (c *codegen) getConstant(data string) string {
	// GetLeaf
	for _, co := range c.constants {
		if co.original == data {
			return co.identifier
		}
	}

	// Create
	identifier := c.globals.unnamedRaw()

	c.constants = append(c.constants, constant{
		identifier: identifier,
		original:   data,
	})

	c.constants[len(c.constants)-1].escape()

	return identifier
}

func (c *constant) escape() {
	data := make([]uint8, 0, len(c.original)+1)

	for i := 0; i < len(c.original); i++ {
		char := c.original[i]

		if char == '\\' {
			i++

			switch c.original[i] {
			case '0':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, '0')

			case 'n':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, 'A')

			case 'r':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, 'D')

			case 't':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, '9')
			}

			c.length++
		} else {
			data = append(data, char)
			c.length++
		}
	}

	c.data = data
}

// Types

func (c *codegen) getType(type_ types.Type) value {
	// TODO: Linear search for types, very bad

	// Try cache
	for _, pair := range c.types {
		if pair.type_.Equals(type_) {
			return pair.val
		}
	}

	// Array
	if v, ok := type_.(*types.ArrayType); ok {
		val := c.globals.constant(fmt.Sprintf("[%d x %s]", v.Count, c.getType(v.Base)), type_)

		c.types = append(c.types, typeValuePair{
			type_: type_,
			val:   val,
		})

		return val
	}

	// Pointer
	if _, ok := type_.(*types.PointerType); ok {
		val := c.globals.constant("ptr", type_)

		c.types = append(c.types, typeValuePair{
			type_: type_,
			val:   val,
		})

		return val
	}

	// Primitive
	if v, ok := type_.(*types.PrimitiveType); ok {
		name := ""

		switch v.Kind {
		case types.Bool:
			name = "i1"

		case types.U8:
			name = "i8"
		case types.U16:
			name = "i16"
		case types.U32:
			name = "i32"
		case types.U64:
			name = "i64"

		case types.F32:
			name = "float"
		case types.F64:
			name = "double"

		default:
			name = v.String()
		}

		val := c.globals.constant(name, type_)

		c.types = append(c.types, typeValuePair{
			type_: type_,
			val:   val,
		})

		return val
	}

	// Error
	log.Fatalln("codegen.getType() - Invalid type")
	return value{}
}

// Functions

func (c *codegen) getFunction(name scanner.Token) *function {
	for i := 0; i < len(c.functions); i++ {
		if c.functions[i].name.Lexeme == name.Lexeme {
			return &c.functions[i]
		}
	}

	return nil
}

func (c *codegen) addFunction(name scanner.Token, type_ types.Type) {
	c.functions = append(c.functions, function{
		name: name,
		val:  c.globals.named(name.Lexeme, type_),
	})
}

// Scope / Variables

func (c *codegen) getVariable(name scanner.Token) *variable {
	for i := len(c.variables) - 1; i >= 0; i-- {
		if c.variables[i].name.Lexeme == name.Lexeme {
			return &c.variables[i]
		}
	}

	return nil
}

func (c *codegen) addVariable(name scanner.Token, val value) *variable {
	val.needsLoading = true

	c.variables = append(c.variables, variable{
		name: name,
		val:  val,
	})

	c.peekScope().variableCount++
	return &c.variables[len(c.variables)-1]
}

func (c *codegen) pushScope() {
	c.scopes = append(c.scopes, scope{
		variableI:     len(c.variables),
		variableCount: 0,
	})
}

func (c *codegen) popScope() {
	c.variables = c.variables[:c.peekScope().variableI]
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *codegen) peekScope() *scope {
	return &c.scopes[len(c.scopes)-1]
}

// Debug

func (c *codegen) getDbgType(type_ types.Type) string {
	// TODO: Linear search for types, very bad

	// Try cache
	for _, pair := range c.dbgTypes {
		if pair.type_.Equals(type_) {
			return pair.name
		}
	}

	// Struct
	if v, ok := type_.(*types.StructType); ok {
		members := make([]string, len(v.Fields))
		offset := 0

		for i, field := range v.Fields {
			size := field.Type.Size() * 8
			members[i] = c.debug.derivedType(MemberDTag, field.Name, c.getDbgType(field.Type), size, offset)
			offset += size
		}

		name := c.debug.compositeType(StructureTypeCTag, "", v.Range().Start.Line, v.Size()*8, c.debug.tuple(members))
		c.dbgTypes = append(c.dbgTypes, typeDbgPair{
			type_: type_,
			name:  name,
		})

		return name
	}

	// Array
	if v, ok := type_.(*types.ArrayType); ok {
		subRange := c.debug.subrange(int(v.Count), 0)
		name := c.debug.compositeType(ArrayTypeCTag, c.getDbgType(v.Base), v.Range().Start.Line, v.Size()*8, "!{"+subRange+"}")

		c.dbgTypes = append(c.dbgTypes, typeDbgPair{
			type_: type_,
			name:  name,
		})

		return name
	}

	// Pointer
	if v, ok := type_.(*types.PointerType); ok {
		name := c.debug.derivedType(PointerTypeDTag, "", c.getDbgType(v.Pointee), v.Size()*8, 0)
		c.dbgTypes = append(c.dbgTypes, typeDbgPair{
			type_: type_,
			name:  name,
		})

		return name
	}

	// Primitive
	if v, ok := type_.(*types.PrimitiveType); ok {
		var size int
		var encoding Encoding

		switch v.Kind {
		case types.Void:
			c.dbgTypes = append(c.dbgTypes, typeDbgPair{
				type_: type_,
				name:  "null",
			})

			return "null"

		case types.Bool:
			size = 1
			encoding = BooleanEncoding

		case types.U8:
			size = 8
			encoding = UnsignedEncoding
		case types.U16:
			size = 16
			encoding = UnsignedEncoding
		case types.U32:
			size = 32
			encoding = UnsignedEncoding
		case types.U64:
			size = 64
			encoding = UnsignedEncoding

		case types.I8:
			size = 8
			encoding = SignedEncoding
		case types.I16:
			size = 16
			encoding = SignedEncoding
		case types.I32:
			size = 32
			encoding = SignedEncoding
		case types.I64:
			size = 64
			encoding = SignedEncoding

		case types.F32:
			size = 32
			encoding = FloatEncoding
		case types.F64:
			size = 64
			encoding = FloatEncoding

		default:
			log.Fatalln("codegen.getDbgType() - Invalid primitive type kind")
		}

		name := c.debug.basicType(v.String(), size, encoding)
		c.dbgTypes = append(c.dbgTypes, typeDbgPair{
			type_: type_,
			name:  name,
		})

		return name
	}

	// Error
	log.Fatalln("codegen.getDbgType() - Invalid type")
	return ""
}

// Accept

func (c *codegen) acceptDecl(decl ast.Decl) {
	if decl != nil {
		decl.Accept(c)
	}
}

func (c *codegen) acceptStmt(stmt ast.Stmt) {
	if stmt != nil {
		stmt.Accept(c)
	}
}

func (c *codegen) acceptExpr(expr ast.Expr) value {
	if expr != nil {
		expr.Accept(c)
		return c.exprValue
	}

	return value{}
}

// Write

func (c *codegen) writeFmt(format string, args ...any) {
	c.writeStr(fmt.Sprintf(format, args...))
}

func (c *codegen) writeStr(str string) {
	if endsWith(str, '}') {
		c.depth--
	}

	for i := 0; i < c.depth; i++ {
		_, _ = c.writer.Write([]byte("\t"))
	}

	c.writeRaw(str)

	if endsWith(str, '{') {
		c.depth++
	}
}

func (c *codegen) writeRaw(str string) {
	_, _ = c.writer.Write([]byte(str))
}

func endsWith(str string, char uint8) bool {
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] != '\n' {
			return str[i] == char
		}
	}

	return false
}

// Utils

func ternary[T any](cond bool, true T, false T) T {
	if cond {
		return true
	}

	return false
}

// Value

type value struct {
	identifier   string
	type_        types.Type
	needsLoading bool
}

func (v value) String() string {
	return v.identifier
}

type values struct {
	char         string
	unnamedCount int
}

func (v *values) reset() {
	v.unnamedCount = 0
}

func (v *values) named(identifier string, type_ types.Type) value {
	return value{
		identifier: v.char + identifier,
		type_:      type_,
	}
}

func (v *values) unnamed(type_ types.Type) value {
	return value{
		identifier: v.unnamedRaw(),
		type_:      type_,
	}
}

func (v *values) unnamedRaw() string {
	v.unnamedCount++
	return v.char + strconv.Itoa(v.unnamedCount-1)
}

func (v *values) constant(constant string, type_ types.Type) value {
	return value{
		identifier: constant,
		type_:      type_,
	}
}
