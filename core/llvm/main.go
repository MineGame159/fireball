package llvm

import (
	"bufio"
	"fireball/core/ir"
	"io"
)

func WriteText(m *ir.Module, writer io.Writer) {
	w := textWriter{
		m: m,
		w: bufio.NewWriter(writer),

		globalNameCounts: make(map[string]uint64),
		globalNames:      make(map[ir.Value]Name),

		localNameCounts: make(map[string]uint64),
		localNames:      make(map[ir.Value]Name),
	}

	w.writeString("source_filename = \"")
	w.writeString(m.Path)
	w.writeString("\"\n\n")

	w.writeStructs()
	w.writeGlobals()
	w.writeFunctions()
	w.writeMetadata()

	_ = w.w.Flush()
}
