package workspace

import (
	"fireball/core/ast"
	"slices"
)

type namespace struct {
	project *Project
	name    string

	parent   *namespace
	children []*namespace

	files []*File
}

func (n *namespace) getOrCreateChild(name string) *namespace {
	// Get
	for _, child := range n.children {
		if child.name == name {
			return child
		}
	}

	// Create
	child := &namespace{
		project: n.project,
		name:    name,
		parent:  n,
	}

	n.children = append(n.children, child)
	return child
}

func (n *namespace) addFile(file *File) {
	n.files = append(n.files, file)
}

func (n *namespace) removeFile(file *File) {
	if i := slices.Index(n.files, file); i != -1 {
		n.files[i] = n.files[len(n.files)-1]
		n.files = n.files[:len(n.files)-1]
	}

	n.deleteIfEmpty()
}

func (n *namespace) removeChild(child *namespace) {
	if i := slices.Index(n.children, child); i != -1 {
		n.children[i] = n.children[len(n.children)-1]
		n.children = n.children[:len(n.children)-1]
	}

	n.deleteIfEmpty()
}

func (n *namespace) deleteIfEmpty() {
	if n.parent != nil && len(n.children) == 0 && len(n.files) == 0 {
		n.parent.removeChild(n)
	}
}

// RootResolver

func (n *namespace) GetResolver(name *ast.NamespaceName) ast.Resolver {
	if name == nil {
		return nil
	}

	var resolver ast.Resolver = &n.project.namespace

	for _, part := range name.Parts {
		resolver = resolver.GetChild(part.String())

		if resolver == nil {
			return nil
		}
	}

	return resolver
}

// Resolver

func (n *namespace) GetChild(name string) ast.Resolver {
	for _, child := range n.children {
		if child.name == name {
			return child
		}
	}

	return nil
}

func (n *namespace) GetType(name string) ast.Type {
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			switch decl := decl.(type) {
			case *ast.Struct:
				if decl.Name != nil && decl.Name.String() == name {
					return decl
				}

			case *ast.Enum:
				if decl.Name != nil && decl.Name.String() == name {
					return decl
				}

			case *ast.Interface:
				if decl.Name != nil && decl.Name.String() == name {
					return decl
				}
			}
		}
	}

	return nil
}

func (n *namespace) GetFunction(name string) *ast.Func {
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			if function, ok := decl.(*ast.Func); ok && function.Name != nil && function.Name.String() == name {
				return function
			}
		}
	}

	return nil
}

func (n *namespace) GetVariable(name string) *ast.GlobalVar {
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			if variable, ok := decl.(*ast.GlobalVar); ok && variable.Name != nil && variable.Name.String() == name {
				return variable
			}
		}
	}

	return nil
}

func (n *namespace) GetMethod(type_ ast.Type, name string, static bool) ast.FuncType {
	guh := type_

	if s, ok := guh.(ast.StructType); ok {
		guh = s.Underlying()
	}

	// Find method
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil && impl.Type.Equals(guh) {
				if method := impl.GetMethod(name, static); method != nil {
					// Specialize if needed
					if s, ok := type_.(*ast.SpecializedStruct); ok && !static {
						return s.SpecializeMethod(method)
					}

					return method
				}
			}
		}
	}

	return nil
}

func (n *namespace) GetMethods(type_ ast.Type, static bool) []*ast.Func {
	if s, ok := type_.(ast.StructType); ok {
		type_ = s.Underlying()
	}

	var methods []*ast.Func

	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil && impl.Type.Equals(type_) {
				for _, method := range impl.Methods {
					if (static && method.IsStatic()) || (!static && !method.IsStatic()) {
						methods = append(methods, method)
					}
				}
			}
		}
	}

	return methods
}

func (n *namespace) GetImpl(type_ ast.Type, inter *ast.Interface) *ast.Impl {
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && type_.Equals(impl.Type) && inter.Equals(impl.Implements) {
				return impl
			}
		}
	}

	return nil
}

func (n *namespace) GetChildren() []string {
	children := make([]string, len(n.children))

	for i, child := range n.children {
		children[i] = child.name
	}

	return children
}

func (n *namespace) GetSymbols(visitor ast.SymbolVisitor) {
	for _, file := range n.files {
		for _, decl := range file.Ast.Decls {
			switch decl.(type) {
			case *ast.Struct, *ast.Impl, *ast.Enum, *ast.Interface, *ast.Func, *ast.GlobalVar:
				visitor.VisitSymbol(decl)
			}
		}
	}
}
