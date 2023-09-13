package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
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

type typeDbgPair struct {
	type_ types.Type
	name  string
}

type importedFunc struct {
	function *ast.Func
	value    exprValue
}

type codegen struct {
	path     string
	resolver utils.Resolver

	debug    debug
	dbgTypes []typeDbgPair

	globals values
	blocks  values
	locals  values

	constants         []constant
	types             []typeValuePair
	importedFunctions []importedFunc

	scopes    []scope
	variables []variable

	block string

	loopStart string
	loopEnd   string

	exprResult exprValue

	writer io.Writer
	depth  int
}

type scope struct {
	variableI     int
	variableCount int
}

type variable struct {
	name  scanner.Token
	value exprValue
}

func Emit(path string, resolver utils.Resolver, decls []ast.Decl, writer io.Writer) {
	// Init codegen
	c := &codegen{
		path:     path,
		resolver: resolver,

		globals: values{char: "@"},
		blocks:  values{char: "bb_"},
		locals:  values{char: "%_"},

		writer: writer,
		depth:  0,
	}

	// File metadata
	c.writeFmt("source_filename = \"%s\"\n", path)
	c.writeRaw("\n")

	file := c.debug.pushScope(c.debug.file(path))
	c.debug.compileUnit(file)

	// Emit types
	c.types = collectTypes(&c.globals, decls)

	for _, pair := range c.types {
		if v, ok := pair.type_.(*ast.Struct); ok {
			c.writeRaw(pair.identifier)
			c.writeRaw(" = type { ")

			for i, field := range v.Fields {
				if i > 0 {
					c.writeRaw(", ")
				}

				c.writeRaw(c.getType(field.Type))
			}

			c.writeRaw(" }\n")
		}
	}

	c.writeRaw("\n")

	// Emit
	for _, decl := range decls {
		c.acceptDecl(decl)
	}

	// Import functions
	for _, f := range c.importedFunctions {
		type_ := f.function

		c.writeFmt("declare %s %s(", c.getType(type_.Returns), f.value.identifier)

		for i, param := range type_.Params {
			if i > 0 {
				c.writeStr(", ")
			}

			c.writeFmt("%s", c.getType(param.Type))
		}

		if type_.Variadic {
			if len(type_.Params) > 0 {
				c.writeStr(", ")
			}

			c.writeStr("...")
		}

		c.writeStr(")\n")
	}

	c.writeRaw("\n")

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

func (c *codegen) load(value exprValue, type_ types.Type) exprValue {
	if value.addressable {
		result := c.locals.unnamed()
		c.writeFmt("%s = load %s, ptr %s\n", result, c.getType(type_), value)

		return result
	}

	return value
}

func (c *codegen) loadExpr(expr ast.Expr) (exprValue, string) {
	value := c.load(c.acceptExpr(expr), expr.Result().Type)

	type_ := "ptr"
	if !value.addressable {
		type_ = c.getType(expr.Result().Type)
	}

	return value, type_
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

func (c *codegen) getType(type_ types.Type) string {
	// Try cache
	for _, pair := range c.types {
		if pair.type_.Equals(type_) {
			return pair.identifier
		}
	}

	// Enum
	if v, ok := type_.(*ast.Enum); ok {
		return c.getType(v.Type)
	}

	// Array
	if v, ok := type_.(*types.ArrayType); ok {
		value := c.globals.constant(fmt.Sprintf("[%d x %s]", v.Count, c.getType(v.Base)))

		c.types = append(c.types, typeValuePair{
			type_:      type_,
			identifier: value.identifier,
		})

		return value.identifier
	}

	// Pointer
	if _, ok := type_.(*types.PointerType); ok {
		value := c.globals.constant("ptr")

		c.types = append(c.types, typeValuePair{
			type_:      type_,
			identifier: value.identifier,
		})

		return value.identifier
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

		value := c.globals.constant(name)

		c.types = append(c.types, typeValuePair{
			type_:      type_,
			identifier: value.identifier,
		})

		return value.identifier
	}

	// Error
	panic("invalid type")
}

// Functions

func mangleName(name string) string {
	return "fb$" + name
}

func (c *codegen) getFunction(name scanner.Token) exprValue {
	// Get function
	function, filePath := c.resolver.GetFunction(name.Lexeme)

	mangledName := name.Lexeme
	if !function.Extern {
		mangledName = mangleName(name.Lexeme)
	}

	value := c.globals.named(mangledName)

	// Check if function needs to be imported
	if filePath != c.path {
		contains := false

		for _, f := range c.importedFunctions {
			if f.value.identifier == value.identifier {
				contains = true
				break
			}
		}

		if !contains {
			c.importedFunctions = append(c.importedFunctions, importedFunc{
				function: function,
				value:    value,
			})
		}
	}

	// Return
	return value
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

func (c *codegen) addVariable(name scanner.Token, value exprValue) *variable {
	value.addressable = true

	c.variables = append(c.variables, variable{
		name:  name,
		value: value,
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
	// Try cache
	for _, pair := range c.dbgTypes {
		if pair.type_.Equals(type_) {
			return pair.name
		}
	}

	// Struct
	if v, ok := type_.(*ast.Struct); ok {
		members := make([]string, len(v.Fields))
		offset := 0

		for i, field := range v.Fields {
			size := field.Type.Size() * 8
			members[i] = c.debug.derivedType(MemberDTag, field.Name.Lexeme, c.getDbgType(field.Type), size, offset)
			offset += size
		}

		name := c.debug.compositeType(StructureTypeCTag, "", v.Range().Start.Line, v.Size()*8, c.debug.tuple(members))
		c.dbgTypes = append(c.dbgTypes, typeDbgPair{
			type_: type_,
			name:  name,
		})

		return name
	}

	// Enum
	if v, ok := type_.(*ast.Enum); ok {
		name := c.debug.derivedType(TypedefDTag, v.Name.Lexeme, c.getDbgType(v.Type), v.Size()*8, 0)
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

func (c *codegen) acceptExpr(expr ast.Expr) exprValue {
	if expr != nil {
		expr.Accept(c)
		return c.exprResult
	}

	return exprValue{}
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

type exprValue struct {
	identifier  string
	addressable bool
}

func (v exprValue) String() string {
	return v.identifier
}

type values struct {
	char         string
	unnamedCount int
}

func (v *values) reset() {
	v.unnamedCount = 0
}

func (v *values) named(identifier string) exprValue {
	return exprValue{
		identifier: v.char + identifier,
	}
}

func (v *values) unnamed() exprValue {
	return exprValue{
		identifier: v.unnamedRaw(),
	}
}

func (v *values) unnamedRaw() string {
	v.unnamedCount++
	return v.char + strconv.Itoa(v.unnamedCount-1)
}

func (v *values) constant(constant string) exprValue {
	return exprValue{
		identifier: constant,
	}
}
