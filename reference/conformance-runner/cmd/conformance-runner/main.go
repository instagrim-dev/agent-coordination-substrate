package main

import (
	"fmt"
	"os"
	"path/filepath"

	conformance "github.com/instagrim-dev/agent-coordination-substrate/reference/conformance-runner"
	substrate "github.com/instagrim-dev/agent-coordination-substrate/reference/go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: conformance-runner <expectations-dir>\n")
		fmt.Fprintf(os.Stderr, "  expectations-dir: path to the spec root (containing advisory/ and enforcement/)\n")
		os.Exit(1)
	}

	specRoot := os.Args[1]

	advisoryPath := filepath.Join(specRoot, "advisory", "conformance", "expectations.yaml")
	enforcementPath := filepath.Join(specRoot, "enforcement", "conformance", "expectations.yaml")

	clock := conformance.NewTestClock(1716000000)
	runner := conformance.NewRunner(
		substrate.NewMemSignalStore(clock),
		substrate.NewMemClaimStore(clock),
		clock,
	)

	var totalPass, totalFail, totalSkip int

	// Advisory layer
	if data, err := os.ReadFile(advisoryPath); err == nil {
		fmt.Println("=== Advisory Layer ===")
		results := runner.RunAdvisory(data)
		for _, r := range results {
			switch r.Status {
			case "pass":
				totalPass++
				fmt.Printf("  PASS  [%s] %s\n", r.Group, r.ScenarioID)
			case "fail":
				totalFail++
				fmt.Printf("  FAIL  [%s] %s: %s\n", r.Group, r.ScenarioID, r.Message)
			case "skip":
				totalSkip++
				fmt.Printf("  SKIP  [%s] %s: %s\n", r.Group, r.ScenarioID, r.Message)
			}
		}
		fmt.Println()
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not read %s: %v\n", advisoryPath, err)
	}

	// Enforcement layer
	if data, err := os.ReadFile(enforcementPath); err == nil {
		fmt.Println("=== Enforcement Layer ===")
		results := runner.RunEnforcement(data)
		for _, r := range results {
			switch r.Status {
			case "pass":
				totalPass++
				fmt.Printf("  PASS  [%s] %s\n", r.Group, r.ScenarioID)
			case "fail":
				totalFail++
				fmt.Printf("  FAIL  [%s] %s: %s\n", r.Group, r.ScenarioID, r.Message)
			case "skip":
				totalSkip++
				fmt.Printf("  SKIP  [%s] %s: %s\n", r.Group, r.ScenarioID, r.Message)
			}
		}
		fmt.Println()
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not read %s: %v\n", enforcementPath, err)
	}

	fmt.Printf("Summary: %d passed, %d failed, %d skipped\n", totalPass, totalFail, totalSkip)
	if totalFail > 0 {
		os.Exit(1)
	}
}
