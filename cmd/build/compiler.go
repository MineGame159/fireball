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
	Crt               bool

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
	// Create command
	cmd := exec.Command("ld.lld", "-dynamic-linker", "/lib64/ld-linux-x86-64.so.2", "-L/usr/lib")

	if c.Crt {
		cmd.Args = append(cmd.Args, "/usr/lib/crt1.o")
		cmd.Args = append(cmd.Args, "/usr/lib/crti.o")
	}

	for _, library := range c.libraries {
		cmd.Args = append(cmd.Args, "-l"+library)
	}

	for _, input := range inputs {
		cmd.Args = append(cmd.Args, withExtension(input, "o"))
	}

	if c.Crt {
		cmd.Args = append(cmd.Args, "/usr/lib/crtn.o")
	}

	cmd.Args = append(cmd.Args, "-o")
	cmd.Args = append(cmd.Args, output)

	// Execute
	return execute(cmd)
}

func execute(cmd *exec.Cmd) error {
	out := bytes.Buffer{}
	cmd.Stderr = &out

	err := cmd.Run()

	if !cmd.ProcessState.Success() {
		return errors.New(out.String())
	}

	return err
}

func withExtension(path, extension string) string {
	dot := strings.LastIndexByte(path, '.')
	return path[:dot+1] + extension
}
