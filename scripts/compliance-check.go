// Constitutional compliance checker for mesh probe system
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := runComplianceCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "Compliance check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All constitutional compliance checks passed!")
}

func runComplianceCheck() error {
	// Check if we're in a Go module
	if !isGoModule() {
		return fmt.Errorf("not in a Go module")
	}

	// Run go vet
	if err := runCommand("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %v", err)
	}

	// Run golangci-lint if available
	if isToolAvailable("golangci-lint") {
		if err := runCommand("golangci-lint", "run"); err != nil {
			return fmt.Errorf("golangci-lint failed: %v", err)
		}
	}

	return nil
}

func isGoModule() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

func isToolAvailable(name string) bool {
	_, err := os.LookupEnv(name)
	return err != false
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir, _ = filepath.Abs(".")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command %s %v failed:\n%s\n", name, args, string(output))
		return err
	}

	if len(output) > 0 {
		fmt.Printf("%s\n", string(output))
	}

	return nil
}
