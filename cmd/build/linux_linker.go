package build

import (
	"os/exec"
)

type linuxLinker struct {
	libraries []string
	inputs    []string
}

func (l *linuxLinker) Check() error {
	l.libraries = append(l.libraries, "c")
	l.libraries = append(l.libraries, "m")

	return nil
}

func (l *linuxLinker) AddLibrary(library string) {
	l.libraries = append(l.libraries, library)
}

func (l *linuxLinker) AddInput(input string) {
	l.inputs = append(l.inputs, input)
}

func (l *linuxLinker) Link(output string) error {
	cmd := exec.Command("ld.lld")

	cmd.Args = append(cmd.Args, "-L/usr/lib")

	cmd.Args = append(cmd.Args, "-dynamic-linker")
	cmd.Args = append(cmd.Args, "/lib64/ld-linux-x86-64.so.2")

	cmd.Args = append(cmd.Args, "/usr/lib/crt1.o")
	cmd.Args = append(cmd.Args, "/usr/lib/crti.o")

	for _, library := range l.libraries {
		cmd.Args = append(cmd.Args, "-l"+library)
	}

	for _, input := range l.inputs {
		cmd.Args = append(cmd.Args, withExtension(input, "o"))
	}

	cmd.Args = append(cmd.Args, "/usr/lib/crtn.o")

	cmd.Args = append(cmd.Args, "-o")
	cmd.Args = append(cmd.Args, output)

	return execute(cmd)
}
