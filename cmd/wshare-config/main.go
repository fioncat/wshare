package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/spf13/cobra"
)

var editors = []string{
	"nvim", "vim", "vi",
}

var Root = &cobra.Command{
	Use:   "wshare-config",
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

var EditConfig = &cobra.Command{
	Use:   "edit",
	Short: "Open editor to edit config",

	RunE: func(_ *cobra.Command, _ []string) error {
		var useEditor string
		for _, editor := range editors {
			if osutil.CommandExists(editor) {
				useEditor = editor
				break
			}
		}
		if useEditor == "" {
			return fmt.Errorf("cannot find editor within %v", editors)
		}
		err := config.Init()
		if err != nil {
			return err
		}

		path := config.Path()
		exists, err := osutil.FileExists(path)
		if err != nil {
			return fmt.Errorf("failed to check config exists: %v", err)
		}
		if !exists {
			dir := filepath.Dir(path)
			err = osutil.EnsureDir(dir)
			if err != nil {
				return fmt.Errorf("failed to ensure dir: %v", err)
			}

			err = os.WriteFile(path, config.Default, 0644)
			if err != nil {
				return fmt.Errorf("failed to write default config: %v", err)
			}
		}

		cmd := exec.Command(useEditor, config.Path())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	},
}

func main() {
	Root.AddCommand(EditConfig)

	err := Root.Execute()
	if err != nil {
		osutil.Exit(err)
	}
}
