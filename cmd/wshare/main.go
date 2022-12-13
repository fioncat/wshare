package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "wshare",
	Short: "wshare manager",

	SilenceErrors: true,
	SilenceUsage:  true,

	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func main() {
	Root.AddCommand(Config, EditConfig)
	Root.AddCommand(Start, Status, Stop)
	Root.AddCommand(Logs)

	err := Root.Execute()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
