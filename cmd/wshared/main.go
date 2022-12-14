package main

import (
	"github.com/fioncat/wshare/pkg/app"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/fioncat/wshare/share"
	"github.com/fioncat/wshare/share/client"
	"github.com/fioncat/wshare/share/handler/clipboard"
)

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

func main() {
	cmd := app.CreateManager("wshared", "wshared", startClient)
	err := cmd.Execute()
	if err != nil {
		osutil.Exit(err)
	}
}
