package llvm

type Location interface {
	Line() int
	Column() int
}

type Module struct {
	source string

	types []Type

	constants    []*constant
	constantType Type

	declares []*functionType
	defines  []*Function

	namedMetadata map[string]Metadata
	metadata      []Metadata

	typeMetadata map[Type]int
	scopes       []int
}

func NewModule() *Module {
	return &Module{
		namedMetadata: make(map[string]Metadata),
		typeMetadata:  make(map[Type]int),
	}
}

func (m *Module) Source(path string) {
	m.source = path
	producer := "fireball version 0.1.0"

	file := m.addMetadata(fileMetadata(path))
	m.scopes = append(m.scopes, file)

	m.namedMetadata["llvm.dbg.cu"] = Metadata{Fields: []MetadataField{
		{Value: refMetadataValue(m.addMetadata(compileUnitMetadata(file, producer)))},
	}}

	m.namedMetadata["llvm.module.flags"] = Metadata{Fields: []MetadataField{
		{Value: refMetadataValue(m.addFlagMetadata(7, "Dwarf Version", 4))},
		{Value: refMetadataValue(m.addFlagMetadata(2, "Debug Info Version", 3))},
		{Value: refMetadataValue(m.addFlagMetadata(1, "wchar_size", 4))},
		{Value: refMetadataValue(m.addFlagMetadata(8, "PIC Level", 2))},
		{Value: refMetadataValue(m.addFlagMetadata(7, "PIE Level", 2))},
		{Value: refMetadataValue(m.addFlagMetadata(7, "uwtable", 2))},
		{Value: refMetadataValue(m.addFlagMetadata(7, "frame-pointer", 2))},
	}}

	m.namedMetadata["llvm.ident"] = Metadata{Fields: []MetadataField{
		{Value: refMetadataValue(m.addMetadata(Metadata{Fields: []MetadataField{{Value: stringMetadataValue(producer)}}}))},
	}}
}

// Functions

type declare struct {
	type_ Type
	name  string
}

func (d *declare) Kind() ValueKind {
	return GlobalValue
}

func (d *declare) Type() Type {
	return d.type_
}

func (d *declare) Name() string {
	return d.name
}

func (m *Module) Declare(type_ Type) Value {
	m.declares = append(m.declares, type_.(*functionType))

	return &declare{
		type_: type_,
		name:  type_.(*functionType).name,
	}
}

func (m *Module) Define(type_ Type) *Function {
	t := type_.(*functionType)

	metadata := m.addMetadata(Metadata{
		Distinct: true,
		Type:     "DISubprogram",
		Fields: []MetadataField{
			{
				Name:  "name",
				Value: stringMetadataValue(t.name),
			},
			{
				Name:  "scope",
				Value: refMetadataValue(m.getScope()),
			},
			{
				Name:  "file",
				Value: refMetadataValue(m.getFile()),
			},
			{
				Name:  "type",
				Value: refMetadataValue(m.typeMetadata[type_]),
			},
			{
				Name:  "spFlags",
				Value: enumMetadataValue("DISPFlagDefinition"),
			},
			{
				Name:  "unit",
				Value: refMetadataValue(m.getCompileUnit()),
			},
		},
	})

	f := &Function{
		module: m,

		type_: t,

		parameters: make([]NameableValue, len(t.parameters)),
		blocks:     make([]*Block, 0, 4),

		metadata: metadata,
	}

	for i, parameter := range t.parameters {
		f.parameters[i] = &instruction{
			type_: parameter,
		}
	}

	m.defines = append(m.defines, f)
	return f
}

// Types

func (m *Module) Void() Type {
	return &voidType{}
}

