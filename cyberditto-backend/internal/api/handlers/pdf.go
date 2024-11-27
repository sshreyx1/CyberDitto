package handlers

import (
	"bytes"
	"cyberditto-backend/internal/core/emulation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/jung-kurt/gofpdf"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func GeneratePDF(w http.ResponseWriter, r *http.Request) {
	dataSources, err := emulation.LoadDashboardDataFromDirectory()
	if err != nil {
		http.Error(w, "Failed to load data for PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(0, 20, "Adversary Emulation Report with AI Insights")
	pdf.Ln(20)

	for source, data := range dataSources {
		insights, aiErr := fetchAIInsights(data)
		if aiErr != nil {
			fmt.Printf("Error fetching AI insights for %s: %v\n", source, aiErr)
			insights = "AI insights not available for this section."
		}

		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, source+" Benchmarks")
		pdf.Ln(15)

		chartFilename, err := createComplianceBarChart(source, data)
		if err != nil {
			fmt.Println("Error creating chart:", err)
			continue
		}
		defer os.Remove(chartFilename)

		chartWidth := 140.0
		chartHeight := 80.0
		currentY := pdf.GetY()
		pdf.ImageOptions(chartFilename, 10, currentY, chartWidth, chartHeight, false, gofpdf.ImageOptions{}, 0, "")
		pdf.Ln(chartHeight + 10)

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 10, "AI Insights")
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 10, insights, "", "", false)

		benchmarkNameWidth := 140.0
		complianceWidth := 40.0
		rowHeight := 10.0
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(benchmarkNameWidth, rowHeight, "Benchmark Name", "1", 0, "L", false, 0, "")
		pdf.CellFormat(complianceWidth, rowHeight, "Compliance", "1", 1, "R", false, 0, "")

		pdf.SetFont("Arial", "", 10)
		for _, benchmark := range data.BenchmarkData {
			benchmarkName := truncateString(benchmark.Name, 60)
			pdf.CellFormat(benchmarkNameWidth, rowHeight, benchmarkName, "1", 0, "L", false, 0, "")
			complianceColor := getComplianceColor(benchmark.Compliance)
			pdf.SetTextColor(complianceColor[0], complianceColor[1], complianceColor[2])
			pdf.CellFormat(complianceWidth, rowHeight, fmt.Sprintf("%.2f%%", benchmark.Compliance), "1", 1, "R", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
		}
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		http.Error(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(buffer.Bytes())
}

func fetchAIInsights(data emulation.DashboardData) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API key not found in environment variables")
	}

	client := &http.Client{}
	apiURL := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert in cybersecurity emulation reporting.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Provide insights for the following compliance data: %v", data.BenchmarkData),
			},
		},
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed: %s (Status Code: %d)", string(body), resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices := result["choices"].([]interface{})
	if len(choices) > 0 {
		message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
		return message["content"].(string), nil
	}

	return "", fmt.Errorf("no insights returned from API")
}

func createComplianceBarChart(source string, data emulation.DashboardData) (string, error) {
	p := plot.New()

	values := make(plotter.Values, len(data.BenchmarkData))
	labels := make([]string, len(data.BenchmarkData))

	for i, benchmark := range data.BenchmarkData {
		values[i] = benchmark.Compliance
		labels[i] = truncateString(benchmark.Name, 15)
	}

	barChart, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		return "", fmt.Errorf("error creating bar chart: %v", err)
	}

	p.Title.Text = fmt.Sprintf("%s Compliance Overview", source)
	p.X.Label.Text = "Benchmarks"
	p.Y.Label.Text = "Compliance (%)"
	p.Add(barChart)
	p.NominalX(labels...)

	chartFilename := fmt.Sprintf("%s_chart.png", source)
	if err := p.Save(4*vg.Inch, 3*vg.Inch, chartFilename); err != nil {
		return "", fmt.Errorf("error saving plot: %v", err)
	}

	return chartFilename, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func getComplianceColor(compliance float64) [3]int {
	switch {
	case compliance == 0:
		return [3]int{255, 0, 0}
	case compliance < 50:
		return [3]int{255, 165, 0}
	case compliance < 100:
		return [3]int{255, 255, 0}
	default:
		return [3]int{0, 255, 0}
	}
}
