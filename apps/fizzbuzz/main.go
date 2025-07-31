package main

import (
	"strconv"
	"strings"

	"github.com/sunfish-shogi/go-monorepo-expt/pkgs/fmt"
)

func main() {
	sb := strings.Builder{}
	for i := 1; i <= 15; i++ {
		if i%3 == 0 && i%5 == 0 {
			sb.WriteString("FizzBuzz\n")
		} else if i%3 == 0 {
			sb.WriteString("Fizz\n")
		} else if i%5 == 0 {
			sb.WriteString("Buzz\n")
		} else {
			sb.WriteString(strconv.Itoa(i) + "\n")
		}
	}
	print(fmt.Highlight(strings.TrimSpace(sb.String())))
}