func (m *Module) Primitive(name string, bitSize int, encoding Encoding) Type {
	t := &primitiveType{
		name:     name,
		bitSize:  bitSize,
		encoding: encoding,
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Type: "DIBasicType",
		Fields: []MetadataField{
			{
				Name:  "name",
				Value: stringMetadataValue(name),
			},
			{
				Name:  "size",
				Value: numberMetadataValue(bitSize),
			},
			{
				Name:  "encoding",
				Value: enumMetadataValue(string(encoding)),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

func (m *Module) Array(name string, count int, base Type) Type {
	t := &arrayType{
		name:  name,
		count: count,
		base:  base,
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Type: "DICompositeType",
		Fields: []MetadataField{
			{
				Name:  "tag",
				Value: enumMetadataValue("DW_TAG_array_type"),
			},
			{
				Name:  "baseType",
				Value: refMetadataValue(m.typeMetadata[base]),
			},
			{
				Name:  "size",
				Value: numberMetadataValue(t.Size()),
			},
			{
				Name: "elements",
				Value: refMetadataValue(m.addMetadata(Metadata{Fields: []MetadataField{
					{Value: refMetadataValue(m.addMetadata(Metadata{
						Type: "DISubrange",
						Fields: []MetadataField{
							{
								Name:  "count",
								Value: numberMetadataValue(count),
							},
						},
					}))},
				}})),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

func (m *Module) Pointer(name string, pointee Type) Type {
	t := &pointerType{
		name:    name,
		pointee: pointee,
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Type: "DIDerivedType",
		Fields: []MetadataField{
			{
				Name:  "tag",
				Value: enumMetadataValue("DW_TAG_pointer_type"),
			},
			{
				Name:  "baseType",
				Value: refMetadataValue(m.typeMetadata[pointee]),
			},
			{
				Name:  "size",
				Value: numberMetadataValue(t.Size()),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

func (m *Module) Function(name string, parameters []Type, variadic bool, returns Type) Type {
	t := &functionType{
		name:       name,
		parameters: parameters,
		variadic:   variadic,
		returns:    returns,
	}

	types := make([]MetadataField, len(parameters)+1)

	if _, ok := returns.(*voidType); ok {
		types[0] = MetadataField{Value: enumMetadataValue("null")}
	} else {
		types[0] = MetadataField{Value: refMetadataValue(m.typeMetadata[returns])}
	}

	for i, parameter := range parameters {
		types[i+1] = MetadataField{Value: refMetadataValue(m.typeMetadata[parameter])}
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Type: "DISubroutineType",
		Fields: []MetadataField{
			{
				Name:  "types",
				Value: refMetadataValue(m.addMetadata(Metadata{Fields: types})),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

func (m *Module) Struct(name string, fields []Field) Type {
	t := &structType{
		name:   name,
		fields: fields,
	}

	elements := make([]MetadataField, len(fields))
	offset := 0

	for i, field := range fields {
		elements[i] = MetadataField{Value: refMetadataValue(m.addMetadata(Metadata{
			Type: "DIDerivedType",
			Fields: []MetadataField{
				{
					Name:  "tag",
					Value: enumMetadataValue("DW_TAG_member"),
				},
				{
					Name:  "baseType",
					Value: refMetadataValue(m.typeMetadata[field.Type]),
				},
				{
					Name:  "name",
					Value: stringMetadataValue(field.Name),
				},
				{
					Name:  "scope",
					Value: refMetadataValue(m.getScope()),
				},
				{
					Name:  "file",
					Value: refMetadataValue(m.getFile()),
				},
				{
					Name:  "offset",
					Value: numberMetadataValue(offset),
				},
			},
		}))}

		offset += field.Type.Size()
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Distinct: true,
		Type:     "DICompositeType",
		Fields: []MetadataField{
			{
				Name:  "tag",
				Value: enumMetadataValue("DW_TAG_structure_type"),
			},
			{
				Name:  "name",
				Value: stringMetadataValue(name),
			},
			{
				Name:  "file",
				Value: refMetadataValue(m.getFile()),
			},
			{
				Name:  "size",
				Value: numberMetadataValue(t.Size()),
			},
			{
				Name:  "flags",
				Value: enumMetadataValue("DIFlagTypePassByValue"),
			},
			{
				Name:  "elements",
				Value: refMetadataValue(m.addMetadata(Metadata{Fields: elements})),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

func (m *Module) Alias(name string, underlying Type) Type {
	t := &aliasType{
		name:       name,
		underlying: underlying,
	}

	m.typeMetadata[t] = m.addMetadata(Metadata{
		Type: "DIDerivedType",
		Fields: []MetadataField{
			{
				Name:  "tag",
				Value: enumMetadataValue("DW_TAG_typedef"),
			},
			{
				Name:  "baseType",
				Value: refMetadataValue(m.typeMetadata[underlying]),
			},
			{
				Name:  "size",
				Value: numberMetadataValue(underlying.Size()),
			},
		},
	})

	m.types = append(m.types, t)
	return t
}

// Constants

type constant struct {
	type_ Type
	data  string
}

func (c *constant) Kind() ValueKind {
	return GlobalValue
}

func (c *constant) Type() Type {
	return c.type_
}

func (c *constant) Name() string {
	return ""
}

func (m *Module) Constant(data string) Value {
	// Get constant from already created constants
	for _, c := range m.constants {
		if c.data == data {
			return c
		}
	}

	// Create constant type
	if m.constantType == nil {
		m.constantType = m.Pointer("*u8", m.Primitive("u8", 8, UnsignedEncoding))
	}

	// Create constant
	c := &constant{
		type_: m.constantType,
		data:  data,
	}

	m.constants = append(m.constants, c)
	return c
}

// Metadata

func (m *Module) PushScope(location Location) {
	m.scopes = append(m.scopes, m.addMetadata(Metadata{
		Distinct: true,
		Type:     "DILexicalBlock",
		Fields: []MetadataField{
			{
				Name:  "scope",
				Value: refMetadataValue(m.getScope()),
			},
			{
				Name:  "file",
				Value: refMetadataValue(m.getFile()),
			},
			{
				Name:  "line",
				Value: numberMetadataValue(location.Line()),
			},
			{
				Name:  "column",
				Value: numberMetadataValue(location.Line()),
			},
		},
	}))
}

func (m *Module) PopScope() {
	m.scopes = m.scopes[:len(m.scopes)-1]
}

func (m *Module) getScope() int {
	return m.scopes[len(m.scopes)-1]
}

func (m *Module) getFile() int {
	return m.scopes[0]
}

func (m *Module) getCompileUnit() int {
	return 1
}

func (m *Module) addFlagMetadata(num1 int, str string, num2 int) int {
	return m.addMetadata(Metadata{Fields: []MetadataField{
		{Value: numberMetadataValue(num1)},
		{Value: stringMetadataValue(str)},
		{Value: numberMetadataValue(num2)},
	}})
}

func (m *Module) addMetadata(metadata Metadata) int {
	m.metadata = append(m.metadata, metadata)
	return len(m.metadata) - 1
}
