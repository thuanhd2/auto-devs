package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// APITestHelper provides utilities for API testing
type APITestHelper struct {
	router *gin.Engine
}

// NewAPITestHelper creates a new APITestHelper
func NewAPITestHelper(router *gin.Engine) *APITestHelper {
	return &APITestHelper{
		router: router,
	}
}

// MakeRequest makes an HTTP request and returns the response
func (h *APITestHelper) MakeRequest(method, url string, body interface{}, headers ...map[string]string) *httptest.ResponseRecorder {
	var bodyStr string
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyStr = string(bodyBytes)
	}

	req, _ := http.NewRequest(method, url, strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)

	return w
}

// AssertJSONResponse asserts that the response has the expected status and JSON body
func (h *APITestHelper) AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(t, expectedStatus, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	if expectedBody != nil {
		expectedJSON, err := json.Marshal(expectedBody)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedJSON), w.Body.String())
	}
}

// AssertErrorResponse asserts that the response is an error with expected status and message
func (h *APITestHelper) AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(t, expectedStatus, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], expectedMessage)
}

// ParseJSONResponse parses the JSON response into the provided struct
func (h *APITestHelper) ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, dest interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), dest)
	require.NoError(t, err)
}

// ValidationTestCase represents a test case for validation testing
type ValidationTestCase struct {
	Name           string
	Body           interface{}
	ExpectedStatus int
	ExpectedError  string
}

// RunValidationTests runs a series of validation test cases
func (h *APITestHelper) RunValidationTests(t *testing.T, method, url string, testCases []ValidationTestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			w := h.MakeRequest(method, url, tc.Body)
			
			if tc.ExpectedError != "" {
				h.AssertErrorResponse(t, w, tc.ExpectedStatus, tc.ExpectedError)
			} else {
				assert.Equal(t, tc.ExpectedStatus, w.Code)
			}
		})
	}
}

// TestLogger provides a test-friendly logger
type TestLogger struct {
	logs []string
}

// NewTestLogger creates a new TestLogger
func NewTestLogger() *TestLogger {
	return &TestLogger{
		logs: make([]string, 0),
	}
}

// Log adds a log entry
func (l *TestLogger) Log(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}

// GetLogs returns all log entries
func (l *TestLogger) GetLogs() []string {
	return l.logs
}

// Clear clears all log entries
func (l *TestLogger) Clear() {
	l.logs = make([]string, 0)
}

// AssertionHelper provides custom assertions for testing
type AssertionHelper struct{}

// NewAssertionHelper creates a new AssertionHelper
func NewAssertionHelper() *AssertionHelper {
	return &AssertionHelper{}
}

// AssertTimeEqual asserts that two times are equal within a tolerance
func (h *AssertionHelper) AssertTimeEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	// Implementation would depend on your time type
	// This is a placeholder for time comparison logic
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertContainsAllKeys asserts that a map contains all expected keys
func (h *AssertionHelper) AssertContainsAllKeys(t *testing.T, m map[string]interface{}, expectedKeys []string, msgAndArgs ...interface{}) {
	for _, key := range expectedKeys {
		assert.Contains(t, m, key, msgAndArgs...)
	}
}

// AssertValidUUID asserts that a string is a valid UUID
func (h *AssertionHelper) AssertValidUUID(t *testing.T, str string, msgAndArgs ...interface{}) {
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, str, msgAndArgs...)
}

// DatabaseTestHelper provides utilities for database testing
type DatabaseTestHelper struct {
	container *TestContainer
}

// NewDatabaseTestHelper creates a new DatabaseTestHelper
func NewDatabaseTestHelper(container *TestContainer) *DatabaseTestHelper {
	return &DatabaseTestHelper{
		container: container,
	}
}

// CountRecords counts the number of records in a table
func (h *DatabaseTestHelper) CountRecords(t *testing.T, tableName string) int64 {
	var count int64
	err := h.container.GormDB.Table(tableName).Count(&count).Error
	require.NoError(t, err)
	return count
}

// AssertRecordExists asserts that a record exists with the given condition
func (h *DatabaseTestHelper) AssertRecordExists(t *testing.T, tableName string, condition string, args ...interface{}) {
	var count int64
	err := h.container.GormDB.Table(tableName).Where(condition, args...).Count(&count).Error
	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Expected record to exist but found none")
}

// AssertRecordNotExists asserts that no record exists with the given condition
func (h *DatabaseTestHelper) AssertRecordNotExists(t *testing.T, tableName string, condition string, args ...interface{}) {
	var count int64
	err := h.container.GormDB.Table(tableName).Where(condition, args...).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Expected no records but found some")
}

// Performance testing utilities

// BenchmarkHelper provides utilities for performance testing
type BenchmarkHelper struct{}

// NewBenchmarkHelper creates a new BenchmarkHelper
func NewBenchmarkHelper() *BenchmarkHelper {
	return &BenchmarkHelper{}
}

// BenchmarkOperation benchmarks an operation and returns the duration
func (h *BenchmarkHelper) BenchmarkOperation(operation func()) float64 {
	// This would contain timing logic for performance testing
	// Implementation depends on your performance requirements
	operation()
	return 0.0 // Placeholder
}

// SetupGinTestMode configures Gin for testing
func SetupGinTestMode() {
	gin.SetMode(gin.TestMode)
}

// TeardownGinTestMode resets Gin mode after testing
func TeardownGinTestMode() {
	gin.SetMode(gin.ReleaseMode)
}