package ast

type SymbolVisitor interface {
	VisitSymbol(node Node)
}

type Resolver interface {
	GetChild(name string) Resolver
	GetType(name string) Type

	GetFunction(name string) *Func
	GetVariable(name string) *GlobalVar

	GetMethod(type_ Type, name string, static bool) FuncType
	GetMethods(type_ Type, static bool) []*Func
	GetImpl(type_ Type, inter *Interface) *Impl

	GetChildren() []string
	GetSymbols(visitor SymbolVisitor)
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

func (c *CombinedResolver) GetVariable(name string) *GlobalVar {
	for _, resolver := range c.resolvers {
		if variable := resolver.GetVariable(name); variable != nil {
			return variable
		}
	}

	return nil
}

func (c *CombinedResolver) GetMethod(type_ Type, name string, static bool) FuncType {
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

func (c *CombinedResolver) GetImpl(type_ Type, inter *Interface) *Impl {
	for _, resolver := range c.resolvers {
		if impl := resolver.GetImpl(type_, inter); impl != nil {
			return impl
		}
	}

	return nil
}

func (c *CombinedResolver) GetChildren() []string {
	var children []string

	for _, resolver := range c.resolvers {
		children = append(children, resolver.GetChildren()...)
	}

	return children
}

func (c *CombinedResolver) GetSymbols(visitor SymbolVisitor) {
	for _, resolver := range c.resolvers {
		resolver.GetSymbols(visitor)
	}
}

// GenericResolver

type genericResolver struct {
	base     Resolver
	generics []*Generic
}

func NewGenericResolver(base Resolver, generics []*Generic) Resolver {
	return &genericResolver{
		base:     base,
		generics: generics,
	}
}

func (g *genericResolver) GetChild(name string) Resolver {
	return g.base.GetChild(name)
}

func (g *genericResolver) GetType(name string) Type {
	for _, generic := range g.generics {
		if generic.String() == name {
			return generic
		}
	}

	return g.base.GetType(name)
}

func (g *genericResolver) GetFunction(name string) *Func {
	return g.base.GetFunction(name)
}

func (g *genericResolver) GetVariable(name string) *GlobalVar {
	return g.base.GetVariable(name)
}

func (g *genericResolver) GetMethod(type_ Type, name string, static bool) FuncType {
	return g.base.GetMethod(type_, name, static)
}

func (g *genericResolver) GetMethods(type_ Type, static bool) []*Func {
	return g.base.GetMethods(type_, static)
}

func (g *genericResolver) GetImpl(type_ Type, inter *Interface) *Impl {
	return g.base.GetImpl(type_, inter)
}

func (g *genericResolver) GetChildren() []string {
	return g.base.GetChildren()
}

func (g *genericResolver) GetSymbols(visitor SymbolVisitor) {
	for _, generic := range g.generics {
		visitor.VisitSymbol(generic)
	}

	g.base.GetSymbols(visitor)
}
