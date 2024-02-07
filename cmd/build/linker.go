package build

import "runtime"

type Linker interface {
	Check() error

	AddLibrary(library string)
	AddInput(input string)

	Link(output string) error
}

func GetLinker() Linker {
	switch runtime.GOOS {
	case "windows":
		return &windowsLinker{}
	case "linux":
		return &linuxLinker{}

	default:
		panic("Operating system not implemented")
	}
}
