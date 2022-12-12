package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/fioncat/wshare/share"
	"github.com/fioncat/wshare/share/client"
	"github.com/fioncat/wshare/share/handler/clipboard"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

func localPath() (string, error) {
	localPath := filepath.Join(config.HomeDir(), ".local", "share", "wshare")
	err := osutil.EnsureDir(localPath)
	return localPath, err
}

func pidPath() (string, error) {
	local, err := localPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(local, "daemon.pid"), nil
}

func getPid() (int, error) {
	path, err := pidPath()
	if err != nil {
		return 0, err
	}
	exists, err := osutil.FileExists(path)
	if err != nil {
		return 0, err
	}
	if !exists {
		return -1, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid daemon.pid content: %q", string(data))
	}
	return pid, nil
}

var Start = &cobra.Command{
	Use:   "start",
	Short: "Start wshare daemon",

	RunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		err = log.Init()
		if err != nil {
			return err
		}

		pidPath, err := pidPath()
		if err != nil {
			return err
		}

		dctx := &daemon.Context{
			PidFileName: pidPath,
			PidFilePerm: 0644,
			LogFileName: config.Get().Log.Path,
			LogFilePerm: 0640,
			WorkDir:     "./",
			Umask:       027,
		}

		d, err := dctx.Reborn()
		if err != nil {
			return err
		}
		if d != nil {
			return nil
		}
		defer dctx.Release()

		err = share.InitHandlers()
		if err != nil {
			return err
		}

		client, err := client.New()
		if err != nil {
			return err
		}

		share.RegisterHandler("clipboard", clipboard.New)

		client.Start()
		return nil
	},
}
