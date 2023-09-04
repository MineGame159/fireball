package lsp

import "github.com/MineGame159/protocol"

type document struct {
	hasDiagnostics bool

	semanticTokens *protocol.SemanticTokens
	inlayHints     []protocol.InlayHint
}

func (d *document) setDirty() {
	d.semanticTokens = nil
	d.inlayHints = nil
}
