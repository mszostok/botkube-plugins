package osx

import (
	"fmt"
	"os"
)

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
