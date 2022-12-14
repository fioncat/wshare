package main

import (
	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/app"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/fioncat/wshare/share/server"
)

func startServer() error {
	addr := config.Get().Listen
	return server.Start(addr)
}

func main() {
	cmd := app.CreateManager("server", "wshare-server", startServer)
	err := cmd.Execute()
	if err != nil {
		osutil.Exit(err)
	}
}
