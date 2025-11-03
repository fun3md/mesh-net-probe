// Test coverage checker for constitutional requirements
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	minCoverage = 80.0 // 80% minimum as per constitution
)

func main() {
	if err := runCoverageCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "Coverage check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Coverage requirements satisfied (>=%.0f%%)\n", minCoverage)
}

func runCoverageCheck() error {
	// Run go test with coverage
	cmd := exec.Command("go", "test", "-race", "-cover", "-coverprofile=coverage.out", "./...")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("coverage test failed: %v\nOutput: %s", err, string(output))
	}

	// Parse coverage percentage from output
	coverage, err := parseCoveragePercentage(string(output))
	if err != nil {
		return fmt.Errorf("failed to parse coverage: %w", err)
	}

	// Check against minimum requirement
	if coverage < minCoverage {
		return fmt.Errorf("coverage %.2f%% is below required minimum %.2f%%", coverage, minCoverage)
	}

	return nil
}

func parseCoveragePercentage(output string) (float64, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "coverage:") && strings.Contains(line, "%") {
			// Extract coverage percentage
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.Contains(part, "%") {
					if idx := strings.Index(part, "%"); idx != -1 {
						coverageStr := part[:idx]
						if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
							return coverage, nil
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("no coverage percentage found in output")
}
