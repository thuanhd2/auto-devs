# End-to-End Testing Framework

This package provides comprehensive end-to-end testing for the AI automation workflow in the Auto-Devs project.

## Overview

The E2E testing framework tests the complete AI automation workflow from task creation to pull request merge, including:

- **Task Management**: Complete task lifecycle from TODO to DONE
- **AI Planning**: Plan generation, review, and approval
- **AI Implementation**: Code implementation via AI CLI tools
- **Git Integration**: Worktree management and branch operations
- **GitHub Integration**: Pull request creation and monitoring
- **WebSocket Communication**: Real-time updates and notifications

## Architecture

### Test Infrastructure (`test_infrastructure.go`)

- **E2ETestSuite**: Main test suite that sets up complete test environment
- **TestRepositories**: Repository layer for data access
- **TestServices**: Service layer with mock implementations
- **TestHandlers**: HTTP handlers for API testing

### Mock Services (`mocks.go`)

- **GitManagerMock**: Mocks Git operations (clone, branch, checkout)
- **WorktreeServiceMock**: Mocks Git worktree operations
- **GitHubServiceMock**: Mocks GitHub API operations
- **AIServiceMock**: Mocks AI planning and implementation services
- **ProcessManagerMock**: Mocks process management operations

### Test Data Generation (`test_data.go`)

- **TestDataGenerator**: Generates realistic test data
- **Configuration Structs**: Configurable test data creation
- **CompleteTaskFlow**: Creates complete workflows for testing

### Test Categories

#### 1. Happy Path Tests (`happy_path_test.go`)

- **TestCompleteTaskAutomationFlow**: Full task lifecycle
- **TestMultiTaskProjectWorkflow**: Multiple concurrent tasks
- **TestPlanGenerationAndApproval**: Planning workflow
- **TestImplementationExecution**: Implementation workflow  
- **TestPRCreationAndMonitoring**: Pull request workflow
- **TestWebSocketRealTimeUpdates**: Real-time notifications

#### 2. Error Scenarios (`error_scenarios_test.go`)

- **TestPlanningServiceFailure**: AI service failures
- **TestImplementationServiceFailure**: Implementation failures
- **TestGitOperationFailures**: Git-related failures
- **TestGitHubAPIFailures**: GitHub API issues
- **TestDatabaseConnectionFailures**: Database connectivity
- **TestConcurrencyIssues**: Race conditions and concurrent access
- **TestResourceExhaustion**: Resource constraint handling

#### 3. Edge Cases & Performance (`edge_cases_performance_test.go`)

- **TestEdgeCasesTaskStates**: Invalid state transitions
- **TestLargeDataHandling**: Large datasets and content
- **TestHighVolumeTaskProcessing**: High load scenarios
- **TestWebSocketPerformance**: WebSocket scalability
- **TestSystemLimits**: System boundary testing
- **TestDataConsistency**: Data integrity under stress

### Test Reporting (`test_reporting.go`)

- **TestReportGenerator**: Comprehensive test reports
- **FailureAnalyzer**: Automated failure analysis
- **Multiple Formats**: HTML, Markdown, JSON, JUnit XML
- **Performance Metrics**: Detailed performance analysis

### Utilities (`test_utils.go`)

- **TestUtils**: General testing utilities
- **HTTPClient**: HTTP client for API testing
- **DatabaseHelper**: Database testing utilities
- **TaskTestHelper**: Task-specific test helpers
- **PerformanceTestHelper**: Performance measurement
- **LoadTestHelper**: Load testing capabilities

## Test Scenarios

### Core Workflows

1. **Complete Task Automation**
   ```
   TODO → PLANNING → PLAN_REVIEWING → IMPLEMENTING → CODE_REVIEWING → DONE
   ```

2. **Plan Generation and Approval**
   - AI generates detailed implementation plan
   - Human reviews and approves/rejects plan
   - Plan revision cycles

3. **Implementation Execution**
   - Git worktree creation
   - AI CLI tool execution
   - Code generation and testing
   - Artifact creation

4. **Pull Request Workflow**
   - Automatic PR creation
   - GitHub webhook processing
   - PR merge detection
   - Task completion

### Error Handling

- **Service Failures**: AI services unavailable
- **Network Issues**: GitHub API rate limits, timeouts
- **Git Failures**: Repository access, branch conflicts
- **Database Issues**: Connection loss, deadlocks
- **Resource Constraints**: Memory, disk, concurrent limits

### Performance Testing

- **High Volume**: 1000+ tasks, concurrent processing
- **WebSocket Scale**: 50+ concurrent connections
- **Memory Usage**: Memory leak detection
- **Query Performance**: Database optimization
- **Throughput**: Tasks per second metrics

