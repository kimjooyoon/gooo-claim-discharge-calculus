package calculus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type RunManifest struct {
	Schema             string      `json:"schema"`
	SourcePath         string      `json:"source_path"`
	SourceDigest       string      `json:"source_digest"`
	SemanticDigest     string      `json:"semantic_digest"`
	MachineReportDigest string     `json:"machine_report_digest"`
	Artifacts          []Artifact  `json:"artifacts"`
}

func Run(options RunOptions) (RunResult, error) {
	if options.SourcePath == "" {
		return RunResult{}, fmt.Errorf("source path is required")
	}
	if options.OutputDir == "" {
		return RunResult{}, fmt.Errorf("output directory is required")
	}
	if !filepath.IsAbs(options.OutputDir) {
		return RunResult{}, fmt.Errorf("output directory must be absolute")
	}
	sourcePath, err := filepath.Abs(options.SourcePath)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolve source path: %w", err)
	}
	root, err := filepath.Abs(options.Root)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolve root path: %w", err)
	}
	outputDir, err := filepath.Abs(options.OutputDir)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolve output path: %w", err)
	}
	if err := prepareOutputDir(outputDir); err != nil {
		return RunResult{}, err
	}
	started := time.Now()
	source, err := ParseFile(sourcePath)
	if err != nil {
		return RunResult{}, err
	}
	report, err := Evaluate(source, sourcePath)
	if err != nil {
		return RunResult{}, err
	}
	if err := ValidateConformance(source, report); err != nil {
		return RunResult{}, err
	}
	measurements, err := measureWorkspace(root, outputDir)
	if err != nil {
		return RunResult{}, fmt.Errorf("measure workspace: %w", err)
	}
	measurements.WallMS = time.Since(started).Milliseconds()
	report.Runtime = RuntimeMetadata{
		Runner:    options.Runner,
		GoVersion: runtime.Version(),
		Authority: source.Policy.Runtime,
		Measurements: measurements,
	}
	if report.Runtime.Runner == "" {
		report.Runtime.Runner = "unspecified"
	}
	semanticPath := filepath.Join(outputDir, "semantic-ir.json")
	dossierPath := filepath.Join(outputDir, "human-dossier.md")
	machinePath := filepath.Join(outputDir, "machine.json")
	manifestPath := filepath.Join(outputDir, "run-manifest.json")
	if err := writeJSON(semanticPath, report.SemanticIR); err != nil {
		return RunResult{}, err
	}
	initialArtifacts, err := collectArtifacts(outputDir)
	if err != nil {
		return RunResult{}, err
	}
	report.Runtime.Measurements.GeneratedArtifacts = initialArtifacts
	if err := writeText(dossierPath, renderDossier(report)); err != nil {
		return RunResult{}, err
	}
	initialArtifacts, err = collectArtifacts(outputDir)
	if err != nil {
		return RunResult{}, err
	}
	report.Runtime.Measurements.GeneratedArtifacts = initialArtifacts
	report.ReportDigest = ""
	report.ReportDigest, err = DigestValue(report)
	if err != nil {
		return RunResult{}, fmt.Errorf("digest machine report: %w", err)
	}
	if err := writeJSON(machinePath, report); err != nil {
		return RunResult{}, err
	}
	artifactsBeforeManifest, err := collectArtifacts(outputDir)
	if err != nil {
		return RunResult{}, err
	}
	manifest := RunManifest{
		Schema:              Schema,
		SourcePath:          sourcePath,
		SourceDigest:        report.SourceDigest,
		SemanticDigest:      report.SemanticDigest,
		MachineReportDigest: report.ReportDigest,
		Artifacts:           artifactsBeforeManifest,
	}
	if err := writeJSON(manifestPath, manifest); err != nil {
		return RunResult{}, err
	}
	artifacts, err := collectArtifacts(outputDir)
	if err != nil {
		return RunResult{}, err
	}
	return RunResult{Report: report, Artifacts: artifacts}, nil
}

func prepareOutputDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path is not a directory: %s", path)
		}
		entries, readErr := os.ReadDir(path)
		if readErr != nil {
			return fmt.Errorf("inspect output directory: %w", readErr)
		}
		if len(entries) != 0 {
			return fmt.Errorf("output directory must be empty: %s", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output directory: %w", err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeText(path string, value string) error {
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
