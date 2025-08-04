# Testing Guide for Auto-Devs

This document provides comprehensive guidance on testing practices, infrastructure, and tools used in the Auto-Devs project.

## Table of Contents

- [Overview](#overview)
- [Testing Infrastructure](#testing-infrastructure)
- [Test Types](#test-types)
- [Running Tests](#running-tests)
- [Test Data Management](#test-data-management)
- [Code Coverage](#code-coverage)
- [API Testing](#api-testing)
- [WebSocket Testing](#websocket-testing)
- [Performance Testing](#performance-testing)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The Auto-Devs project implements a comprehensive testing strategy with multiple layers:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions and API endpoints
- **Database Tests**: Test database operations and data integrity
- **WebSocket Tests**: Test real-time communication features
- **API Tests**: Automated end-to-end API testing with Newman/Postman
- **Performance Tests**: Load and stress testing

### Testing Goals

- **Coverage**: Maintain >80% code coverage
- **Reliability**: All tests should pass consistently
- **Performance**: Tests should run quickly and efficiently
- **Maintainability**: Tests should be easy to understand and maintain

## Testing Infrastructure

### Test Database Setup

We use **Testcontainers** for isolated database testing:

```go
// Example from internal/testutil/database.go
container, cleanup := testutil.SetupTestDB(t)
defer cleanup()

// Use container.DB for database operations
repo := postgres.NewProjectRepository(container.DB)
```

### Key Components

1. **TestContainers**: PostgreSQL containers for integration tests
2. **Test Utilities**: Helper functions and factories in `internal/testutil/`
3. **Mock Objects**: Generated mocks for external dependencies
4. **Test Data Factories**: Consistent test data generation

## Test Types

### 1. Unit Tests

Located alongside source code with `_test.go` suffix.

**Example: Repository Unit Test**
```go
func TestProjectRepository_Create(t *testing.T) {
    container, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    repo := NewProjectRepository(container.DB)
    project := testutil.NewProjectFactory().CreateProject()
    
    err := repo.Create(context.Background(), project)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, project.ID)
}
```

**Example: Usecase Unit Test**
```go
func TestProjectUsecase_Create(t *testing.T) {
    mockRepo := &testutil.MockProjectRepository{}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    usecase := NewProjectUsecase(mockRepo)
    result, err := usecase.Create(ctx, createRequest)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 2. Integration Tests

Test complete request/response flows:

```go
func TestIntegration_ProjectCRUD(t *testing.T) {
    suite, cleanup := SetupIntegrationTestSuite(t)
    defer cleanup()
    
    // Test project creation
    w := suite.apiHelper.MakeRequest("POST", "/api/v1/projects", createReq)
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### 3. Database Integration Tests

Test database constraints, transactions, and concurrency:

```go
func TestDatabaseIntegration_TransactionHandling(t *testing.T) {
    container, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    
    tx := container.GormDB.Begin()
    // ... perform operations in transaction
    tx.Rollback()
    
    // Verify rollback worked
}
```

### 4. WebSocket Tests

Test real-time communication:

```go
func TestWebSocket_MessageBroadcasting(t *testing.T) {
    suite, cleanup := SetupWebSocketTestSuite(t)
    defer cleanup()
    
    conn := suite.connectWebSocket(t)
    defer conn.Close()
    
    // Test message broadcasting
    suite.hub.BroadcastToAll(message, nil)
    
    // Verify message received
    var receivedMessage Message
    err := conn.ReadJSON(&receivedMessage)
    assert.NoError(t, err)
}
```

## Running Tests

### Basic Test Commands (via Makefile)

```bash
# Run all tests
make test

# Run unit tests only (short mode)
make test-unit

# Run integration tests only
make test-integration

# Run tests with race detection
make test-race

# Run tests with coverage
make test-coverage

# Run benchmark tests
make test-bench

# Clean test cache
make test-clean
```

### Advanced Test Options

```bash
# Run specific package tests
go test ./internal/usecase -v

# Run specific test function
go test ./internal/usecase -run TestProjectUsecase_Create -v

# Run tests with timeout
go test ./... -timeout 30s

# Run tests multiple times
go test ./... -count 5

# Run tests with custom build tags
go test ./... -tags integration
```

### Environment Variables

```bash
# Set coverage threshold
export COVERAGE_THRESHOLD=85

# Set test database configuration
export DB_HOST=localhost
export DB_PORT=5432
export TEST_DB_NAME=autodevs_test

# Skip slow tests
export SKIP_INTEGRATION_TESTS=true
```

## Test Data Management

### Test Factories

Use test factories for consistent data creation:

```go
// Create a project with default values
projectFactory := testutil.NewProjectFactory()
project := projectFactory.CreateProject()

// Create a project with custom values
project := projectFactory.CreateProject(func(p *entity.Project) {
    p.Name = "Custom Project"
    p.Description = "Custom description"
})

// Create project with tasks
project, tasks := projectFactory.CreateProjectWithTasks(5)
```

### Mock Objects

Use mocks for external dependencies:

```go
mockRepo := &testutil.MockProjectRepository{}
mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
mockRepo.On("GetByID", mock.Anything, projectID).Return(project, nil)

// Use in tests
usecase := NewProjectUsecase(mockRepo)
```

### Database Cleanup

Tests automatically clean up using:

1. **Isolated containers**: Each test gets fresh database
2. **Transaction rollback**: For unit tests
3. **Truncate tables**: For integration tests

```go
// Automatic cleanup with container
container, cleanup := testutil.SetupTestDB(t)
defer cleanup() // Automatically destroys container

// Manual table cleanup
container.TruncateTables(t)
```

## Code Coverage

### Coverage Reports

Generate coverage reports:

```bash
# Generate coverage with HTML report
make test-coverage

# View function-level coverage
make test-coverage-func

# Open HTML report (macOS)
open coverage.html
```

### Coverage Targets

- **Overall**: >80% coverage
- **Critical paths**: >90% coverage
- **New code**: 100% coverage

### Viewing Coverage

1. **Terminal**: Function-level coverage summary
2. **HTML Report**: Interactive coverage visualization
3. **IDE Integration**: In-editor coverage highlighting

## API Testing

### Newman/Postman Integration

We use Newman for automated API testing:

```bash
# Run API tests (requires server running)
./tests/scripts/api-tests.sh

# Run with HTML report
./tests/scripts/api-tests.sh --open

# Run performance tests
./tests/scripts/api-tests.sh --performance

# Run load tests
./tests/scripts/api-tests.sh --load
```

### Test Collection Structure

```
tests/postman/
├── auto-devs-collection.json  # Main test collection
├── environment.json           # Test environment variables
└── reports/                   # Generated reports
    ├── api-tests-report.html
    ├── api-tests-junit.xml
    └── api-tests-results.json
```

### Writing API Tests

Add new tests to the Postman collection:

1. **Create request** in appropriate folder
2. **Add test scripts** for assertions
3. **Set environment variables** for dynamic data
4. **Update collection** with proper test logic

Example test script:
```javascript
pm.test('Status code is 201', function () {
    pm.response.to.have.status(201);
});

pm.test('Response has required fields', function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('id');
    pm.expect(jsonData).to.have.property('name');
});
```

## WebSocket Testing

### Test Suite Setup

```go
suite, cleanup := SetupWebSocketTestSuite(t)
defer cleanup()

// Connect test clients
conn1 := suite.connectWebSocket(t)
conn2 := suite.connectWebSocket(t)
defer conn1.Close()
defer conn2.Close()
```

### Testing Scenarios

1. **Connection handling**: Connect, disconnect, error scenarios
2. **Message broadcasting**: All clients, project-specific, excluding sender
3. **Subscription management**: Subscribe/unsubscribe to projects
4. **Rate limiting**: Message throttling and limits
5. **Concurrent operations**: Multiple clients, simultaneous messages

## Performance Testing

### Benchmark Tests

Write benchmark tests for critical paths:

```go
func BenchmarkProjectRepository_Create(b *testing.B) {
    container, cleanup := testutil.SetupTestDB(b)
    defer cleanup()
    
    repo := NewProjectRepository(container.DB)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        project := &entity.Project{
            Name:    fmt.Sprintf("Benchmark Project %d", i),
            RepoURL: "https://github.com/test/benchmark.git",
        }
        repo.Create(context.Background(), project)
    }
}
```

### Load Testing

```bash
# Run API load tests
./tests/scripts/api-tests.sh --load

# Run concurrent database operations
go test ./internal/repository -run TestConcurrent -v
```

### Performance Metrics

Monitor:
- **Response times**: API endpoint latency
- **Throughput**: Requests per second
- **Database performance**: Query execution time
- **Memory usage**: Memory allocation patterns
- **WebSocket performance**: Message delivery time

## Best Practices

### Test Organization

```
internal/
├── handler/
│   ├── project.go
│   ├── project_test.go          # Unit tests
│   └── integration_test.go      # Integration tests
├── usecase/
│   ├── project.go
│   └── project_test.go
└── repository/
    ├── postgres/
    │   ├── project_repository.go
    │   ├── project_repository_test.go
    │   └── database_integration_test.go
```

### Test Naming

- **Unit tests**: `Test[Type]_[Method]`
- **Integration tests**: `TestIntegration_[Feature]`
- **Database tests**: `TestDatabaseIntegration_[Scenario]`
- **Benchmarks**: `Benchmark[Type]_[Method]`

### Test Structure

Use the **Arrange-Act-Assert** pattern:

```go
func TestProjectUsecase_Create(t *testing.T) {
    // Arrange
    mockRepo := &testutil.MockProjectRepository{}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    usecase := NewProjectUsecase(mockRepo)
    
    // Act
    result, err := usecase.Create(ctx, request)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Error Testing

Always test error scenarios:

```go
t.Run("repository error", func(t *testing.T) {
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
    
    result, err := usecase.Create(ctx, request)
    
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "db error")
})
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestValidateInput(t *testing.T) {
    testCases := []struct {
        name      string
        input     string
        shouldErr bool
        errorMsg  string
    }{
        {"valid input", "valid", false, ""},
        {"empty input", "", true, "required"},
        {"invalid format", "invalid", true, "format"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateInput(tc.input)
            
            if tc.shouldErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Test Database Connection Issues

```bash
# Check if database container is running
docker ps | grep postgres

# Check database logs
docker logs <container_id>

# Restart containers
make test-clean
```

#### 2. Flaky Tests

- **Use deterministic test data**: Avoid random values in assertions
- **Add proper timeouts**: For async operations
- **Clean state between tests**: Reset global state
- **Avoid test dependencies**: Each test should be independent

#### 3. Slow Tests

```bash
# Profile test execution
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...

# Skip slow tests in development
go test -short ./...

# Run specific package only
go test ./internal/usecase
```

#### 4. Coverage Issues

- **Missing test files**: Ensure all packages have test files
- **Unreachable code**: Check for dead code paths
- **Interface implementations**: Test all interface methods
- **Error paths**: Test error handling code

### Debug Commands

```bash
# Verbose test output
go test -v ./...

# Run with race detector
go test -race ./...

# Debug specific test
go test -run TestSpecificTest -v

# Print test coverage by function
go test -coverprofile=coverage.out -covermode=count ./...
go tool cover -func=coverage.out

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Getting Help

1. **Check test logs**: Look for specific error messages
2. **Run tests individually**: Isolate failing tests
3. **Check test data**: Verify test setup and expectations
4. **Review recent changes**: Check for breaking changes
5. **Consult documentation**: Review API documentation and examples

## Continuous Integration

### GitHub Actions Integration

Tests run automatically on:
- **Pull requests**: All test suites
- **Main branch**: Full test suite with coverage
- **Releases**: Complete test suite including performance tests

### Local Pre-commit Hooks

Set up pre-commit hooks:

```bash
# Install pre-commit hooks
git config core.hooksPath .githooks

# Make hooks executable
chmod +x .githooks/pre-commit
```

The pre-commit hook runs:
1. Unit tests
2. Linting
3. Code formatting
4. Coverage check

---

## Summary

This testing infrastructure provides:

- **Comprehensive coverage**: Unit, integration, and end-to-end tests
- **Reliable infrastructure**: Testcontainers for consistent environments
- **Performance monitoring**: Benchmarks and load tests
- **Automated testing**: CI/CD integration with reporting
- **Developer tools**: Easy-to-use test utilities and helpers

Follow the practices outlined in this guide to maintain high code quality and reliable tests throughout the development lifecycle.