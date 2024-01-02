package main

import (
	"fmt"
	"os"
	"strings"
)

type Writer struct {
	str   strings.Builder
	depth int
}

func newWriter() *Writer {
	return &Writer{
		str:   strings.Builder{},
		depth: 0,
	}
}

func (w *Writer) flush(file string) {
	_ = os.WriteFile(file, []byte(w.str.String()), 0666)
}

func (w *Writer) write(format string, args ...any) {
	str := fmt.Sprintf(format, args...)

	if strings.HasPrefix(str, "}") {
		w.depth--
	}

	for i := 0; i < w.depth; i++ {
		w.str.WriteRune('\t')
	}

	w.str.WriteString(str)
	w.str.WriteRune('\n')

	if strings.HasSuffix(str, "{") {
		w.depth++
	}
}
