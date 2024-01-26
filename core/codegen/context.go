package codegen

import "fireball/core/ast"

type TypeId struct {
	Type ast.Type
	Id   uint32
}

type Context struct {
	TypeIds []TypeId
}

func (c *Context) GetTypeID(type_ ast.Type) uint32 {
	// Check cache
	for _, id := range c.TypeIds {
		if id.Type.Equals(type_) {
			return id.Id
		}
	}

	// Insert into cache
	c.TypeIds = append(c.TypeIds, TypeId{
		Type: type_,
		Id:   uint32(len(c.TypeIds)),
	})

	return uint32(len(c.TypeIds) - 1)
}
