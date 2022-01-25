package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alkazarix/talang/interpreter"
	"github.com/alkazarix/talang/lexer"
	"github.com/alkazarix/talang/parser"
	"github.com/spf13/cobra"
)

var sourceFile string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run `talang` program file.",
	Long:  "run `talang` program file.",
	Run: func(cmd *cobra.Command, args []string) {
		source, err := ioutil.ReadFile(sourceFile)
		if err != nil {
			panic(err)
		}

		l := lexer.New(string(source))
		lexemes := l.Lexeme()

		p := parser.New(lexemes)

		program, err := p.Parse()
		if err != nil {
			printError(err)
			return
		}

		interpretor := interpreter.New()
		result, err := interpretor.Evaluate(&program)
		if err != nil {
			printError(err)
			return
		}

		if result != nil {
			fmt.Println(result.Inspect())
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&sourceFile, "file", "f", "myprogram.tal", "path to the `talang` source file to run (required)")
	runCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(runCmd)
}

func printError(err error) {
	fmt.Fprintf(os.Stderr, "Oops! something wrong append here!\n")
	fmt.Fprintf(os.Stderr, "\t"+err.Error()+"\n")
}
