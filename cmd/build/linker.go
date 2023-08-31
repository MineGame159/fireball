package build

import (
	"bytes"
	"errors"
	"os/exec"
)

type Linker struct {
	Crt bool

	libraries []string
	inputs    []string
}

func (l *Linker) AddLibrary(library string) {
	l.libraries = append(l.libraries, library)
}

func (l *Linker) AddInput(input string) {
	l.inputs = append(l.inputs, input)
}

func (l *Linker) Link(output string) error {
	// Create command
	cmd := exec.Command("ld.lld", "-dynamic-linker", "/lib64/ld-linux-x86-64.so.2", "-L/usr/lib")

	if l.Crt {
		cmd.Args = append(cmd.Args, "/usr/lib/crt1.o")
		cmd.Args = append(cmd.Args, "/usr/lib/crti.o")
	}

	for _, library := range l.libraries {
		cmd.Args = append(cmd.Args, "-l"+library)
	}

	for _, input := range l.inputs {
		cmd.Args = append(cmd.Args, input)
	}

	if l.Crt {
		cmd.Args = append(cmd.Args, "/usr/lib/crtn.o")
	}

	cmd.Args = append(cmd.Args, "-o")
	cmd.Args = append(cmd.Args, output)

	// Execute
	out := bytes.Buffer{}
	cmd.Stderr = &out

	err := cmd.Run()

	if !cmd.ProcessState.Success() {
		return errors.New(out.String())
	}

	return err
}
