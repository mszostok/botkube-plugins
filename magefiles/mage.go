//go:build ignore

package main

import (
	"os"

	"github.com/magefile/mage/mage"
)

// This file allows someone to run mage commands without mage installed
// by running:
//
//	go run -modfile magefiles/go.mod magefiles/mage.go TARGET
//
// You can also install it via:
//
//	go install -modfile ./magefiles/go.mod ./magefiles/mage.go
//
// And use it as:
//
//	mage TARGET
func main() { os.Exit(mage.Main()) }
