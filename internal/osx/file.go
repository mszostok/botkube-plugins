package osx

import (
	"fmt"
	"os"
)

type TmpDir string

func (t TmpDir) Get() (string, bool) {
	if t != "" {
		return string(t), true
	}

	depDir := os.Getenv("PLUGIN_DEPENDENCY_DIR")
	if depDir != "" {
		return depDir, false
	}

	return "/tmp/bin", true
}

func (t TmpDir) GetDirectory() string {
	dir, _ := t.Get()
	return dir
}

func RunIfFileDoesNotExist(path string, fn func() error) error {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		fmt.Println("already downloaded")
	case os.IsNotExist(err):
		return fn()
	default:
		return err
	}
	return nil
}
