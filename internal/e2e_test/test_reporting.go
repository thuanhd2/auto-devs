package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TestReport represents a comprehensive test execution report
type TestReport struct {
	Summary    TestSummary      `json:"summary"`
	Suites     []TestSuite      `json:"suites"`
	Performance PerformanceData  `json:"performance"`
	Coverage   CoverageData     `json:"coverage"`
	Artifacts  []Artifact       `json:"artifacts"`
	Metadata   ReportMetadata   `json:"metadata"`
}

// TestSummary provides overall test execution summary
type TestSummary struct {
	TotalTests     int           `json:"total_tests"`
	PassedTests    int           `json:"passed_tests"`
	FailedTests    int           `json:"failed_tests"`
	SkippedTests   int           `json:"skipped_tests"`
	Duration       time.Duration `json:"duration"`
	SuccessRate    float64       `json:"success_rate"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Environment    string        `json:"environment"`
	Version        string        `json:"version"`
}

// TestSuite represents results from a specific test suite
type TestSuite struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Tests        []TestCase    `json:"tests"`
	Duration     time.Duration `json:"duration"`
	Status       TestStatus    `json:"status"`
	SetupTime    time.Duration `json:"setup_time"`
	TeardownTime time.Duration `json:"teardown_time"`
	Errors       []TestError   `json:"errors"`
}

// TestCase represents an individual test case result
type TestCase struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      TestStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	Error       *TestError    `json:"error,omitempty"`
	Assertions  int           `json:"assertions"`
	Steps       []TestStep    `json:"steps"`
	Tags        []string      `json:"tags"`
	Metadata    interface{}   `json:"metadata,omitempty"`
}

// TestStep represents a step within a test case
type TestStep struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      TestStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	Error       *TestError    `json:"error,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// TestError represents a test failure or error
type TestError struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Details   string      `json:"details"`
	Stack     string      `json:"stack"`
	Location  string      `json:"location"`
	Context   interface{} `json:"context,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// TestStatus represents the status of a test
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusError   TestStatus = "error"
)

// PerformanceData contains performance metrics
type PerformanceData struct {
	Benchmarks     []Benchmark    `json:"benchmarks"`
	Metrics        []Metric       `json:"metrics"`
	Thresholds     []Threshold    `json:"thresholds"`
	ResourceUsage  ResourceUsage  `json:"resource_usage"`
	Trends         []TrendData    `json:"trends"`
}

// Benchmark represents benchmark test results
type Benchmark struct {
	Name           string        `json:"name"`
	Operations     int           `json:"operations"`
	NsPerOp        int64         `json:"ns_per_op"`
	MBPerSec       float64       `json:"mb_per_sec"`
	AllocsPerOp    int           `json:"allocs_per_op"`
	BytesPerOp     int           `json:"bytes_per_op"`
	Duration       time.Duration `json:"duration"`
	MemoryProfile  string        `json:"memory_profile,omitempty"`
	CPUProfile     string        `json:"cpu_profile,omitempty"`
}

// Metric represents a performance metric
type Metric struct {
	Name        string      `json:"name"`
	Value       float64     `json:"value"`
	Unit        string      `json:"unit"`
	Timestamp   time.Time   `json:"timestamp"`
	Context     string      `json:"context"`
	Tags        []string    `json:"tags"`
	Threshold   *Threshold  `json:"threshold,omitempty"`
	Status      TestStatus  `json:"status"`
}

// Threshold represents a performance threshold
type Threshold struct {
	Name      string  `json:"name"`
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"` // >, <, >=, <=, ==
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Severity  string  `json:"severity"` // warning, error, critical
}

// ResourceUsage contains system resource usage metrics
type ResourceUsage struct {
	Memory     MemoryUsage `json:"memory"`
	CPU        CPUUsage    `json:"cpu"`
	Disk       DiskUsage   `json:"disk"`
	Network    NetworkUsage `json:"network"`
	Database   DBUsage     `json:"database"`
	Goroutines int         `json:"goroutines"`
}

// MemoryUsage contains memory usage statistics
type MemoryUsage struct {
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	GCCycles     uint32 `json:"gc_cycles"`
	LastGCTime   time.Time `json:"last_gc_time"`
}

