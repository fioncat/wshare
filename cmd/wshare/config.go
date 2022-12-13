package main

import (
	"encoding/json"
	"fmt"

	"github.com/fioncat/wshare/config"
	"github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "Show config values in json style",

	RunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		if err != nil {
			return err
		}
		cfg := config.Get()
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}
		fmt.Printf("Use config: %s\n", config.Path())
		fmt.Println(string(data))
		return nil
	},
}
