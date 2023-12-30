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
		fmt.Printf("- '%s' (%s)\n", node.Token.Lexeme, node.Kind)
	} else {
		fmt.Printf("+ %s\n", node.Kind)

		for _, child := range node.Children {
			printNode(child, depth+1)
		}
	}
}
