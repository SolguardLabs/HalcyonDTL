package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func RunCLI(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return 0
	}
	if args[0] == "--list" || args[0] == "list" {
		for _, name := range ScenarioNames() {
			fmt.Fprintln(stdout, name)
		}
		return 0
	}
	switch args[0] {
	case "scenario":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: halcyondtl scenario <name>")
			return 2
		}
		engine, err := BuildScenario(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		report := engine.Report(args[1])
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "validate":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: halcyondtl validate <name>")
			return 2
		}
		engine, err := BuildScenario(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		failures := engine.Validate()
		if len(failures) > 0 {
			fmt.Fprintf(stderr, "validation failed for %s: %v\n", args[1], failures)
			return 1
		}
		fmt.Fprintf(stdout, "ok %s\n", args[1])
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %s\n", args[0])
		return 2
	}
}

func printHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "halcyondtl --list")
	fmt.Fprintln(stdout, "halcyondtl scenario <name>")
	fmt.Fprintln(stdout, "halcyondtl validate <name>")
}
