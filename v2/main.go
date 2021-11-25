package main

import (
	"fmt"
	"github.com/alkazarix/monkey-lang/repl"
	"os"
)

func main() {
	fmt.Println("this is the monkey programming language!")
	fmt.Println("feel free to write some code:")
	repl.Start(os.Stdin, os.Stdout)
}
