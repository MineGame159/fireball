package codegen

import "fireball/core/ast"

type typeId struct {
	type_ ast.Type
	id    uint32
}

type Context struct {
	typeIds []typeId
}

func (c *Context) GetTypeID(type_ ast.Type) uint32 {
	// Check cache
	for _, id := range c.typeIds {
		if id.type_.Equals(type_) {
			return id.id
		}
	}

	// Insert into cache
	c.typeIds = append(c.typeIds, typeId{
		type_: type_,
		id:    uint32(len(c.typeIds)),
	})

	return uint32(len(c.typeIds) - 1)
}
