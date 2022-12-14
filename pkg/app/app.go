package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/crypto"
	"github.com/fioncat/wshare/pkg/daemon"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/spf13/cobra"
)

func CreateManager(name, full string, start func() error) *cobra.Command {
	var d *daemon.Daemon

	startCmd := &cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("Start %s", name),

		RunE: func(_ *cobra.Command, _ []string) error {
			return d.Start(start)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: fmt.Sprintf("Show %s status", name),

		RunE: func(cmd *cobra.Command, args []string) error {
			return d.ShowStatus()
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop %s", name),

		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Stop()
		},
	}

	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: fmt.Sprintf("Restart %s", name),

		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Restart(start)
		},
	}

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: fmt.Sprintf("Show %s log", name),

		DisableFlagParsing: true,

		RunE: func(_ *cobra.Command, args []string) error {
			path := d.LogPath()
			args = append(args, path)
			cmd := exec.Command("tail", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			return cmd.Run()
		},
	}

	root := &cobra.Command{
		Use:   full,
		Short: fmt.Sprintf("%s manager", full),

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

			d, err = daemon.New(name)
			if err != nil {
				return fmt.Errorf("failed to init daemon: %v", err)
			}
			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,

		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	root.AddCommand(startCmd, stopCmd, restartCmd, logsCmd, statusCmd)
	return root
}
