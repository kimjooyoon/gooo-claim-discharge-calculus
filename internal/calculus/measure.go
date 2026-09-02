package calculus

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func measureWorkspace(root string, outputDir string) (WorkspaceMeasurements, error) {
	measurements := WorkspaceMeasurements{GeneratedArtifacts: []Artifact{}, Resolution: "UNKNOWN"}
	root, err := filepath.Abs(root)
	if err != nil {
		return measurements, err
	}
	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return measurements, err
	}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == outputDir {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			measurements.Directories++
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		measurements.Files++
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		lineCount := physicalLineCount(data)
		extension := filepath.Ext(path)
		if extension == ".go" {
			measurements.GoLines += lineCount
		}
		if extension == ".gooo" && filepath.Base(path) != "README.md" {
			measurements.GoooLinesExcludingRootREADME += lineCount
		}
		return nil
	})
	if err != nil {
		return measurements, err
	}
	measurements.PeakRSSKiB = peakRSSKiB()
	if measurements.PeakRSSKiB != nil {
		measurements.Resolution = "CLOSED"
	}
	return measurements, nil
}

func physicalLineCount(data []byte) int64 {
	if len(data) == 0 {
		return 0
	}
	count := int64(bytes.Count(data, []byte{'\n'}))
	if data[len(data)-1] != '\n' {
		count++
	}
	return count
}

func peakRSSKiB() *int64 {
	if runtime.GOOS != "linux" {
		return nil
	}
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || value < 0 {
			return nil
		}
		if len(fields) >= 3 && fields[2] == "kB" {
			return &value
		}
		return nil
	}
	return nil
}

func collectArtifacts(directory string) ([]Artifact, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	artifacts := make([]Artifact, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, Artifact{Name: entry.Name(), Bytes: info.Size(), SHA256: DigestBytes(data)})
	}
	return artifacts, nil
}
