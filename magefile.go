//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Default target to run when none is specified
var Default = Build

// Build builds the binary for the current platform (defaults to macOS)
func Build() error {
	fmt.Println("Building for macOS...")
	return buildFor("darwin", "amd64", "tg-dl")
}

// BuildLinux builds the binary for Linux (amd64)
func BuildLinux() error {
	fmt.Println("Building for Linux...")
	return buildFor("linux", "amd64", "tg-dl")
}

// BuildAll builds binaries for both macOS and Linux
func BuildAll() error {
	if err := Build(); err != nil {
		return fmt.Errorf("failed to build macOS binary: %w", err)
	}
	if err := BuildLinux(); err != nil {
		return fmt.Errorf("failed to build Linux binary: %w", err)
	}
	return nil
}

func buildFor(goos, goarch, output string) error {
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", goos),
		fmt.Sprintf("GOARCH=%s", goarch),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean removes all binaries and temporary files
func Clean() error {
	fmt.Println("Cleaning...")
	files := []string{
		"tg-dl",
		"session.json",
	}
	for _, file := range files {
		_ = os.Remove(file) // ignore errors if files don't exist
	}
	return nil
}

// Test runs the test suite
func Test() error {
	fmt.Println("Running tests...")
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Lint runs golangci-lint
func Lint() error {
	fmt.Println("Running linter...")
	cmd := exec.Command("golangci-lint", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install installs the binary to $GOPATH/bin
func Install() error {
	fmt.Println("Installing...")
	cmd := exec.Command("go", "install", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run builds and runs the application
func Run() error {
	if err := Build(); err != nil {
		return err
	}

	fmt.Println("Running tg-dl...")
	cmd := exec.Command("./tg-dl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
