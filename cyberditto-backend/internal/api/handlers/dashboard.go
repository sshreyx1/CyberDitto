package handlers

import (
	"encoding/json"
	"net/http"
	"cyberditto-backend/internal/core/emulation"
)

// GetDashboardData serves aggregated dashboard data for the security dashboard
func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	// Load all data sources from the emulation parser
	dataSources, err := emulation.LoadDashboardDataFromDirectory()
	if err != nil {
		http.Error(w, "Failed to load dashboard data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate aggregate metrics
	aggregateMetrics := calculateAggregateMetrics(dataSources)

	// Prepare response
	response := struct {
		CISCompliance     map[string]float64                        `json:"cisCompliance"`
		MITREEmulation    []emulation.EmulationResult               `json:"mitreEmulation"`
		AggregateMetrics  map[string]interface{}                   `json:"aggregateMetrics"`
		BenchmarkStatuses map[string][]emulation.BenchmarkData      `json:"benchmarkStatuses"`
	}{
		CISCompliance:     calculateCISCompliance(dataSources["cis-benchmark"].BenchmarkData),
		MITREEmulation:    dataSources["mitre-attack"].EmulationData,
		AggregateMetrics:  aggregateMetrics,
		BenchmarkStatuses: map[string][]emulation.BenchmarkData{
			"CIS Benchmark": dataSources["cis-benchmark"].BenchmarkData,
			"MITRE ATT&CK":  dataSources["mitre-attack"].BenchmarkData,
		},
	}

	// Encode response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// calculateAggregateMetrics computes aggregate metrics across all data sources
func calculateAggregateMetrics(dataSources map[string]emulation.DashboardData) map[string]interface{} {
	totalCompliance := 0.0
	totalSuccess := 0
	totalFailure := 0
	totalBenchmarks := 0

	// Aggregate metrics
	for _, data := range dataSources {
		for _, benchmark := range data.BenchmarkData {
			totalCompliance += benchmark.Compliance
			totalBenchmarks++
		}
		for _, emulation := range data.EmulationData {
			totalSuccess += emulation.Success
			totalFailure += emulation.Failure
		}
	}

	averageCompliance := 0.0
	if totalBenchmarks > 0 {
		averageCompliance = totalCompliance / float64(totalBenchmarks)
	}

	return map[string]interface{}{
		"averageCompliance": averageCompliance,
		"totalSuccess":      totalSuccess,
		"totalFailure":      totalFailure,
		"totalBenchmarks":   totalBenchmarks,
	}
}

// calculateCISCompliance computes compliance metrics for CIS Benchmark data
func calculateCISCompliance(benchmarkData []emulation.BenchmarkData) map[string]float64 {
	totalBenchmarks := len(benchmarkData)
	compliantBenchmarks := 0

	for _, benchmark := range benchmarkData {
		if benchmark.Compliance >= 100.0 { // Assuming 100% compliance is "PASS"
			compliantBenchmarks++
		}
	}

	compliancePercentage := 0.0
	if totalBenchmarks > 0 {
		compliancePercentage = (float64(compliantBenchmarks) / float64(totalBenchmarks)) * 100
	}

	return map[string]float64{
		"totalBenchmarks":      float64(totalBenchmarks),
		"compliantBenchmarks":  float64(compliantBenchmarks),
		"compliancePercentage": compliancePercentage,
	}
}
