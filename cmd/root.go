package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "talang",
	Short: "talang programing language. ",
	Long: `talang programing language.
	see more at: "https://github.com/alkazarix/talang"
	`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {}
