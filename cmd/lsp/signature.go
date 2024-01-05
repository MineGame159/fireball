package lsp

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fmt"
	"github.com/MineGame159/protocol"
	"strings"
)

func getSignatureHelp(node ast.Node, pos core.Pos) *protocol.SignatureHelp {
	// Get node under position
	node = getNodeUnderPos(node, pos)
	if node == nil {
		return nil
	}

	// Get call or type call signature
	signature, ok := getSignature(node, pos)
	if !ok {
		return nil
	}

	return &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{signature},
		ActiveParameter: signature.ActiveParameter,
		ActiveSignature: 0,
	}
}

func getSignature(node ast.Node, pos core.Pos) (protocol.SignatureInformation, bool) {
	for node != nil {
		switch node := node.(type) {
		case *ast.Call:
			signature, ok := getCallSignature(pos, node)
			if ok {
				return signature, true
			}

			if node.Parent() != nil {
				return getSignature(node.Parent(), pos)
			}

		case *ast.TypeCall:
			signature, ok := getTypeCallSignature(pos, node)
			if ok {
				return signature, true
			}

			if node.Parent() != nil {
				return getSignature(node.Parent(), pos)
			}
		}

		node = node.Parent()
	}

	return protocol.SignatureInformation{}, false
}

func getCallSignature(pos core.Pos, call *ast.Call) (protocol.SignatureInformation, bool) {
	// Check position
	if !isBetween(pos, call, scanner.LeftParen, scanner.RightParen) {
		return protocol.SignatureInformation{}, false
	}

	// Get function
	var function *ast.Func

	if f, ok := ast.As[*ast.Func](call.Callee.Result().Type); ok {
		function = f
	} else {
		return protocol.SignatureInformation{}, false
	}

	// Get label and parameters
	label := strings.Builder{}
	parameters := make([]protocol.ParameterInformation, 0, len(function.Params)+1)
	activeParameter := -1

	if function.Name != nil {
		label.WriteString(function.Name.String())
	}

	label.WriteRune('(')

	for i, param := range function.Params {
		// Label
		paramLabel := fmt.Sprintf("%s %s", param.Name, ast.PrintType(param.Type))

		if len(parameters) > 0 {
			label.WriteString(", ")
		}

		label.WriteString(paramLabel)
		parameters = append(parameters, protocol.ParameterInformation{Label: paramLabel})

		// Active parameter
		if i < len(call.Args) && call.Args[i].Cst() != nil {
			range_ := call.Args[i].Cst().Range

			if i > 0 && call.Args[i-1].Cst() != nil {
				range_.Start = call.Args[i-1].Cst().Range.End
				range_.Start.Column++
			}

			if range_.Contains(pos) {
				activeParameter = len(parameters) - 1
			}
		}
	}

	if function.IsVariadic() {
		if len(parameters) > 0 {
			label.WriteString(", ")
		}

		label.WriteString("...")
		parameters = append(parameters, protocol.ParameterInformation{Label: "..."})

		if activeParameter == -1 {
			activeParameter = len(parameters) - 1
		}
	}

	if activeParameter == -1 && (len(call.Args) == 0 || pos.IsAfter(call.Args[len(call.Args)-1].Cst().Range.End)) {
		activeParameter = min(max(0, len(call.Args)), len(parameters))
	}

	label.WriteRune(')')

	if !ast.IsPrimitive(function.Returns, ast.Void) {
		label.WriteRune(' ')
		label.WriteString(ast.PrintType(function.Returns))
	}

	// Return
	return protocol.SignatureInformation{
		Label:           label.String(),
		Parameters:      parameters,
		ActiveParameter: uint32(activeParameter),
	}, true
}

func getTypeCallSignature(pos core.Pos, call *ast.TypeCall) (protocol.SignatureInformation, bool) {
	// Check position
	if !isBetween(pos, call, scanner.LeftParen, scanner.RightParen) || call.Callee == nil {
		return protocol.SignatureInformation{}, false
	}

	// Return
	return protocol.SignatureInformation{
		Label: call.Callee.String() + "(<type>) u32",
		Parameters: []protocol.ParameterInformation{{
			Label: "<type>",
		}},
		ActiveParameter: 0,
	}, true
}