// CPUUsage contains CPU usage statistics
type CPUUsage struct {
	UserTime    time.Duration `json:"user_time"`
	SystemTime  time.Duration `json:"system_time"`
	Utilization float64       `json:"utilization"`
}

// DiskUsage contains disk usage statistics
type DiskUsage struct {
	BytesRead    uint64 `json:"bytes_read"`
	BytesWritten uint64 `json:"bytes_written"`
	IOCount      uint64 `json:"io_count"`
}

// NetworkUsage contains network usage statistics
type NetworkUsage struct {
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
	PacketsSent   uint64 `json:"packets_sent"`
	PacketsRecv   uint64 `json:"packets_received"`
	Connections   int    `json:"connections"`
}

// DBUsage contains database usage statistics
type DBUsage struct {
	Connections    int           `json:"connections"`
	Queries        uint64        `json:"queries"`
	QueryTime      time.Duration `json:"query_time"`
	SlowQueries    int           `json:"slow_queries"`
	DeadLocks      int           `json:"deadlocks"`
	CacheHitRatio  float64       `json:"cache_hit_ratio"`
}

// TrendData represents historical trend data
type TrendData struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Build     string    `json:"build"`
	Branch    string    `json:"branch"`
}

// CoverageData contains code coverage information
type CoverageData struct {
	Overall    CoverageStats              `json:"overall"`
	Packages   map[string]CoverageStats   `json:"packages"`
	Files      map[string]CoverageStats   `json:"files"`
	Functions  map[string]CoverageStats   `json:"functions"`
	Lines      []LineCoverage             `json:"lines"`
	Reports    []string                   `json:"reports"`
}

// CoverageStats represents coverage statistics
type CoverageStats struct {
	Total     int     `json:"total"`
	Covered   int     `json:"covered"`
	Percentage float64 `json:"percentage"`
}

// LineCoverage represents line-by-line coverage
type LineCoverage struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Covered bool   `json:"covered"`
	Count   int    `json:"count"`
}

// Artifact represents test artifacts
type Artifact struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReportMetadata contains report metadata
type ReportMetadata struct {
	ReportVersion string            `json:"report_version"`
	GeneratedAt   time.Time         `json:"generated_at"`
	GeneratedBy   string            `json:"generated_by"`
	Environment   map[string]string `json:"environment"`
	GitCommit     string            `json:"git_commit"`
	GitBranch     string            `json:"git_branch"`
	BuildNumber   string            `json:"build_number"`
	Tags          []string          `json:"tags"`
}

// TestReportGenerator generates comprehensive test reports
type TestReportGenerator struct {
	reportDir string
	artifacts []Artifact
	metadata  ReportMetadata
}

// NewTestReportGenerator creates a new test report generator
func NewTestReportGenerator(reportDir string) *TestReportGenerator {
	return &TestReportGenerator{
		reportDir: reportDir,
		artifacts: make([]Artifact, 0),
		metadata: ReportMetadata{
			ReportVersion: "1.0",
			GeneratedAt:   time.Now(),
			GeneratedBy:   "E2E Test Suite",
			Environment:   make(map[string]string),
			Tags:          make([]string, 0),
		},
	}
}

