package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/stretchr/testify/require"
)

// TestUtils provides utility functions for E2E tests
type TestUtils struct {
	suite *E2ETestSuite
}

// NewTestUtils creates a new test utilities instance
func NewTestUtils(suite *E2ETestSuite) *TestUtils {
	return &TestUtils{suite: suite}
}

// WaitForCondition waits for a condition to be true within timeout
func (u *TestUtils) WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) bool {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if condition() {
				return true
			}
			if time.Now().After(deadline) {
				t.Logf("Condition timeout: %s", message)
				return false
			}
		}
	}
}

// AssertEventuallyTrue asserts that a condition becomes true within timeout
func (u *TestUtils) AssertEventuallyTrue(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	if !u.WaitForCondition(t, condition, timeout, message) {
		t.Fatalf("Condition was not met within %v: %s", timeout, message)
	}
}

// CollectMetrics collects system metrics during test execution
func (u *TestUtils) CollectMetrics(ctx context.Context) *SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemMetrics{
		Memory: MemoryMetrics{
			HeapAlloc:    m.HeapAlloc,
			HeapSys:      m.HeapSys,
			HeapIdle:     m.HeapIdle,
			HeapInuse:    m.HeapInuse,
			TotalAlloc:   m.TotalAlloc,
			Mallocs:      m.Mallocs,
			Frees:        m.Frees,
			GCCycles:     m.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		Timestamp:  time.Now(),
	}
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	Memory     MemoryMetrics `json:"memory"`
	Goroutines int           `json:"goroutines"`
	Timestamp  time.Time     `json:"timestamp"`
}

// MemoryMetrics represents memory metrics
type MemoryMetrics struct {
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
	HeapIdle   uint64 `json:"heap_idle"`
	HeapInuse  uint64 `json:"heap_inuse"`
	TotalAlloc uint64 `json:"total_alloc"`
	Mallocs    uint64 `json:"mallocs"`
	Frees      uint64 `json:"frees"`
	GCCycles   uint32 `json:"gc_cycles"`
}

// HTTPClient provides HTTP client utilities for testing
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GET performs a GET request
func (c *HTTPClient) GET(path string, headers map[string]string) (*http.Response, error) {
	return c.request("GET", path, nil, headers)
}

// POST performs a POST request
func (c *HTTPClient) POST(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.request("POST", path, body, headers)
}

// PUT performs a PUT request
func (c *HTTPClient) PUT(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.request("PUT", path, body, headers)
}

// DELETE performs a DELETE request
func (c *HTTPClient) DELETE(path string, headers map[string]string) (*http.Response, error) {
	return c.request("DELETE", path, nil, headers)
}

