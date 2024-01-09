package ast

type Resolver interface {
	GetChild(name string) Resolver
	GetType(name string) Type

	GetFunction(name string) *Func

	GetMethod(type_ Type, name string, static bool) *Func
	GetMethods(type_ Type, static bool) []*Func

	GetChildren() []string
	GetSymbols() []Node
}

type RootResolver interface {
	Resolver

	GetResolver(name *NamespaceName) Resolver
}

// CombinedResolver

type CombinedResolver struct {
	resolvers []Resolver
}

func NewCombinedResolver(base Resolver) *CombinedResolver {
	return &CombinedResolver{resolvers: []Resolver{base}}
}

func (c *CombinedResolver) Add(resolver Resolver) {
	c.resolvers = append(c.resolvers, resolver)
}

func (c *CombinedResolver) GetChild(name string) Resolver {
	for _, resolver := range c.resolvers {
		if child := resolver.GetChild(name); child != nil {
			return child
		}
	}

	return nil
}

func (c *CombinedResolver) GetType(name string) Type {
	for _, resolver := range c.resolvers {
		if type_ := resolver.GetType(name); type_ != nil {
			return type_
		}
	}

	return nil
}

func (c *CombinedResolver) GetFunction(name string) *Func {
	for _, resolver := range c.resolvers {
		if function := resolver.GetFunction(name); function != nil {
			return function
		}
	}

	return nil
}

func (c *CombinedResolver) GetMethod(type_ Type, name string, static bool) *Func {
	for _, resolver := range c.resolvers {
		if method := resolver.GetMethod(type_, name, static); method != nil {
			return method
		}
	}

	return nil
}

func (c *CombinedResolver) GetMethods(type_ Type, static bool) []*Func {
	var methods []*Func

	for _, resolver := range c.resolvers {
		methods = append(methods, resolver.GetMethods(type_, static)...)
	}

	return methods
}

func (c *CombinedResolver) GetChildren() []string {
	var children []string

	for _, resolver := range c.resolvers {
		children = append(children, resolver.GetChildren()...)
	}

	return children
}

func (c *CombinedResolver) GetSymbols() []Node {
	var symbols []Node

	for _, resolver := range c.resolvers {
		symbols = append(symbols, resolver.GetSymbols()...)
	}

	return symbols
}
