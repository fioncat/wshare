package main

import (
	"errors"
	"fmt"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/crypto"
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
			return fmt.Errorf("failed to init config: %v", err)
		}
		err = log.Init()
		if err != nil {
			return fmt.Errorf("failed to init log: %v", err)
		}

		password := config.Get().Password
		if password == "" {
			return errors.New("password cannot be empty")
		}
		err = crypto.Init(password)
		if err != nil {
			return fmt.Errorf("failed to init password: %v", err)
		}
		return nil
	},

	SilenceErrors: true,
	SilenceUsage:  true,

	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func main() {
	Root.AddCommand(Config, EditConfig)
	Root.AddCommand(Start, Status, Stop, Restart)
	Root.AddCommand(Logs)

	err := Root.Execute()
	if err != nil {
		osutil.Exit(err)
	}
}
