package main

import (
	"os"
	"os/exec"

	"github.com/fioncat/wshare/pkg/daemon"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/fioncat/wshare/share"
	"github.com/fioncat/wshare/share/client"
	"github.com/fioncat/wshare/share/handler/clipboard"
	"github.com/spf13/cobra"
)

func getDaemon() *daemon.Daemon {
	d, err := daemon.New("wshared")
	if err != nil {
		osutil.Exit(err)
	}
	return d
}

func startClient() error {
	share.RegisterHandler("clipboard", clipboard.New)
	err := share.InitHandlers()
	if err != nil {
		return err
	}

	client, err := client.New()
	if err != nil {
		return err
	}

	client.Start()
	return nil
}

var Start = &cobra.Command{
	Use:   "start",
	Short: "Start wshare daemon",

	RunE: func(_ *cobra.Command, _ []string) error {
		return getDaemon().Start(startClient)
	},
}

var Status = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",

	RunE: func(_ *cobra.Command, _ []string) error {
		return getDaemon().ShowStatus()
	},
}

var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Stop daemon",

	RunE: func(_ *cobra.Command, _ []string) error {
		return getDaemon().Stop()
	},
}

var Logs = &cobra.Command{
	Use:   "logs",
	Short: "Show daemon logs",

	DisableFlagParsing: true,

	RunE: func(_ *cobra.Command, args []string) error {
		path := getDaemon().LogPath()
		args = append(args, path)
		cmd := exec.Command("tail", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	},
}
