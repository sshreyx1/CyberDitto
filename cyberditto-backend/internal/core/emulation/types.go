// internal/core/emulation/types.go

package emulation

import (
    "time"
)

type EmulationMode string

const (
    ModeTactic    EmulationMode = "tactic"
    ModeTechnique EmulationMode = "technique"
    ModeCustom    EmulationMode = "custom"
)

type EmulationStatus struct {
    Phase       string    `json:"phase"`
    Progress    float64   `json:"progress"`
    Message     string    `json:"message,omitempty"`
    Error       string    `json:"error,omitempty"`
    StartTime   time.Time `json:"start_time"`
    CompletedAt time.Time `json:"completed_at,omitempty"`
}

type TestSummary struct {
    TotalTests     int     `json:"total_tests"`
    PassedTests    int     `json:"passed_tests"`
    FailedTests    int     `json:"failed_tests"`
    SuccessRate    float64 `json:"success_rate"`
}

type TestResult struct {
    Timestamp     string `json:"timestamp"`
    Status        string `json:"status"`
    Tactic        string `json:"tactic"`
    Technique     string `json:"technique"`
    TechniqueName string `json:"technique_name"`
    TestNumber    string `json:"test_number"`
    TestName      string `json:"test_name"`
    TestGUID      string `json:"test_guid"`
    ExecutorName  string `json:"executor_name"`
    ErrorMessage  string `json:"error_message,omitempty"`
    LogFile       string `json:"log_file,omitempty"`
}

type EmulationResult struct {
    ID          string      `json:"id"`
    DeployID    string      `json:"deploy_id"`
    Status      string      `json:"status"`
    Summary     TestSummary `json:"summary"`
    Results     []TestResult`json:"results"`
    CreatedAt   time.Time   `json:"created_at"`
    CompletedAt time.Time   `json:"completed_at"`
    Error       string      `json:"error,omitempty"`
}

type Execution struct {
    ID           string           `json:"id"`
    DeploymentID string           `json:"deployment_id"`
    Mode         EmulationMode    `json:"mode"`
    Target       string           `json:"target"`
    TestNumbers  string           `json:"test_numbers,omitempty"`
    Status       *EmulationStatus `json:"status"`
    Result       *EmulationResult `json:"result,omitempty"`
    StartTime    time.Time        `json:"start_time"`
    CompletedAt  time.Time        `json:"completed_at,omitempty"`
}

type ExecutionConfig struct {
    DeploymentID string
    Mode         EmulationMode
    Target       string
    TestNumbers  string
    OutputPath   string
}

type ExecutionResult struct {
    ResultsPath string
    LogPath     string
    Error       error
}