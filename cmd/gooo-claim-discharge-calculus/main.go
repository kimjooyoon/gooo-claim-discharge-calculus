package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-claim-discharge-calculus/internal/calculus"
)

const version = "v0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	if os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if os.Args[1] != "generate" && os.Args[1] != "evaluate" && os.Args[1] != "conformance" {
		usage()
		os.Exit(2)
	}
	flags := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	source := flags.String("source", "examples/claim-discharge-calculus.gooo", "path to the authoritative .gooo source")
	root := flags.String("root", ".", "repository root to measure")
	outputDir := flags.String("output-dir", "", "absolute empty directory for user outputs")
	runner := flags.String("runner", "unspecified", "stable runner identity")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s %s --output-dir /absolute/output [options]\n", os.Args[0], os.Args[1])
		flags.PrintDefaults()
	}
	if err := flags.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}
	if *outputDir == "" {
		fmt.Fprintln(os.Stderr, "--output-dir is required")
		os.Exit(2)
	}
	result, err := calculus.Run(calculus.RunOptions{
		Root:       *root,
		SourcePath: *source,
		OutputDir:  *outputDir,
		Runner:     *runner,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("schema=%s cases=%d closed=%d unknown=%d refuted=%d active_claims=%d removed_claims=%d artifacts=%d\n", result.Report.Schema, len(result.Report.Cases), result.Report.FixedVector.Cases.Closed, result.Report.FixedVector.Cases.Unknown, result.Report.FixedVector.Cases.Refuted, len(result.Report.ActiveFrontier.ActiveClaimIDs), len(result.Report.ActiveFrontier.RemovedClaimIDs), len(result.Artifacts))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <generate|evaluate|conformance|version> [options]\n", os.Args[0])
}
