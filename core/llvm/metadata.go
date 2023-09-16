package llvm

import "path/filepath"

type MetadataValueKind uint8

const (
	StringMetadataValueKind MetadataValueKind = iota
	EnumMetadataValueKind
	NumberMetadataValueKind
	RefMetadataValueKind
)

type MetadataValue struct {
	Kind MetadataValueKind

	String string
	Number int
}

type MetadataField struct {
	Name  string
	Value MetadataValue
}

type Metadata struct {
	Distinct bool
	Type     string
	Fields   []MetadataField
}

func fileMetadata(path string) Metadata {
	return Metadata{
		Type: "DIFile",
		Fields: []MetadataField{
			{
				Name:  "filename",
				Value: stringMetadataValue(filepath.Base(path)),
			},
			{
				Name:  "directory",
				Value: stringMetadataValue(filepath.Dir(path)),
			},
		},
	}
}

func compileUnitMetadata(file int, producer string) Metadata {
	return Metadata{
		Distinct: true,
		Type:     "DICompileUnit",
		Fields: []MetadataField{
			{
				Name:  "language",
				Value: enumMetadataValue("DW_LANG_C"),
			},
			{
				Name:  "file",
				Value: refMetadataValue(file),
			},
			{
				Name:  "producer",
				Value: stringMetadataValue(producer),
			},
			{
				Name:  "isOptimized",
				Value: enumMetadataValue("false"),
			},
			{
				Name:  "runtimeVersion",
				Value: numberMetadataValue(0),
			},
			{
				Name:  "emissionKind",
				Value: enumMetadataValue("FullDebug"),
			},
			{
				Name:  "splitDebugInlining",
				Value: enumMetadataValue("false"),
			},
			{
				Name:  "nameTableKind",
				Value: enumMetadataValue("None"),
			},
		},
	}
}

func stringMetadataValue(str string) MetadataValue {
	return MetadataValue{
		Kind:   StringMetadataValueKind,
		String: str,
	}
}

func enumMetadataValue(str string) MetadataValue {
	return MetadataValue{
		Kind:   EnumMetadataValueKind,
		String: str,
	}
}

func numberMetadataValue(num int) MetadataValue {
	return MetadataValue{
		Kind:   NumberMetadataValueKind,
		Number: num,
	}
}

func refMetadataValue(index int) MetadataValue {
	return MetadataValue{
		Kind:   RefMetadataValueKind,
		Number: index,
	}
}
