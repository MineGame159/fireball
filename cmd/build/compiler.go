package build

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type Compiler struct {
	OptimizationLevel int

	inputs    []string
	libraries []string
}

func (c *Compiler) AddInput(input string) {
	c.inputs = append(c.inputs, input)
}

func (c *Compiler) AddLibrary(library string) {
	c.libraries = append(c.libraries, library)
}

func (c *Compiler) Compile(output string) error {
	if c.OptimizationLevel == 0 {
		// Compile each IR file individually
		for _, input := range c.inputs {
			err := c.compileIr(input)
			if err != nil {
				return err
			}
		}

		return c.linkExecutable(c.inputs, output)
	}

	// Link all IR files together
	ir := filepath.Join(filepath.Dir(output), "__ir.bc")

	err := c.linkIr(ir)
	if err != nil {
		return err
	}

	err = c.optimizeIr(ir)
	if err != nil {
		return err
	}

	err = c.compileIr(ir)
	if err != nil {
		return err
	}

	return c.linkExecutable([]string{ir}, output)
}

func (c *Compiler) linkIr(output string) error {
	// Create command
	cmd := exec.Command("llvm-link")

	for _, input := range c.inputs {
		cmd.Args = append(cmd.Args, input)
	}

	cmd.Args = append(cmd.Args, "-o")
	cmd.Args = append(cmd.Args, output)

	// Execute
	return execute(cmd)
}

func (c *Compiler) optimizeIr(input string) error {
	// Create command
	cmd := exec.Command("opt", input, fmt.Sprintf("-O%d", c.OptimizationLevel), "-o", withExtension(input, "bc"))

	// Execute
	return execute(cmd)
}

func (c *Compiler) compileIr(input string) error {
	// Create command
	cmd := exec.Command("llc", input, fmt.Sprintf("-O%d", c.OptimizationLevel))

	if c.OptimizationLevel == 0 {
		cmd.Args = append(cmd.Args, "--frame-pointer")
		cmd.Args = append(cmd.Args, "all")
	}

	cmd.Args = append(cmd.Args, "--filetype")
	cmd.Args = append(cmd.Args, "obj")

	cmd.Args = append(cmd.Args, "-o")
	cmd.Args = append(cmd.Args, withExtension(input, "o"))

	// Execute
	return execute(cmd)
}

func (c *Compiler) linkExecutable(inputs []string, output string) error {
	l := GetLinker()

	err := l.Check()
	if err != nil {
		return err
	}

	for _, library := range c.libraries {
		l.AddLibrary(library)
	}

	for _, input := range inputs {
		l.AddInput(withExtension(input, "o"))
	}

	return l.Link(output)
}

func execute(cmd *exec.Cmd) error {
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.New(err.Error() + " : " + stderr.String())
	}

	if !cmd.ProcessState.Success() {
		return errors.New(stderr.String())
	}

	return err
}

func withExtension(path, extension string) string {
	dot := strings.LastIndexByte(path, '.')
	return path[:dot+1] + extension
}
