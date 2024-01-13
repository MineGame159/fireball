package llvm

import (
	"fireball/core/ir"
	"regexp"
)

type Name struct {
	char rune

	unnamed int64

	named       string
	namedSuffix uint64
}

func (w *textWriter) writeValue(v ir.Value) {
	switch v := v.(type) {
	case *ir.Block:
		w.writeString("label")
	default:
		w.writeType(v.Type())
	}

	w.writeRune(' ')

	w.writeValueValue(v)
}

func (w *textWriter) writeValueValue(v ir.Value) {
	switch v := v.(type) {
	case ir.Const:
		w.writeConst(v)
	case ir.MetaID:
		w.writeMetaRef(v)
	default:
		w.writeName(v)
	}
}

func (w *textWriter) writeName(v ir.Value) {
	switch v := v.(type) {
	case *ir.GlobalVar, *ir.Func:
		w.writeName_(v, true)
	case *ir.Param, *ir.Block, ir.Inst:
		w.writeName_(v, false)
	}
}

func (w *textWriter) writeName_(v ir.Value, global bool) {
	w.cacheName(v, global)

	if global {
		w.writeNameImpl(w.globalNames[v])
	} else {
		w.writeNameImpl(w.localNames[v])
	}
}

func (w *textWriter) cacheName(v ir.Value, global bool) {
	// Global
	if global {
		// Cache
		if _, ok := w.globalNames[v]; ok {
			return
		}

		// Unnamed
		if v.Name() == "" {
			name := Name{
				char:    '@',
				unnamed: w.globalUnnamedCount,
			}

			w.globalUnnamedCount++
			w.globalNames[v] = name

			return
		}

		// Named
		suffix := uint64(0)

		if count, ok := w.globalNameCounts[v.Name()]; ok {
			suffix = count
			w.globalNameCounts[v.Name()]++
		} else {
			w.globalNameCounts[v.Name()] = 1
		}

		name := Name{
			char:        '@',
			unnamed:     -1,
			named:       surroundName(v.Name()),
			namedSuffix: suffix,
		}

		w.globalNames[v] = name

		return
	}

	// Local
	if _, ok := w.localNames[v]; ok {
		return
	}

	// Unnamed
	if v.Name() == "" {
		name := Name{
			char:    '%',
			unnamed: w.localUnnamedCount,
		}

		w.localUnnamedCount++
		w.localNames[v] = name

		return
	}

	// Named
	suffix := uint64(0)

	if count, ok := w.localNameCounts[v.Name()]; ok {
		suffix = count
		w.localNameCounts[v.Name()]++
	} else {
		w.localNameCounts[v.Name()] = 1
	}

	name := Name{
		char:        '%',
		unnamed:     -1,
		named:       surroundName(v.Name()),
		namedSuffix: suffix,
	}

	w.localNames[v] = name
}

func (w *textWriter) writeNameImpl(name Name) {
	if !w.skipNameChar {
		w.writeRune(name.char)
	}

	if name.unnamed >= 0 {
		w.writeInt(name.unnamed)
	} else {
		w.writeString(name.named)

		if name.namedSuffix > 0 {
			w.writeUint(name.namedSuffix, 10)
		}
	}
}

func (w *textWriter) resetLocalNames() {
	w.localUnnamedCount = 0
	clear(w.localNameCounts)
	clear(w.localNames)
}

var namePattern = regexp.MustCompile("^[-a-zA-Z$._][-a-zA-Z$._0-9]*$")

func surroundName(name string) string {
	if namePattern.MatchString(name) {
		return name
	}

	return "\"" + name + "\""
}
