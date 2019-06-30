//+build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

var (
	binPath = "bin/mail2most"
)

// Build - mage build
func Build() error {
	err := Clean()
	if err != nil {
		return err
	}

	err = Test()
	if err != nil {
		return err
	}

	return sh.RunV("go", "build", "-o", binPath)
}

// Test - running tests and code coverage
func Test() error {
	return sh.RunV("go", "test", "-v", "-cover", "./...", "-coverprofile=coverage.out")
}

// Run - mage run
func Run() error {
	return sh.RunV("go", "run", "main.go")
}

// Coverage - checking code coverage
func Coverage() error {
	if _, err := os.Stat("./coverage.out"); err != nil {
		return fmt.Errorf("run mage test befor checking the code coverage")
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out")
}

// Clean cleans up the client generation and binarys
func Clean() error {
	fmt.Println("cleaning up")
	if _, err := os.Stat("coverage.out"); err == nil {
		err = os.Remove("coverage.out")
		if err != nil {
			return err
		}
	}
	return nil
}
