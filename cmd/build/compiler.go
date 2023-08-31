package build

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type Compiler struct {
	OptimizationLevel int
}

func (o *Compiler) Compile(input, output string) error {
	// Create command
	cmd := exec.Command("llc", input, fmt.Sprintf("-O%d", o.OptimizationLevel), "--filetype", "obj", "-o", output)

	// Execute
	out := bytes.Buffer{}
	cmd.Stderr = &out

	err := cmd.Run()

	if !cmd.ProcessState.Success() {
		return errors.New(out.String())
	}

	return err
}
