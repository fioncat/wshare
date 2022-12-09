package main

import (
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

var EditConfig = &cobra.Command{
	Use:   "edit-config",
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
