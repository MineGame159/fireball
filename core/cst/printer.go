package cst

import "fmt"

func Print(node Node) {
	printNode(node, 0)
}

func printNode(node Node, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}

	if node.Leaf() {
		fmt.Printf("- '%s' (%s)", node.Token.Lexeme, node.Kind)
		fmt.Printf("                [(%d, %d) - (%d, %d)]\n", node.Range.Start.Line, node.Range.Start.Column, node.Range.End.Line, node.Range.End.Column)
	} else {
		fmt.Printf("+ %s", node.Kind)
		fmt.Printf("                [(%d, %d) - (%d, %d)]\n", node.Range.Start.Line, node.Range.Start.Column, node.Range.End.Line, node.Range.End.Column)

		for _, child := range node.Children {
			printNode(child, depth+1)
		}
	}
}
