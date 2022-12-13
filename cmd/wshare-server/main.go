package main

import (
	"fmt"
	"os"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share/server"
	"github.com/sevlyar/go-daemon"
)

func main() {
	addr := ":6679"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}

	err := config.Init()
	if err != nil {
		errorExit(err)
	}

	err = log.Init()
	if err != nil {
		errorExit(err)
	}

	pidPath, err := config.LocalFile("server.pid")
	if err != nil {
		errorExit(err)
	}
	logPath, err := config.LocalFile("server.log")
	if err != nil {
		errorExit(err)
	}

	dctx := &daemon.Context{
		PidFileName: pidPath,
		PidFilePerm: 0644,
		LogFileName: logPath,
		LogFilePerm: 0644,

		Umask: 027,
	}

	d, err := dctx.Reborn()
	if err != nil {
		if err == daemon.ErrWouldBlock {
			fmt.Printf("server is still running, log: %s\n", logPath)
			return
		}

		errorExit(err)
	}
	if d != nil {
		return
	}
	defer dctx.Release()

	err = server.Start(addr)
	if err != nil {
		errorExit(err)
	}
}

func errorExit(err error) {
	fmt.Printf("failed to start server: %v\n", err)
	os.Exit(1)
}
