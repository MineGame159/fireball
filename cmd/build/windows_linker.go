package build

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type windowsLinker struct {
	ucrtPath string
	umPath   string
	msvcPath string

	libraries []string
	inputs    []string
}

func (w *windowsLinker) Check() error {
	if err := w.findKit(); err != nil {
		return err
	}
	if err := w.findMsvc(); err != nil {
		return err
	}

	w.libraries = append(w.libraries, "libucrt.lib")
	w.libraries = append(w.libraries, "libcmt.lib")

	return nil
}

func (w *windowsLinker) findKit() error {
	base := "C:\\Program Files (x86)\\Windows Kits\\10\\Lib"

	entries, err := os.ReadDir(base)
	if err != nil {
		return errors.New("failed to find a valid windows development kit")
	}

	w.ucrtPath = filepath.Join(base, entries[0].Name(), "ucrt", "x64")
	_, err = os.Stat(w.ucrtPath)
	if err != nil {
		return errors.New("failed to find a valid windows development kit, ucrt")
	}

	w.umPath = filepath.Join(base, entries[0].Name(), "um", "x64")
	_, err = os.Stat(w.umPath)
	if err != nil {
		return errors.New("failed to find a valid windows development kit, um")
	}

	return nil
}

func (w *windowsLinker) findMsvc() error {
	base := "C:\\Program Files (x86)\\Microsoft Visual Studio\\2022\\BuildTools\\VC\\Tools\\MSVC"

	entries, err := os.ReadDir(base)
	if err != nil {
		return errors.New("failed to find a MSVC installation")
	}

	w.msvcPath = filepath.Join(base, entries[0].Name(), "lib", "x64")
	_, err = os.Stat(w.msvcPath)
	if err != nil {
		return errors.New("failed to find a MSVC installation, lib")
	}

	return nil
}

func (w *windowsLinker) AddLibrary(library string) {
	w.libraries = append(w.libraries, library)
}

func (w *windowsLinker) AddInput(input string) {
	w.inputs = append(w.inputs, input)
}

func (w *windowsLinker) Link(output string) error {
	cmd := exec.Command("lld-link")

	cmd.Args = append(cmd.Args, fmt.Sprintf("/libpath:%s", w.ucrtPath))
	cmd.Args = append(cmd.Args, fmt.Sprintf("/libpath:%s", w.umPath))
	cmd.Args = append(cmd.Args, fmt.Sprintf("/libpath:%s", w.msvcPath))

	for _, library := range w.libraries {
		cmd.Args = append(cmd.Args, library)
	}

	for _, input := range w.inputs {
		cmd.Args = append(cmd.Args, fmt.Sprintf("%s", input))
	}

	cmd.Args = append(cmd.Args, "/machine:x64")
	cmd.Args = append(cmd.Args, "/subsystem:console")
	cmd.Args = append(cmd.Args, "/debug")
	cmd.Args = append(cmd.Args, fmt.Sprintf("/out:%s", output))

	return execute(cmd)
}
