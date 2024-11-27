package emulation

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Data Structures
type DashboardData struct {
	BenchmarkData  []BenchmarkData   `json:"benchmarkData"`
	EmulationData  []EmulationResult `json:"emulationData"`
	SecurityStatus []SecurityStatus  `json:"securityStatus"`
}

type BenchmarkData struct {
	Name       string            `json:"name"`
	Compliance float64           `json:"compliance"`
	Details    map[string]string `json:"details"`
}

type EmulationResult struct {
	Name    string `json:"name"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

type SecurityStatus struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

type MITREAttackResult struct {
	Tactic         string  `json:"tactic"`
	Technique      string  `json:"technique"`
	TechniqueName  string  `json:"techniqueName"`
	PassCount      int     `json:"passCount"`
	FailCount      int     `json:"failCount"`
	TotalTests     int     `json:"totalTests"`
	ComplianceRate float64 `json:"complianceRate"`
}

// MITRE ATT&CK Parsing Function
func ParseMITREAttackCSV(filePath string) (DashboardData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return DashboardData{}, fmt.Errorf("failed to open MITRE ATT&CK CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return DashboardData{}, fmt.Errorf("failed to read MITRE ATT&CK CSV: %v", err)
	}

	techniqueResults := make(map[string]*MITREAttackResult)

	for _, record := range records[1:] { // Skip header
		if len(record) < 9 {
			continue // Skip malformed rows
		}

		technique := record[3]
		techniqueName := record[4]
		status := strings.ToLower(record[1])

		result, exists := techniqueResults[technique]
		if !exists {
			result = &MITREAttackResult{
				Tactic:        record[2],
				Technique:     technique,
				TechniqueName: techniqueName,
			}
			techniqueResults[technique] = result
		}

		result.TotalTests++
		if status == "pass" {
			result.PassCount++
		} else {
			result.FailCount++
		}
	}

	var benchmarkData []BenchmarkData
	var emulationData []EmulationResult
	for _, result := range techniqueResults {
		result.ComplianceRate = (float64(result.PassCount) / float64(result.TotalTests)) * 100

		benchmarkData = append(benchmarkData, BenchmarkData{
			Name:       result.TechniqueName,
			Compliance: result.ComplianceRate,
		})

		emulationData = append(emulationData, EmulationResult{
			Name:    result.TechniqueName,
			Success: result.PassCount,
			Failure: result.FailCount,
		})
	}

	return DashboardData{
		BenchmarkData: benchmarkData,
		EmulationData: emulationData,
	}, nil
}

// ParseCISBenchmarkCSV processes the CIS Benchmark data from the CSV file
func ParseCISBenchmarkCSV(filePath string) (DashboardData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return DashboardData{}, fmt.Errorf("failed to open CIS Benchmark CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return DashboardData{}, fmt.Errorf("failed to read CIS Benchmark CSV: %v", err)
	}

	var benchmarkData []BenchmarkData
	totalControls := 0
	compliantControls := 0

	for _, record := range records[1:] { // Skip header row
		if len(record) < 6 {
			continue // Skip invalid or malformed rows
		}

		// Extract necessary fields
		benchmarkNumber := record[0]
		benchmarkName := record[1]
		expectedValue := record[2]
		actualValue := record[4]
		complianceStatus := strings.ToLower(record[5])

		// Determine compliance score
		complianceScore := 0.0
		if complianceStatus == "pass" {
			complianceScore = 100.0
			compliantControls++
		}
		totalControls++

		// Add to benchmark data
		benchmarkData = append(benchmarkData, BenchmarkData{
			Name:       fmt.Sprintf("%s: %s", benchmarkNumber, benchmarkName),
			Compliance: complianceScore,
			Details: map[string]string{
				"Expected": expectedValue,
				"Actual":   actualValue,
				"Level":    record[3],
			},
		})
	}

	// Create the dashboard data structure
	return DashboardData{
		BenchmarkData: benchmarkData,
	}, nil
}

// Unified Function to Parse CSV Files
func ParseEmulationCSV(filePath string) (DashboardData, error) {
	if isMITREAttackCSV(filePath) {
		return ParseMITREAttackCSV(filePath)
	}

	if isCISBenchmarkCSV(filePath) {
		return ParseCISBenchmarkCSV(filePath)
	}

	return DashboardData{}, fmt.Errorf("unsupported CSV format for file: %s", filePath)
}

// Helper Functions for CSV Detection
func isMITREAttackCSV(filePath string) bool {
	return containsHeaders(filePath, []string{"Timestamp", "Status", "Tactic", "Technique", "Technique Name"})
}

func isCISBenchmarkCSV(filePath string) bool {
    expectedHeaders := []string{"Benchmark_Number", "Benchmark", "Expected_Value", "Level", "Actual_Value", "Compliance_Status"}
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Printf("Debug: Failed to open file: %v\n", err)
        return false
    }
    defer file.Close()

    reader := csv.NewReader(file)
    headers, err := reader.Read()
    if err != nil {
        fmt.Printf("Debug: Failed to read headers: %v\n", err)
        return false
    }

    // Print actual headers for debugging
    fmt.Printf("Debug: Found headers: %v\n", headers)
    fmt.Printf("Debug: Expected headers: %v\n", expectedHeaders)

    for i, header := range expectedHeaders {
        if i >= len(headers) || headers[i] != header {
            fmt.Printf("Debug: Mismatch at position %d. Expected '%s', got '%s'\n", 
                      i, header, headers[i])
            return false
        }
    }
    return true
}

func containsHeaders(filePath string, expectedHeaders []string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil || len(headers) < len(expectedHeaders) {
		return false
	}

	for i, header := range expectedHeaders {
		if i >= len(headers) || headers[i] != header {
			return false
		}
	}
	return true
}

func LoadDashboardDataFromDirectory() (map[string]DashboardData, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Possible relative paths to emulation-results
	possiblePaths := []string{
		"./emulation-results",
		"../emulation-results",
		"../../emulation-results",
		filepath.Join(currentDir, "emulation-results"),
		filepath.Join(currentDir, "..", "emulation-results"),
		filepath.Join(currentDir, "..", "..", "emulation-results"),
	}

	var fullPath string
	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		fmt.Printf("Checking path: %s (Absolute: %s)\n", path, absPath)

		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			// Verify it contains the expected subdirectories
			if _, err1 := os.Stat(filepath.Join(absPath, "mitre-attack")); err1 == nil {
				if _, err2 := os.Stat(filepath.Join(absPath, "cis-benchmark")); err2 == nil {
					fullPath = absPath
					break
				}
			}
		}
	}

	if fullPath == "" {
		return nil, fmt.Errorf("could not find emulation-results directory. Current working directory: %s", currentDir)
	}

	fmt.Printf("Using emulation results directory: %s\n", fullPath)

	dataSources := map[string]DashboardData{}
	subdirs := []string{"mitre-attack", "cis-benchmark"}

	for _, subdir := range subdirs {
		fullSubdirPath := filepath.Join(fullPath, subdir)

		fmt.Printf("Checking subdirectory: %s\n", fullSubdirPath)

		files, err := os.ReadDir(fullSubdirPath)
		if err != nil {
			fmt.Printf("ERROR: Unable to read directory %s: %v\n", fullSubdirPath, err)
			continue
		}

		fmt.Printf("Found %d files in %s\n", len(files), fullSubdirPath)

		var combinedData DashboardData
		for _, file := range files {
			fmt.Printf("Processing file: %s\n", file.Name())

			if filepath.Ext(file.Name()) != ".csv" {
				fmt.Printf("Skipping non-CSV file: %s\n", file.Name())
				continue
			}

			filePath := filepath.Join(fullSubdirPath, file.Name())

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Printf("File does not exist: %s\n", filePath)
				continue
			}

			data, err := ParseEmulationCSV(filePath)
			if err != nil {
				fmt.Printf("ERROR parsing %s: %v\n", filePath, err)
				continue
			}

			combinedData.BenchmarkData = append(combinedData.BenchmarkData, data.BenchmarkData...)
			combinedData.EmulationData = append(combinedData.EmulationData, data.EmulationData...)
			combinedData.SecurityStatus = append(combinedData.SecurityStatus, data.SecurityStatus...)
		}

		if len(combinedData.BenchmarkData) > 0 || len(combinedData.EmulationData) > 0 {
			dataSources[subdir] = combinedData
			fmt.Printf("Added data for %s: %d benchmark entries, %d emulation entries\n",
				subdir, len(combinedData.BenchmarkData), len(combinedData.EmulationData))
		} else {
			fmt.Printf("NO DATA found for %s\n", subdir)
		}
	}

	if len(dataSources) == 0 {
		return nil, fmt.Errorf("no data sources found in directory: %s", fullPath)
	}

	return dataSources, nil
}
