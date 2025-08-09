package main

import (
	_ "embed"

	"github.com/sunfish-shogi/go-monorepo-expt/pkgs/fmt"
)

//go:embed hello.txt
var hello string

func main() {
	print(fmt.Highlight(hello))
}
