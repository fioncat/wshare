package osutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func EnsureDir(dir string) error {
	exists, err := DirExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

func DirExists(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, fmt.Errorf("%s is not a directory", dir)
	}
	return true, nil
}

func FileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if stat.IsDir() {
		return false, fmt.Errorf("%s is not a file", path)
	}
	return true, nil
}

// OpenAppend opens a writer to append values. This is always used to
// open a log file. The file won't be closed.
func OpenAppend(path string) (io.Writer, error) {
	dir := filepath.Dir(path)
	err := EnsureDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure dir: %v", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func CommandExists(name string) bool {
	args := []string{
		"-c",
		fmt.Sprintf("command -v %s", name),
	}
	cmd := exec.Command("bash", args...)
	return cmd.Run() == nil
}
