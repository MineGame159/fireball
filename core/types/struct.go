package types

type StructType struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type Type
}

func (s *StructType) Size() int {
	return 0
}

func (s *StructType) CanAssignTo(other Type) bool {
	return s == other
}

func (s *StructType) AcceptTypes(visitor Visitor) {
	for i := range s.Fields {
		visitor.VisitType(&s.Fields[i].Type)
	}
}

func (s *StructType) String() string {
	return s.Name
}

func (s *StructType) GetField(name string) (int, *Field) {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return i, &s.Fields[i]
		}
	}

	return 0, nil
}
