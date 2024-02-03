package llvm

import (
	"bufio"
	"fireball/core/ir"
	"strconv"
)

type textWriter struct {
	m *ir.Module

	// Writing

	w      *bufio.Writer
	buffer [65]byte

	isArgument bool

	// Names

	globalUnnamedCount int64
	globalNameCounts   map[string]uint64
	globalNames        map[ir.Value]Name

	localUnnamedCount int64
	localNameCounts   map[string]uint64
	localNames        map[ir.Value]Name

	skipNameChar bool
}

// Write

func (w *textWriter) writeString(s string) {
	_, _ = w.w.WriteString(s)
}

func (w *textWriter) writeRune(r rune) {
	_, _ = w.w.WriteRune(r)
}

func (w *textWriter) writeByte(b byte) {
	_ = w.w.WriteByte(b)
}

func (w *textWriter) writeBool(b bool) {
	if b {
		w.writeString("true")
	} else {
		w.writeString("false")
	}
}

func (w *textWriter) writeQuotedString(s string) {
	w.writeRune('"')
	w.writeString(s)
	w.writeRune('"')
}

func (w *textWriter) writeInt(v int64) {
	buffer := strconv.AppendInt(w.buffer[0:0], v, 10)
	_, _ = w.w.Write(buffer)
}

func (w *textWriter) writeUint(v uint64, base int) {
	buffer := strconv.AppendUint(w.buffer[0:0], v, base)
	_, _ = w.w.Write(buffer)
}
