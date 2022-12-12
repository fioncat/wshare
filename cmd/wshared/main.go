package main

import (
	"fmt"
	"os"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share"
	"github.com/fioncat/wshare/share/client"
	"github.com/fioncat/wshare/share/handler/clipboard"
)

func main() {
	err := config.Init()
	if err != nil {
		errorExit(err)
	}
	err = log.Init()
	if err != nil {
		errorExit(err)
	}
	client, err := client.New()
	if err != nil {
		errorExit(err)
	}

	share.RegisterHandler("clipboard", clipboard.New)

	client.Start()
}

func errorExit(err error) {
	fmt.Fprintf(os.Stderr, "start wshared with error: %v", err)
	os.Exit(1)
}
