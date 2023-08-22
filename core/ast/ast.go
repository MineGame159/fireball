package ast

import "fireball/core/scanner"

type Node interface {
	Token() scanner.Token
}