// GenerateReport generates a comprehensive test report
func (g *TestReportGenerator) GenerateReport(suites []TestSuite, performance PerformanceData, coverage CoverageData) (*TestReport, error) {
	summary := g.calculateSummary(suites)
	
	report := &TestReport{
		Summary:     summary,
		Suites:      suites,
		Performance: performance,
		Coverage:    coverage,
		Artifacts:   g.artifacts,
		Metadata:    g.metadata,
	}

	// Ensure report directory exists
	if err := os.MkdirAll(g.reportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}

	// Generate various report formats
	if err := g.generateJSONReport(report); err != nil {
		return nil, fmt.Errorf("failed to generate JSON report: %w", err)
	}

	if err := g.generateHTMLReport(report); err != nil {
		return nil, fmt.Errorf("failed to generate HTML report: %w", err)
	}

	if err := g.generateMarkdownReport(report); err != nil {
		return nil, fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	if err := g.generateJUnitXMLReport(report); err != nil {
		return nil, fmt.Errorf("failed to generate JUnit XML report: %w", err)
	}

	return report, nil
}

// calculateSummary calculates test execution summary
func (g *TestReportGenerator) calculateSummary(suites []TestSuite) TestSummary {
	var totalTests, passedTests, failedTests, skippedTests int
	var totalDuration time.Duration
	var startTime, endTime time.Time

	for _, suite := range suites {
		for _, test := range suite.Tests {
			totalTests++
			switch test.Status {
			case TestStatusPassed:
				passedTests++
			case TestStatusFailed, TestStatusError:
				failedTests++
			case TestStatusSkipped:
				skippedTests++
			}
		}
		totalDuration += suite.Duration
	}

	successRate := 0.0
	if totalTests > 0 {
		successRate = float64(passedTests) / float64(totalTests) * 100
	}

	return TestSummary{
		TotalTests:  totalTests,
		PassedTests: passedTests,
		FailedTests: failedTests,
		SkippedTests: skippedTests,
		Duration:    totalDuration,
		SuccessRate: successRate,
		StartTime:   startTime,
		EndTime:     endTime,
		Environment: g.getEnvironment(),
		Version:     g.getVersion(),
	}
}

// generateJSONReport generates JSON format report
func (g *TestReportGenerator) generateJSONReport(report *TestReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(g.reportDir, "test-report.json")
	return ioutil.WriteFile(path, data, 0644)
}

// generateHTMLReport generates HTML format report
func (g *TestReportGenerator) generateHTMLReport(report *TestReport) error {
	html := g.buildHTMLReport(report)
	path := filepath.Join(g.reportDir, "test-report.html")
	return ioutil.WriteFile(path, []byte(html), 0644)
}

// generateMarkdownReport generates Markdown format report
func (g *TestReportGenerator) generateMarkdownReport(report *TestReport) error {
	markdown := g.buildMarkdownReport(report)
	path := filepath.Join(g.reportDir, "test-report.md")
	return ioutil.WriteFile(path, []byte(markdown), 0644)
}

// generateJUnitXMLReport generates JUnit XML format report
func (g *TestReportGenerator) generateJUnitXMLReport(report *TestReport) error {
	xml := g.buildJUnitXML(report)
	path := filepath.Join(g.reportDir, "junit.xml")
	return ioutil.WriteFile(path, []byte(xml), 0644)
}

// buildHTMLReport builds HTML report content
func (g *TestReportGenerator) buildHTMLReport(report *TestReport) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E2E Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
        .summary { display: flex; justify-content: space-around; margin: 20px 0; }
        .metric { text-align: center; padding: 10px; }
        .metric-value { font-size: 24px; font-weight: bold; }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .suite { margin: 20px 0; border: 1px solid #dee2e6; border-radius: 5px; }
        .suite-header { background-color: #e9ecef; padding: 15px; font-weight: bold; }
        .test-case { padding: 10px 15px; border-bottom: 1px solid #dee2e6; }
        .test-case:last-child { border-bottom: none; }
        .error-details { background-color: #f8d7da; padding: 10px; margin: 10px 0; border-radius: 3px; }
        .performance { margin: 20px 0; }
        .benchmark { display: flex; justify-content: space-between; padding: 10px; background-color: #f8f9fa; margin: 5px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>End-to-End Test Report</h1>
        <p>Generated: %s</p>
        <p>Environment: %s</p>
        <p>Version: %s</p>
    </div>

    <div class="summary">
        <div class="metric">
            <div class="metric-value">%d</div>
            <div>Total Tests</div>
        </div>
        <div class="metric">
            <div class="metric-value passed">%d</div>
            <div>Passed</div>
        </div>
        <div class="metric">
            <div class="metric-value failed">%d</div>
            <div>Failed</div>
        </div>
        <div class="metric">
            <div class="metric-value skipped">%d</div>
            <div>Skipped</div>
        </div>
        <div class="metric">
            <div class="metric-value">%.1f%%</div>
            <div>Success Rate</div>
        </div>
        <div class="metric">
            <div class="metric-value">%s</div>
            <div>Duration</div>
        </div>
    </div>`,
		report.Metadata.GeneratedAt.Format(time.RFC3339),
		report.Summary.Environment,
		report.Summary.Version,
		report.Summary.TotalTests,
		report.Summary.PassedTests,
		report.Summary.FailedTests,
		report.Summary.SkippedTests,
		report.Summary.SuccessRate,
		report.Summary.Duration.String())

	// Add test suites
	for _, suite := range report.Suites {
		html += fmt.Sprintf(`
    <div class="suite">
        <div class="suite-header">
            %s (%d tests, %s)
        </div>`, suite.Name, len(suite.Tests), suite.Duration.String())

		for _, test := range suite.Tests {
			statusClass := strings.ToLower(string(test.Status))
			html += fmt.Sprintf(`
        <div class="test-case">
            <div>
                <span class="%s">●</span> %s (%s)
            </div>`, statusClass, test.Name, test.Duration.String())

			if test.Error != nil {
				html += fmt.Sprintf(`
            <div class="error-details">
                <strong>%s:</strong> %s<br>
                <details>
                    <summary>Details</summary>
                    <pre>%s</pre>
                </details>
            </div>`, test.Error.Type, test.Error.Message, test.Error.Details)
			}

			html += `
        </div>`
		}

		html += `
    </div>`
	}

	// Add performance section
	if len(report.Performance.Benchmarks) > 0 {
		html += `
    <div class="performance">
        <h2>Performance Benchmarks</h2>`

		for _, benchmark := range report.Performance.Benchmarks {
			html += fmt.Sprintf(`
        <div class="benchmark">
            <span>%s</span>
            <span>%d ops, %.2f ns/op</span>
        </div>`, benchmark.Name, benchmark.Operations, float64(benchmark.NsPerOp))
		}

		html += `
    </div>`
	}

	html += `
</body>
</html>`

	return html
}

// buildMarkdownReport builds Markdown report content
func (g *TestReportGenerator) buildMarkdownReport(report *TestReport) string {
	md := fmt.Sprintf(`# E2E Test Report

**Generated:** %s  
**Environment:** %s  
**Version:** %s  

## Summary

| Metric | Value |
|--------|-------|
| Total Tests | %d |
| Passed | %d |
| Failed | %d |
| Skipped | %d |
| Success Rate | %.1f%% |
| Duration | %s |

`,
		report.Metadata.GeneratedAt.Format(time.RFC3339),
		report.Summary.Environment,
		report.Summary.Version,
		report.Summary.TotalTests,
		report.Summary.PassedTests,
		report.Summary.FailedTests,
		report.Summary.SkippedTests,
		report.Summary.SuccessRate,
		report.Summary.Duration.String())

	// Add test suites
	for _, suite := range report.Suites {
		md += fmt.Sprintf("\n## %s\n\n", suite.Name)
		md += fmt.Sprintf("**Duration:** %s  \n", suite.Duration.String())
		md += fmt.Sprintf("**Tests:** %d  \n\n", len(suite.Tests))

		for _, test := range suite.Tests {
			status := "✅"
			if test.Status == TestStatusFailed || test.Status == TestStatusError {
				status = "❌"
			} else if test.Status == TestStatusSkipped {
				status = "⏭️"
			}

			md += fmt.Sprintf("- %s **%s** (%s)\n", status, test.Name, test.Duration.String())

			if test.Error != nil {
				md += fmt.Sprintf("  - **Error:** %s: %s\n", test.Error.Type, test.Error.Message)
			}
		}
	}

	// Add performance section
	if len(report.Performance.Benchmarks) > 0 {
		md += "\n## Performance Benchmarks\n\n"
		md += "| Benchmark | Operations | ns/op | MB/s |\n"
		md += "|-----------|------------|-------|------|\n"

		for _, benchmark := range report.Performance.Benchmarks {
			md += fmt.Sprintf("| %s | %d | %.2f | %.2f |\n",
				benchmark.Name, benchmark.Operations, float64(benchmark.NsPerOp), benchmark.MBPerSec)
		}
	}

	return md
}

// buildJUnitXML builds JUnit XML report content
func (g *TestReportGenerator) buildJUnitXML(report *TestReport) string {
	xml := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	xml += fmt.Sprintf(`<testsuites name="E2E Tests" tests="%d" failures="%d" time="%.3f">`,
		report.Summary.TotalTests, report.Summary.FailedTests, report.Summary.Duration.Seconds())

	for _, suite := range report.Suites {
		failures := 0
		for _, test := range suite.Tests {
			if test.Status == TestStatusFailed || test.Status == TestStatusError {
				failures++
			}
		}

		xml += fmt.Sprintf(`
  <testsuite name="%s" tests="%d" failures="%d" time="%.3f">`,
			suite.Name, len(suite.Tests), failures, suite.Duration.Seconds())

		for _, test := range suite.Tests {
			xml += fmt.Sprintf(`
    <testcase name="%s" classname="%s" time="%.3f">`,
				test.Name, suite.Name, test.Duration.Seconds())

			if test.Status == TestStatusFailed || test.Status == TestStatusError {
				xml += fmt.Sprintf(`
      <failure message="%s" type="%s">%s</failure>`,
					test.Error.Message, test.Error.Type, test.Error.Details)
			} else if test.Status == TestStatusSkipped {
				xml += `
      <skipped/>`
			}

			xml += `
    </testcase>`
		}

		xml += `
  </testsuite>`
	}

	xml += `
</testsuites>`

	return xml
}

// getEnvironment returns the current environment
func (g *TestReportGenerator) getEnvironment() string {
	if env := os.Getenv("CI_ENVIRONMENT"); env != "" {
		return env
	}
	if os.Getenv("CI") == "true" {
		return "CI"
	}
	return "local"
}

// getVersion returns the current version
func (g *TestReportGenerator) getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	if commit := os.Getenv("GIT_COMMIT"); commit != "" {
		return commit[:8]
	}
	return "dev"
}

// AddArtifact adds an artifact to the report
func (g *TestReportGenerator) AddArtifact(name, artifactType, path, description string, tags []string) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	artifact := Artifact{
		Name:        name,
		Type:        artifactType,
		Path:        path,
		Size:        stat.Size(),
		Description: description,
		Tags:        tags,
		CreatedAt:   time.Now(),
	}

	g.artifacts = append(g.artifacts, artifact)
}

// FailureAnalyzer analyzes test failures and provides insights
type FailureAnalyzer struct {
	report *TestReport
}

// NewFailureAnalyzer creates a new failure analyzer
func NewFailureAnalyzer(report *TestReport) *FailureAnalyzer {
	return &FailureAnalyzer{report: report}
}

// AnalyzeFailures analyzes test failures and returns insights
func (f *FailureAnalyzer) AnalyzeFailures() *FailureAnalysis {
	analysis := &FailureAnalysis{
		TotalFailures: f.report.Summary.FailedTests,
		Categories:    make(map[string][]TestFailure),
		Patterns:      make([]FailurePattern, 0),
		Suggestions:   make([]string, 0),
	}

	// Categorize failures
	for _, suite := range f.report.Suites {
		for _, test := range suite.Tests {
			if test.Status == TestStatusFailed || test.Status == TestStatusError {
				failure := TestFailure{
					Suite:       suite.Name,
					Test:        test.Name,
					Error:       test.Error,
					Duration:    test.Duration,
					Frequency:   1, // This would be populated from historical data
				}

				category := f.categorizeFailure(test.Error)
				analysis.Categories[category] = append(analysis.Categories[category], failure)
			}
		}
	}

	// Identify patterns
	analysis.Patterns = f.identifyPatterns(analysis.Categories)

	// Generate suggestions
	analysis.Suggestions = f.generateSuggestions(analysis.Categories, analysis.Patterns)

	return analysis
}

// FailureAnalysis represents the analysis of test failures
type FailureAnalysis struct {
	TotalFailures int                         `json:"total_failures"`
	Categories    map[string][]TestFailure    `json:"categories"`
	Patterns      []FailurePattern            `json:"patterns"`
	Suggestions   []string                    `json:"suggestions"`
}

// TestFailure represents a test failure
type TestFailure struct {
	Suite       string     `json:"suite"`
	Test        string     `json:"test"`
	Error       *TestError `json:"error"`
	Duration    time.Duration `json:"duration"`
	Frequency   int        `json:"frequency"`
}

// FailurePattern represents a pattern in test failures
type FailurePattern struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Frequency   int     `json:"frequency"`
	Severity    string  `json:"severity"`
	Examples    []string `json:"examples"`
}

// categorizeFailure categorizes a test failure
func (f *FailureAnalyzer) categorizeFailure(error *TestError) string {
	if error == nil {
		return "unknown"
	}

	message := strings.ToLower(error.Message)
	errorType := strings.ToLower(error.Type)

	// Database-related failures
	if strings.Contains(message, "database") || strings.Contains(message, "sql") || 
	   strings.Contains(message, "connection") {
		return "database"
	}

	// Network-related failures
	if strings.Contains(message, "network") || strings.Contains(message, "timeout") ||
	   strings.Contains(message, "connection refused") {
		return "network"
	}

	// Authentication/Authorization failures
	if strings.Contains(message, "auth") || strings.Contains(message, "permission") ||
	   strings.Contains(message, "unauthorized") {
		return "auth"
	}

	// Concurrency issues
	if strings.Contains(message, "race") || strings.Contains(message, "deadlock") ||
	   strings.Contains(message, "concurrent") {
		return "concurrency"
	}

	// Resource exhaustion
	if strings.Contains(message, "memory") || strings.Contains(message, "disk") ||
	   strings.Contains(message, "resource") || strings.Contains(message, "limit") {
		return "resources"
	}

	// Configuration issues
	if strings.Contains(message, "config") || strings.Contains(message, "setting") ||
	   strings.Contains(message, "environment") {
		return "configuration"
	}

	// Assertion failures
	if strings.Contains(errorType, "assertion") || strings.Contains(message, "expected") {
		return "assertion"
	}

	return "other"
}

// identifyPatterns identifies patterns in test failures
func (f *FailureAnalyzer) identifyPatterns(categories map[string][]TestFailure) []FailurePattern {
	patterns := make([]FailurePattern, 0)

	for category, failures := range categories {
		if len(failures) > 1 {
			pattern := FailurePattern{
				Type:        category,
				Description: fmt.Sprintf("Multiple %s failures detected", category),
				Frequency:   len(failures),
				Severity:    f.calculateSeverity(len(failures)),
				Examples:    make([]string, 0),
			}

			// Add examples
			for i, failure := range failures {
				if i < 3 { // Limit to 3 examples
					pattern.Examples = append(pattern.Examples, 
						fmt.Sprintf("%s: %s", failure.Test, failure.Error.Message))
				}
			}

			patterns = append(patterns, pattern)
		}
	}

	// Sort patterns by frequency
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Frequency > patterns[j].Frequency
	})

	return patterns
}

// generateSuggestions generates suggestions based on failure analysis
func (f *FailureAnalyzer) generateSuggestions(categories map[string][]TestFailure, patterns []FailurePattern) []string {
	suggestions := make([]string, 0)

	for _, pattern := range patterns {
		switch pattern.Type {
		case "database":
			suggestions = append(suggestions, 
				"Consider reviewing database connection configuration and retry logic")
		case "network":
			suggestions = append(suggestions, 
				"Check network timeouts and retry mechanisms for external API calls")
		case "concurrency":
			suggestions = append(suggestions, 
				"Review concurrent access patterns and consider adding proper synchronization")
		case "resources":
			suggestions = append(suggestions, 
				"Monitor resource usage and consider increasing limits or optimizing usage")
		case "assertion":
			suggestions = append(suggestions, 
				"Review test assertions and expected vs actual values")
		}
	}

	// General suggestions
	if f.report.Summary.FailedTests > f.report.Summary.TotalTests/4 {
		suggestions = append(suggestions, 
			"High failure rate detected. Consider reviewing test environment setup")
	}

	return suggestions
}

// calculateSeverity calculates the severity of a failure pattern
func (f *FailureAnalyzer) calculateSeverity(frequency int) string {
	if frequency >= 5 {
		return "critical"
	} else if frequency >= 3 {
		return "high"
	} else if frequency >= 2 {
		return "medium"
	}
	return "low"
}