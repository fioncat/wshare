package main

import (
	"fmt"
	"os"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share/server"
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

	err = server.Start(addr)
	if err != nil {
		errorExit(err)
	}
}

func errorExit(err error) {
	fmt.Printf("failed to start server: %v\n", err)
	os.Exit(1)
}