## Usage

### Running Tests Locally

```bash
# Run all E2E tests
./scripts/run-e2e-tests.sh

# Run specific test suite
./scripts/run-e2e-tests.sh -s happy-path

# Run with verbose output
./scripts/run-e2e-tests.sh -s error-scenarios -v

# Run performance tests
./scripts/run-e2e-tests.sh -s performance -t 45m
```

### Running in Go

```go
// Run specific test
go test -v ./internal/e2e_test -run TestCompleteTaskAutomationFlow

// Run test suite with timeout
go test -timeout 30m ./internal/e2e_test -run TestHappyPath

// Run benchmarks
go test -bench=. ./internal/e2e_test
```

### CI/CD Integration

Tests run automatically in GitHub Actions:

- **Pull Requests**: Happy path + error scenarios
- **Main Branch**: Full test suite
- **Nightly**: Complete suite with performance tests
- **Manual**: Configurable test selection

## Configuration

### Environment Variables

```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=autodevs_test

# Redis configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Test configuration
TEST_TIMEOUT=30m
LOG_LEVEL=debug
PARALLEL=4
```

### Test Data Configuration

```go
// Project configuration
projectConfig := ProjectConfig{
    Name:          "Test Project",
    Description:   "E2E test project",
    RepositoryURL: "https://github.com/test/repo.git",
    DefaultBranch: "main",
    AutoMerge:     true,
}

// Task configuration
taskConfig := TaskConfig{
    Title:       "Test Task",
    Description: "E2E test task",
    Status:      entity.TaskStatusTODO,
    Priority:    entity.TaskPriorityHigh,
}
```

## Reports and Analysis

### Report Generation

Tests generate comprehensive reports in multiple formats:

- **HTML**: Interactive web report with charts
- **Markdown**: Human-readable summary
- **JSON**: Machine-readable detailed data
- **JUnit XML**: CI/CD integration format

### Failure Analysis

Automated failure analysis provides:

- **Categorization**: Database, network, auth, concurrency, etc.
- **Pattern Detection**: Recurring failure patterns
- **Root Cause Analysis**: Potential causes and solutions
- **Trend Analysis**: Historical failure patterns

### Performance Metrics

Detailed performance analysis includes:

- **Throughput**: Operations per second
- **Latency**: Response time percentiles
- **Resource Usage**: Memory, CPU, disk, network
- **Scalability**: Performance under load
- **Benchmarks**: Comparative performance data

## Best Practices

### Writing E2E Tests

1. **Test Real Scenarios**: Mirror actual user workflows
2. **Use Realistic Data**: Generate representative test data
3. **Include Error Cases**: Test failure scenarios
4. **Measure Performance**: Track key metrics
5. **Clean Up Resources**: Proper teardown and cleanup

### Mock Configuration

1. **Realistic Responses**: Mock services should behave like real ones
2. **Error Simulation**: Include failure scenarios in mocks
3. **Async Behavior**: Simulate real timing and delays
4. **State Management**: Maintain mock state consistency

### Test Organization

1. **Logical Grouping**: Group related tests in suites
2. **Independent Tests**: Tests should not depend on each other
3. **Clear Naming**: Descriptive test and suite names
4. **Documentation**: Comment complex test scenarios

## Troubleshooting

### Common Issues

1. **Database Connection**: Ensure PostgreSQL is running
2. **Port Conflicts**: Check for port availability
3. **Memory Issues**: Increase test timeouts for large datasets
4. **Mock Failures**: Verify mock expectations are set correctly

### Debug Options

```bash
# Verbose output
VERBOSE=true ./scripts/run-e2e-tests.sh

# Skip cleanup for inspection
CLEANUP=false ./scripts/run-e2e-tests.sh

# Extended timeout
TEST_TIMEOUT=60m ./scripts/run-e2e-tests.sh
```

### Log Analysis

Check test logs for:
- Database connection errors
- Mock expectation failures
- Timeout issues
- Resource exhaustion
- Assertion failures

## Contributing

### Adding New Tests

1. Choose appropriate test file based on category
2. Follow existing naming conventions
3. Use test utilities and helpers
4. Include proper cleanup
5. Add documentation

### Extending Mock Services

1. Add new methods to existing mocks
2. Implement realistic behavior
3. Include error scenarios
4. Update mock expectations in helper functions

### Performance Testing

1. Use performance test helpers
2. Set appropriate thresholds
3. Monitor resource usage
4. Include baseline comparisons

For more details, see individual source files and inline documentation.