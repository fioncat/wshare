package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
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

	if len(data) == 0 {
		return -1, nil
	}

	str := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid daemon.pid content: %q", str)
	}
	return pid, nil
}

func isRunning(p *os.Process) bool {
	err := p.Signal(syscall.Signal(0))
	return err == nil
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
			if err == daemon.ErrWouldBlock {
				return nil
			}
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

var Status = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",

	RunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		pid, err := getPid()
		if err != nil {
			return err
		}
		if pid < 0 {
			fmt.Println("wshared is dead")
			return nil
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		var status string
		if isRunning(process) {
			attr := color.New(color.FgGreen, color.Bold)
			status = attr.Sprint("running")
		} else {
			attr := color.New(color.FgRed, color.Bold)
			status = attr.Sprint("not running")
		}

		fmt.Printf("wshared pid %d, %s\n", pid, status)
		return nil
	},
}

var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Stop daemon",

	RunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		pid, err := getPid()
		if err != nil {
			return err
		}
		if pid < 0 {
			return nil
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		if isRunning(process) {
			fmt.Printf("kill %d...\n", pid)
			err = process.Kill()
			if err != nil {
				return fmt.Errorf("failed to kill process: %v", err)
			}
			time.Sleep(time.Second * 2)
			if isRunning(process) {
				return fmt.Errorf("process is still running after killing, " +
					"please try to kill it manually")
			}
		}

		path, err := pidPath()
		if err != nil {
			return err
		}
		return os.Remove(path)
	},
}

var Logs = &cobra.Command{
	Use:   "logs",
	Short: "Show daemon logs",

	DisableFlagParsing: true,

	RunE: func(_ *cobra.Command, args []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		path := config.Get().Log.Path
		args = append(args, path)
		cmd := exec.Command("tail", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	},
}
