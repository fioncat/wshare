package main

import (
	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "wshare",
	Short: "wshare manager",

	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		return log.Init()
	},

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
		osutil.Exit(err)
	}
}
