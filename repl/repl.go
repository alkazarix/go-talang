package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/alkazarix/talang/interpreter"
	"github.com/alkazarix/talang/lexer"
	"github.com/alkazarix/talang/parser"
)

const (
	PROMPT = ">> "
	EXIT   = "exit"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	interpreter := interpreter.New()
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if line == EXIT {
			return
		}

		lexer := lexer.New(line)
		lexemes := lexer.Lexeme()

		parser := parser.New(lexemes)
		program, err := parser.Parse()
		if err != nil {
			printError(out, err)
			continue
		}

		evaluated, err := interpreter.Evaluate(&program)
		if err != nil {
			printError(out, err)
			continue
		}

		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printError(out io.Writer, err error) {
	io.WriteString(out, "Oops! something wrong append here!\n")
	io.WriteString(out, "\t"+err.Error()+"\n")
}
