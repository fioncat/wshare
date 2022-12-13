package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fioncat/wshare/config"
	"github.com/sevlyar/go-daemon"
)

type Daemon struct {
	name string

	pid int

	pidPath string
	logPath string
}

func New(name string) (*Daemon, error) {
	pidName := fmt.Sprintf("%s.pid", name)
	logName := fmt.Sprintf("%s.log", name)
	pidPath, err := config.LocalFile(pidName)
	if err != nil {
		return nil, err
	}

	logPath, err := config.LocalFile(logName)
	if err != nil {
		return nil, err
	}

	pid, err := getPid(pidPath)
	if err != nil {
		return nil, err
	}

	return &Daemon{
		name:    name,
		pid:     pid,
		pidPath: pidPath,
		logPath: logPath,
	}, nil
}

func getPid(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		}
		return 0, err
	}

	if len(data) == 0 {
		return -1, nil
	}

	str := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("WARN: invalid pid %s: %q\n", path, str)
		return -1, nil
	}
	return pid, nil
}

func (d *Daemon) Start(f func() error) error {
	dctx := &daemon.Context{
		PidFileName: d.pidPath,
		PidFilePerm: 0644,
		LogFileName: d.logPath,
		LogFilePerm: 0640,
		Umask:       027,
	}

	rd, err := dctx.Reborn()
	if err != nil {
		if err == daemon.ErrWouldBlock {
			return nil
		}
		return err
	}
	if rd != nil {
		return nil
	}
	defer dctx.Release()

	return f()
}

func (d *Daemon) GetProcess() (*os.Process, error) {
	if d.pid < 0 {
		return nil, nil
	}
	return os.FindProcess(d.pid)
}

func (d *Daemon) Stop() error {
	process, err := d.GetProcess()
	if err != nil {
		return err
	}
	if process == nil {
		return nil
	}
	if isRunning(process) {
		fmt.Printf("killing %d...\n", d.pid)
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
	return os.Remove(d.pidPath)
}

func (d *Daemon) ShowStatus() error {
	process, err := d.GetProcess()
	if err != nil {
		return fmt.Errorf("failed to get process: %v", err)
	}
	if process == nil {
		fmt.Printf("%s dead\n", d.name)
		return nil
	}
	if isRunning(process) {
		attr := color.New(color.FgGreen, color.Bold)
		status := attr.Sprint("running")
		fmt.Printf("%s %d, %s\n", d.name, d.pid, status)
		return nil
	}
	attr := color.New(color.FgRed, color.Bold)
	status := attr.Sprint("not running")
	fmt.Printf("%s %d, %s\n", d.name, d.pid, status)

	return nil
}

func (d *Daemon) Restart(f func() error) error {
	err := d.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop daemon: %s", err)
	}
	return d.Start(f)
}

func isRunning(p *os.Process) bool {
	err := p.Signal(syscall.Signal(0))
	return err == nil
}

func (d *Daemon) LogPath() string {
	return d.logPath
}