// request performs an HTTP request
func (c *HTTPClient) request(method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path
	
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	var req *http.Request
	var err error
	
	if reqBody != nil {
		req, err = http.NewRequest(method, url, reqBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// DatabaseHelper provides database utilities for testing
type DatabaseHelper struct {
	suite *E2ETestSuite
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(suite *E2ETestSuite) *DatabaseHelper {
	return &DatabaseHelper{suite: suite}
}

// CleanDatabase cleans all test data from the database
func (d *DatabaseHelper) CleanDatabase(t *testing.T) {
	tables := []string{
		"execution_logs",
		"executions",
		"pull_requests",
		"plans",
		"worktrees",
		"tasks",
		"projects",
		"audit_logs",
		"processes",
	}

	for _, table := range tables {
		result := d.suite.db.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if result.Error != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, result.Error)
		}
	}
}

// GetTableRowCount returns the number of rows in a table
func (d *DatabaseHelper) GetTableRowCount(table string) int64 {
	var count int64
	d.suite.db.DB.Table(table).Count(&count)
	return count
}

// AssertTableRowCount asserts the number of rows in a table
func (d *DatabaseHelper) AssertTableRowCount(t *testing.T, table string, expected int64) {
	actual := d.GetTableRowCount(table)
	require.Equal(t, expected, actual, "Row count mismatch for table %s", table)
}

// TaskTestHelper provides task-specific test utilities
type TaskTestHelper struct {
	suite *E2ETestSuite
}

// NewTaskTestHelper creates a new task test helper
func NewTaskTestHelper(suite *E2ETestSuite) *TaskTestHelper {
	return &TaskTestHelper{suite: suite}
}

// CreateTaskFlow creates a complete task flow for testing
func (t *TaskTestHelper) CreateTaskFlow(config TaskFlowConfig) *TaskFlow {
	dataGen := NewTestDataGenerator(t.suite)
	
	// Create project
	project := dataGen.GenerateProject(config.Project)
	
	// Create task
	task := dataGen.GenerateTask(project.ID, config.Task)
	
	flow := &TaskFlow{
		Project: project,
		Task:    task,
	}
	
	// Create plan if requested
	if config.CreatePlan {
		plan := dataGen.GeneratePlan(task.ID, config.Plan)
		flow.Plan = plan
	}
	
	// Create worktree if requested
	if config.CreateWorktree {
		worktree := dataGen.GenerateWorktree(project.ID, task.ID, config.Worktree)
		flow.Worktree = worktree
	}
	
	// Create executions if requested
	for _, execConfig := range config.Executions {
		execution := dataGen.GenerateExecution(task.ID, execConfig)
		flow.Executions = append(flow.Executions, execution)
	}
	
	return flow
}

// TaskFlowConfig configures task flow creation
type TaskFlowConfig struct {
	Project        ProjectConfig
	Task           TaskConfig
	Plan           PlanConfig
	Worktree       WorktreeConfig
	Executions     []ExecutionConfig
	CreatePlan     bool
	CreateWorktree bool
}

// TaskFlow represents a complete task flow
type TaskFlow struct {
	Project    *entity.Project
	Task       *entity.Task
	Plan       *entity.Plan
	Worktree   *entity.Worktree
	Executions []*entity.Execution
}

// AdvanceTaskToStatus advances a task through statuses to reach target status
func (t *TaskTestHelper) AdvanceTaskToStatus(testInstance *testing.T, taskID uuid.UUID, targetStatus entity.TaskStatus) error {
	currentTask, err := t.suite.repositories.Task.GetByID(t.suite.ctx, taskID)
	if err != nil {
		return err
	}

	// Define status progression
	statusProgression := []entity.TaskStatus{
		entity.TaskStatusTODO,
		entity.TaskStatusPLANNING,
		entity.TaskStatusPLANREVIEWING,
		entity.TaskStatusIMPLEMENTING,
		entity.TaskStatusCODEREVIEWING,
		entity.TaskStatusDONE,
	}

	// Find current and target positions
	currentPos := -1
	targetPos := -1
	
	for i, status := range statusProgression {
		if status == currentTask.Status {
			currentPos = i
		}
		if status == targetStatus {
			targetPos = i
		}
	}

	if currentPos == -1 || targetPos == -1 {
		return fmt.Errorf("invalid status progression from %s to %s", currentTask.Status, targetStatus)
	}

	if currentPos >= targetPos {
		return fmt.Errorf("task is already at or beyond target status")
	}

	// Advance through each status
	for i := currentPos + 1; i <= targetPos; i++ {
		nextStatus := statusProgression[i]
		
		switch nextStatus {
		case entity.TaskStatusPLANNING:
			// Start planning
			t.suite.startTaskPlanning(testInstance, taskID)
		case entity.TaskStatusPLANREVIEWING:
			// Wait for planning to complete
			t.suite.WaitForTaskStatus(taskID, entity.TaskStatusPLANREVIEWING, 10*time.Second)
		case entity.TaskStatusIMPLEMENTING:
			// Approve plan
			plans, err := t.suite.repositories.Plan.List(t.suite.ctx, entity.PlanFilters{TaskID: &taskID})
			if err != nil {
				return err
			}
			if len(plans) > 0 {
				t.suite.approvePlan(testInstance, plans[0].ID)
			}
		case entity.TaskStatusCODEREVIEWING:
			// Start implementation
			t.suite.startTaskImplementation(testInstance, taskID)
			t.suite.WaitForTaskStatus(taskID, entity.TaskStatusCODEREVIEWING, 15*time.Second)
		case entity.TaskStatusDONE:
			// Complete task
			t.suite.completeTask(testInstance, taskID)
		}
		
		// Wait for status update
		if !t.suite.WaitForTaskStatus(taskID, nextStatus, 10*time.Second) {
			return fmt.Errorf("task did not reach status %s within timeout", nextStatus)
		}
	}

	return nil
}

// PerformanceTestHelper provides performance testing utilities
type PerformanceTestHelper struct {
	suite *E2ETestSuite
}

// NewPerformanceTestHelper creates a new performance test helper
func NewPerformanceTestHelper(suite *E2ETestSuite) *PerformanceTestHelper {
	return &PerformanceTestHelper{suite: suite}
}

// MeasureOperation measures the performance of an operation
func (p *PerformanceTestHelper) MeasureOperation(name string, operation func() error) *OperationMetrics {
	startTime := time.Now()
	startMetrics := p.collectSystemMetrics()
	
	err := operation()
	
	endTime := time.Now()
	endMetrics := p.collectSystemMetrics()
	
	return &OperationMetrics{
		Name:        name,
		Duration:    endTime.Sub(startTime),
		Success:     err == nil,
		Error:       err,
		StartTime:   startTime,
		EndTime:     endTime,
		MemoryDelta: endMetrics.Memory.HeapAlloc - startMetrics.Memory.HeapAlloc,
		GoroutineDelta: endMetrics.Goroutines - startMetrics.Goroutines,
	}
}

// OperationMetrics represents performance metrics for an operation
type OperationMetrics struct {
	Name           string        `json:"name"`
	Duration       time.Duration `json:"duration"`
	Success        bool          `json:"success"`
	Error          error         `json:"error,omitempty"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	MemoryDelta    uint64        `json:"memory_delta"`
	GoroutineDelta int           `json:"goroutine_delta"`
}

// collectSystemMetrics collects current system metrics
func (p *PerformanceTestHelper) collectSystemMetrics() *SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemMetrics{
		Memory: MemoryMetrics{
			HeapAlloc:  m.HeapAlloc,
			HeapSys:    m.HeapSys,
			HeapIdle:   m.HeapIdle,
			HeapInuse:  m.HeapInuse,
			TotalAlloc: m.TotalAlloc,
			Mallocs:    m.Mallocs,
			Frees:      m.Frees,
			GCCycles:   m.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		Timestamp:  time.Now(),
	}
}

// LoadTestHelper provides load testing utilities
type LoadTestHelper struct {
	suite *E2ETestSuite
}

// NewLoadTestHelper creates a new load test helper
func NewLoadTestHelper(suite *E2ETestSuite) *LoadTestHelper {
	return &LoadTestHelper{suite: suite}
}

// RunLoadTest runs a load test with specified parameters
func (l *LoadTestHelper) RunLoadTest(config LoadTestConfig) *LoadTestResults {
	results := &LoadTestResults{
		Config:      config,
		StartTime:   time.Now(),
		Operations:  make([]*OperationResult, 0),
		Errors:      make([]error, 0),
	}

	// Create worker channels
	workChan := make(chan int, config.ConcurrentUsers)
	resultChan := make(chan *OperationResult, config.TotalOperations)
	
	// Start workers
	for i := 0; i < config.ConcurrentUsers; i++ {
		go l.loadTestWorker(workChan, resultChan, config.Operation)
	}
	
	// Send work
	go func() {
		for i := 0; i < config.TotalOperations; i++ {
			workChan <- i
			if config.RampUpDuration > 0 {
				time.Sleep(config.RampUpDuration / time.Duration(config.TotalOperations))
			}
		}
		close(workChan)
	}()
	
	// Collect results
	timeout := time.After(config.TestDuration)
	completed := 0
	
	for completed < config.TotalOperations {
		select {
		case result := <-resultChan:
			results.Operations = append(results.Operations, result)
			if result.Error != nil {
				results.Errors = append(results.Errors, result.Error)
			}
			completed++
		case <-timeout:
			results.TimedOut = true
			break
		}
	}
	
	results.EndTime = time.Now()
	results.calculateSummary()
	
	return results
}

// LoadTestConfig configures load test parameters
type LoadTestConfig struct {
	ConcurrentUsers  int
	TotalOperations  int
	TestDuration     time.Duration
	RampUpDuration   time.Duration
	Operation        func() error
}

// LoadTestResults contains load test results
type LoadTestResults struct {
	Config           LoadTestConfig
	StartTime        time.Time
	EndTime          time.Time
	Operations       []*OperationResult
	Errors           []error
	TimedOut         bool
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	OperationsPerSec float64
	ErrorRate        float64
}

// OperationResult represents the result of a single operation
type OperationResult struct {
	Index     int           `json:"index"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     error         `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// loadTestWorker is a worker function for load testing
func (l *LoadTestHelper) loadTestWorker(workChan <-chan int, resultChan chan<- *OperationResult, operation func() error) {
	for index := range workChan {
		startTime := time.Now()
		err := operation()
		endTime := time.Now()
		
		result := &OperationResult{
			Index:     index,
			Duration:  endTime.Sub(startTime),
			Success:   err == nil,
			Error:     err,
			StartTime: startTime,
			EndTime:   endTime,
		}
		
		resultChan <- result
	}
}

// calculateSummary calculates summary statistics
func (r *LoadTestResults) calculateSummary() {
	r.TotalDuration = r.EndTime.Sub(r.StartTime)
	
	if len(r.Operations) == 0 {
		return
	}
	
	// Calculate duration statistics
	var totalDuration time.Duration
	r.MinDuration = r.Operations[0].Duration
	r.MaxDuration = r.Operations[0].Duration
	
	successCount := 0
	for _, op := range r.Operations {
		totalDuration += op.Duration
		if op.Success {
			successCount++
		}
		if op.Duration < r.MinDuration {
			r.MinDuration = op.Duration
		}
		if op.Duration > r.MaxDuration {
			r.MaxDuration = op.Duration
		}
	}
	
	r.AverageDuration = totalDuration / time.Duration(len(r.Operations))
	r.OperationsPerSec = float64(len(r.Operations)) / r.TotalDuration.Seconds()
	r.ErrorRate = float64(len(r.Errors)) / float64(len(r.Operations)) * 100
}

// TestStepRecorder records test steps for reporting
type TestStepRecorder struct {
	steps []TestStep
}

// NewTestStepRecorder creates a new test step recorder
func NewTestStepRecorder() *TestStepRecorder {
	return &TestStepRecorder{
		steps: make([]TestStep, 0),
	}
}

// RecordStep records a test step
func (r *TestStepRecorder) RecordStep(name, description string, operation func() error) {
	startTime := time.Now()
	err := operation()
	duration := time.Since(startTime)
	
	status := TestStatusPassed
	var testError *TestError
	
	if err != nil {
		status = TestStatusFailed
		testError = &TestError{
			Type:      "StepError",
			Message:   err.Error(),
			Details:   fmt.Sprintf("Step '%s' failed", name),
			Timestamp: time.Now(),
		}
	}
	
	step := TestStep{
		Name:        name,
		Description: description,
		Status:      status,
		Duration:    duration,
		Error:       testError,
		Timestamp:   startTime,
	}
	
	r.steps = append(r.steps, step)
}

// GetSteps returns recorded test steps
func (r *TestStepRecorder) GetSteps() []TestStep {
	return r.steps
}

// ClearSteps clears recorded test steps
func (r *TestStepRecorder) ClearSteps() {
	r.steps = make([]TestStep, 0)
}