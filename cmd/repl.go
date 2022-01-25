package cmd

import (
	"fmt"
	"os"

	"github.com/alkazarix/talang/repl"
	"github.com/spf13/cobra"
)

const (
	LOGO = `
	▄▀▀▀█▀▀▄  ▄▀▀█▄   ▄▀▀▀▀▄      ▄▀▀█▄   ▄▀▀▄ ▀▄  ▄▀▀▀▀▄   
	█    █  ▐ ▐ ▄▀ ▀▄ █    █      ▐ ▄▀ ▀▄ █  █ █ █ █         
	▐   █       █▄▄▄█ ▐    █        █▄▄▄█ ▐  █  ▀█ █    ▀▄▄  
		 █       ▄▀   █     █        ▄▀   █   █   █  █     █ █ 
	 ▄▀       █   ▄▀    ▄▀▄▄▄▄▄▄▀ █   ▄▀  ▄▀   █   ▐▀▄▄▄▄▀ ▐ 
	█         ▐   ▐     █         ▐   ▐   █    ▐   ▐         
	▐                   ▐                 ▐                  `
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "launch the `repl` for talang program.",
	Long:  "launch the `repl` for talang program.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf(LOGO)
		fmt.Printf("\n\n")

		fmt.Printf("Welcome to `talang` programming language !\n")
		fmt.Printf("start typing code.\n")
		fmt.Printf("enter `exit` to quit the repl.\n\n")

		repl.Start(os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}
