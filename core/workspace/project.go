package workspace

import (
	"errors"
	"fireball/core/ast"
	"fireball/core/ast/cst2ast"
	"fireball/core/checker"
	"fireball/core/cst"
	"fireball/core/typeresolver"
	"fireball/core/utils"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Path   string
	Config Config

	Files     map[string]*File
	namespace namespace
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
	p := &Project{
		Path:   path,
		Config: config,

		Files: make(map[string]*File),
	}

	p.namespace.project = p

	return p, nil
}

func NewEmptyProject(path, name string) *Project {
	p := &Project{
		Path: path,
		Config: Config{
			Name: name,
			Src:  ".",
		},

		Files: make(map[string]*File),
	}

	p.namespace.project = p

	return p
}

// Namespaces

func (p *Project) addFileToNamespace(file *File) {
	// Check namespace against config
	if file.Ast == nil || file.Ast.Namespace == nil || file.Ast.Namespace.Name == nil {
		return
	}

	parts := file.Ast.Namespace.Name.Parts
	ok := true

	for i, part := range p.Config.Namespace {
		if i >= len(parts) || part != parts[i].String() {
			ok = false
			break
		}
	}

	if !ok {
		file.Report(utils.Diagnostic{
			Kind:    utils.ErrorKind,
			Range:   file.Ast.Namespace.Name.Cst().Range,
			Message: fmt.Sprintf("All files in this project need to have a base namespace '%s'", p.Config.Name),
		})

		return
	}

	// Add
	if n := p.getNamespace(file.Ast); n != nil {
		n.addFile(file)
	}
}

func (p *Project) removeFileFromNamespace(file *File) {
	if n := p.getNamespace(file.Ast); n != nil {
		n.removeFile(file)
	}
}

func (p *Project) getNamespace(file *ast.File) *namespace {
	if file == nil || file.Namespace == nil || file.Namespace.Name == nil {
		return nil
	}

	n := &p.namespace

	for _, part := range file.Namespace.Name.Parts {
		n = n.getOrCreateChild(part.String())
	}

	return n
}

func (p *Project) GetResolverFile(file *ast.File) ast.RootResolver {
	if n := p.getNamespace(file); n != nil {
		return n
	}

	return nil
}

func (p *Project) GetResolverName(parts []string) ast.RootResolver {
	n := &p.namespace

	for _, part := range parts {
		n = n.getOrCreateChild(part)
	}

	return n
}

func (p *Project) GetResolverRoot() ast.RootResolver {
	return &p.namespace
}

// Files

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

		p.addFileToNamespace(file)

		file.parseDiagnosticCount = len(file.diagnostics)
		file.parseWaitGroup.Done()
	}

	// Resolve types
	for _, file := range p.Files {
		if resolver := p.getNamespace(file.Ast); resolver != nil {
			file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
			typeresolver.Resolve(file, resolver, file.Ast)
		}
	}

	// Check
	for _, file := range p.Files {
		if resolver := p.getNamespace(file.Ast); resolver != nil {
			checker.Check(file, resolver, file.Ast)
		}

		file.checkWaitGroup.Done()
	}

	return nil
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
	if file, ok := p.Files[path]; ok {
		// Delete file
		p.removeFileFromNamespace(file)
		delete(p.Files, path)

		// Check the rest of the files
		for _, file := range p.Files {
			if resolver := p.getNamespace(file.Ast); resolver != nil {
				file.checkWaitGroup.Add(1)

				file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
				typeresolver.Resolve(file, resolver, file.Ast)
			}
		}

		for _, file := range p.Files {
			if resolver := p.getNamespace(file.Ast); resolver != nil {
				checker.Check(file, resolver, file.Ast)
				file.checkWaitGroup.Done()
			}
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
