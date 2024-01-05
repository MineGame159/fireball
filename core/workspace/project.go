package workspace

import (
	"errors"
	"fireball/core/ast"
	"fireball/core/ast/cst2ast"
	"fireball/core/checker"
	"fireball/core/cst"
	"fireball/core/typeresolver"
	"fireball/core/utils"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Path   string
	Config Config

	Files map[string]*File
}

type Config struct {
	Name string
	Src  string
}

func NewProject(path string) (*Project, error) {
	// Check path
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("project path needs to be a folder")
	}

	// Parse config
	file, err := os.Open(filepath.Join(path, "project.toml"))
	if err != nil {
		return nil, err
	}

	var config Config
	err = toml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	// Validate config
	if config.Name == "" {
		return nil, errors.New("invalid project name")
	}
	if config.Src == "" {
		return nil, errors.New("invalid project src folder")
	}

	// Return
	return &Project{
		Path:   path,
		Config: config,

		Files: make(map[string]*File),
	}, nil
}

func NewEmptyProject(path, name string) *Project {
	return &Project{
		Path: path,
		Config: Config{
			Name: name,
			Src:  ".",
		},

		Files: make(map[string]*File),
	}
}

func (p *Project) LoadFiles() error {
	// Get source files
	files, err := p.GetSourceFiles()
	if err != nil {
		return err
	}

	// Loop files
	for _, path := range files {
		// Read file
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse file
		relative, err := filepath.Rel(p.Path, path)
		if err != nil {
			return err
		}

		file := p.GetOrCreateFile(relative)
		file.SetText(string(contents), false)
	}

	// Parse
	for _, file := range p.Files {
		file.parseWaitGroup.Add(1)
		file.checkWaitGroup.Add(1)
	}

	for _, file := range p.Files {
		file.diagnostics = nil

		file.Cst = cst.Parse(file, file.Text)
		file.Ast = cst2ast.Convert(file, file.AbsolutePath(), file.Cst)

		file.parseDiagnosticCount = len(file.diagnostics)
	}

	for _, file := range p.Files {
		file.CollectTypesAndFunctions()
		file.parseWaitGroup.Done()
	}

	for _, file := range p.Files {
		file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
		typeresolver.Resolve(file, p, file.Ast)
	}

	// Check
	for _, file := range p.Files {
		checker.Check(file, p, file.Ast)
		file.checkWaitGroup.Done()
	}

	return nil
}

func (p *Project) GetType(name string) (ast.Type, string) {
	for _, file := range p.Files {
		if v, ok := file.Types[name]; ok {
			return v, file.Path
		}
	}

	return nil, ""
}

func (p *Project) GetFunction(name string) (*ast.Func, string) {
	for _, file := range p.Files {
		if v, ok := file.Functions[name]; ok {
			return v, file.Path
		}
	}

	return nil, ""
}

func (p *Project) GetMethod(type_ ast.Type, name string, static bool) (*ast.Func, string) {
	for _, file := range p.Files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil && impl.Type.Equals(type_) {
				function := impl.GetMethod(name, static)

				if function != nil {
					return function, file.Path
				}
			}
		}
	}

	return nil, ""
}

func (p *Project) GetMethods(type_ ast.Type, static bool) []*ast.Func {
	var methods []*ast.Func

	staticValue := ast.FuncFlags(0)
	if static {
		staticValue = 1
	}

	for _, file := range p.Files {
		for _, decl := range file.Ast.Decls {
			if impl, ok := decl.(*ast.Impl); ok && impl.Type != nil && impl.Type.Equals(type_) {
				for _, method := range impl.Methods {
					if method.Flags&ast.Static == staticValue {
						methods = append(methods, method)
					}
				}
			}
		}
	}

	return methods
}

func (p *Project) GetFileNodes() []*ast.File {
	nodes := make([]*ast.File, 0, len(p.Files))

	for _, file := range p.Files {
		nodes = append(nodes, file.Ast)
	}

	return nodes
}

func (p *Project) GetOrCreateFile(path string) *File {
	if file, ok := p.Files[path]; ok {
		return file
	}

	file := &File{
		Project:     p,
		Path:        path,
		diagnostics: make([]utils.Diagnostic, 0),
	}

	p.Files[path] = file
	return file
}

func (p *Project) RemoveFileAbs(path string) bool {
	// Get relative path
	relative, err := filepath.Rel(p.Path, path)
	if err != nil {
		return false
	}

	// Remove
	return p.RemoveFile(relative)
}

func (p *Project) RemoveFile(path string) bool {
	// Find file
	if _, ok := p.Files[path]; ok {
		// Delete file
		delete(p.Files, path)

		// Check the rest of the files
		for _, file := range p.Files {
			file.checkWaitGroup.Add(1)

			file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
			typeresolver.Resolve(file, file.Project, file.Ast)
		}

		for _, file := range p.Files {
			checker.Check(file, p, file.Ast)
			file.checkWaitGroup.Done()
		}

		// Return true
		return true
	}

	// Return false
	return false
}

func (p *Project) GetFileAbs(path string) *File {
	relative, err := filepath.Rel(p.Path, path)
	if err != nil {
		return nil
	}

	if file, ok := p.Files[relative]; ok {
		return file
	}

	return nil
}

func (p *Project) GetSourceFiles() ([]string, error) {
	folders := make([]folder, 0, 8)
	files := make([]string, 0, 8)

	// Add initial folder
	path := filepath.Join(p.Path, p.Config.Src)
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	folders = append(folders, folder{
		path:    path,
		entries: entries,
	})

	// Process folders
	for len(folders) > 0 {
		// Get folder
		current := folders[len(folders)-1]
		folders = folders[:len(folders)-1]

		// Check entries
		for _, entry := range current.entries {
			path := filepath.Join(current.path, entry.Name())

			if entry.IsDir() {
				// Folder
				entries, err := os.ReadDir(path)
				if err != nil {
					return nil, err
				}

				folders = append(folders, folder{
					path:    path,
					entries: entries,
				})
			} else {
				// File
				if strings.HasSuffix(path, ".fb") {
					files = append(files, path)
				}
			}
		}
	}

	// Return
	return files, nil
}

type folder struct {
	path    string
	entries []os.DirEntry
}
