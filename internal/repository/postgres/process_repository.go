package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type processRepository struct {
	db *database.GormDB
}

// NewProcessRepository creates a new PostgreSQL process repository
func NewProcessRepository(db *database.GormDB) repository.ProcessRepository {
	return &processRepository{db: db}
}

// Create creates a new process
func (r *processRepository) Create(ctx context.Context, process *entity.Process) error {
	// Generate UUID if not provided
	if process.ID == uuid.Nil {
		process.ID = uuid.New()
	}

	// Set default status if not provided
	if process.Status == "" {
		process.Status = entity.ProcessStatusStarting
	}

	// Set start time if not provided
	if process.StartTime.IsZero() {
		process.StartTime = time.Now()
	}

	result := r.db.WithContext(ctx).Create(process)
	if result.Error != nil {
		return fmt.Errorf("failed to create process: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a process by ID
func (r *processRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Process, error) {
	var process entity.Process

	result := r.db.WithContext(ctx).First(&process, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("process not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get process: %w", result.Error)
	}

	return &process, nil
}

// GetByExecutionID retrieves all processes for a specific execution
func (r *processRepository) GetByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*entity.Process, error) {
	var processes []entity.Process

	result := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Order("start_time DESC").Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get processes by execution: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// Update updates an existing process
func (r *processRepository) Update(ctx context.Context, process *entity.Process) error {
	// First check if process exists
	var existingProcess entity.Process
	result := r.db.WithContext(ctx).First(&existingProcess, "id = ?", process.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("process not found with id %s", process.ID)
		}
		return fmt.Errorf("failed to check process existence: %w", result.Error)
	}

	// Update the process
	result = r.db.WithContext(ctx).Save(process)
	if result.Error != nil {
		return fmt.Errorf("failed to update process: %w", result.Error)
	}

	return nil
}

// Delete deletes a process by ID (soft delete)
func (r *processRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Process{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete process: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// UpdateStatus updates the status of a process
func (r *processRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ProcessStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update process status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// UpdatePID updates the PID of a process
func (r *processRepository) UpdatePID(ctx context.Context, id uuid.UUID, pid int) error {
	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Update("pid", pid)
	if result.Error != nil {
		return fmt.Errorf("failed to update process PID: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// UpdateResourceUsage updates the resource usage of a process
func (r *processRepository) UpdateResourceUsage(ctx context.Context, id uuid.UUID, cpuUsage float64, memoryUsage uint64) error {
	updates := map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
	}

	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update process resource usage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// MarkCompleted marks a process as completed
func (r *processRepository) MarkCompleted(ctx context.Context, id uuid.UUID, endTime time.Time, exitCode *int) error {
	updates := map[string]interface{}{
		"status":   entity.ProcessStatusStopped,
		"end_time": endTime,
	}

	if exitCode != nil {
		updates["exit_code"] = *exitCode
	}

	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to mark process as completed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// MarkFailed marks a process as failed with error
func (r *processRepository) MarkFailed(ctx context.Context, id uuid.UUID, endTime time.Time, error string) error {
	updates := map[string]interface{}{
		"status":   entity.ProcessStatusError,
		"end_time": endTime,
		"error":    error,
	}

	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to mark process as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("process not found with id %s", id)
	}

	return nil
}

// GetByStatus retrieves processes by status
func (r *processRepository) GetByStatus(ctx context.Context, status entity.ProcessStatus) ([]*entity.Process, error) {
	var processes []entity.Process

	result := r.db.WithContext(ctx).Where("status = ?", status).Order("start_time DESC").Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get processes by status: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetByStatuses retrieves processes by multiple statuses
func (r *processRepository) GetByStatuses(ctx context.Context, statuses []entity.ProcessStatus) ([]*entity.Process, error) {
	var processes []entity.Process

	result := r.db.WithContext(ctx).Where("status IN ?", statuses).Order("start_time DESC").Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get processes by statuses: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetRunning retrieves all running processes
func (r *processRepository) GetRunning(ctx context.Context) ([]*entity.Process, error) {
	return r.GetByStatus(ctx, entity.ProcessStatusRunning)
}

// GetCompleted retrieves completed processes with limit
func (r *processRepository) GetCompleted(ctx context.Context, limit int) ([]*entity.Process, error) {
	var processes []entity.Process

	query := r.db.WithContext(ctx).Where("status IN ?", []entity.ProcessStatus{
		entity.ProcessStatusStopped,
		entity.ProcessStatusKilled,
		entity.ProcessStatusError,
	}).Order("end_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get completed processes: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetByPID retrieves a process by PID
func (r *processRepository) GetByPID(ctx context.Context, pid int) (*entity.Process, error) {
	var process entity.Process

	result := r.db.WithContext(ctx).First(&process, "pid = ?", pid)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("process not found with PID %d", pid)
		}
		return nil, fmt.Errorf("failed to get process by PID: %w", result.Error)
	}

	return &process, nil
}

// GetByDateRange retrieves processes within a date range
func (r *processRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Process, error) {
	var processes []entity.Process

	result := r.db.WithContext(ctx).Where("start_time BETWEEN ? AND ?", startDate, endDate).Order("start_time DESC").Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get processes by date range: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetProcessStats retrieves process statistics
func (r *processRepository) GetProcessStats(ctx context.Context, executionID *uuid.UUID) (*repository.ProcessStats, error) {
	var stats repository.ProcessStats

	query := r.db.WithContext(ctx).Model(&entity.Process{})
	if executionID != nil {
		query = query.Where("execution_id = ?", *executionID)
	}

	// Count total processes
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total processes: %w", err)
	}
	stats.TotalProcesses = totalCount

	// Count running processes
	var runningCount int64
	if err := query.Where("status = ?", entity.ProcessStatusRunning).Count(&runningCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count running processes: %w", err)
	}
	stats.RunningProcesses = runningCount

	// Count completed processes
	var completedCount int64
	if err := query.Where("status = ?", entity.ProcessStatusStopped).Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed processes: %w", err)
	}
	stats.CompletedProcesses = completedCount

	// Count failed processes
	var failedCount int64
	if err := query.Where("status = ?", entity.ProcessStatusError).Count(&failedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed processes: %w", err)
	}
	stats.FailedProcesses = failedCount

	// Calculate average CPU usage
	var avgCPU float64
	if err := query.Select("AVG(cpu_usage)").Scan(&avgCPU).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average CPU usage: %w", err)
	}
	stats.AverageCPUUsage = avgCPU

	// Calculate average memory usage
	var avgMemory uint64
	if err := query.Select("AVG(memory_usage)").Scan(&avgMemory).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average memory usage: %w", err)
	}
	stats.AverageMemoryUsage = avgMemory

	// Status distribution
	statusDistribution := make(map[entity.ProcessStatus]int64)
	var statusCounts []struct {
		Status entity.ProcessStatus
		Count  int64
	}

	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}

	for _, sc := range statusCounts {
		statusDistribution[sc.Status] = sc.Count
	}
	stats.StatusDistribution = statusDistribution

	// Recent activity (last 10 processes)
	var recentProcesses []entity.Process
	if err := query.Order("start_time DESC").Limit(10).Find(&recentProcesses).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent processes: %w", err)
	}

	recentPtrs := make([]*entity.Process, len(recentProcesses))
	for i := range recentProcesses {
		recentPtrs[i] = &recentProcesses[i]
	}
	stats.RecentActivity = recentPtrs

	return &stats, nil
}

// GetLongRunningProcesses retrieves processes running longer than threshold
func (r *processRepository) GetLongRunningProcesses(ctx context.Context, threshold time.Duration) ([]*entity.Process, error) {
	var processes []entity.Process

	thresholdTime := time.Now().Add(-threshold)
	result := r.db.WithContext(ctx).Where("status = ? AND start_time < ?", entity.ProcessStatusRunning, thresholdTime).Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get long running processes: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetHighResourceProcesses retrieves processes with high resource usage
func (r *processRepository) GetHighResourceProcesses(ctx context.Context, cpuThreshold float64, memoryThreshold uint64) ([]*entity.Process, error) {
	var processes []entity.Process

	result := r.db.WithContext(ctx).Where("cpu_usage > ? OR memory_usage > ?", cpuThreshold, memoryThreshold).Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get high resource processes: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetRecentProcesses retrieves recent processes with limit
func (r *processRepository) GetRecentProcesses(ctx context.Context, limit int) ([]*entity.Process, error) {
	var processes []entity.Process

	query := r.db.WithContext(ctx).Order("start_time DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent processes: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// GetActiveProcessesByExecution retrieves active processes for an execution
func (r *processRepository) GetActiveProcessesByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.Process, error) {
	var processes []entity.Process

	activeStatuses := []entity.ProcessStatus{
		entity.ProcessStatusStarting,
		entity.ProcessStatusRunning,
	}

	result := r.db.WithContext(ctx).Where("execution_id = ? AND status IN ?", executionID, activeStatuses).Find(&processes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active processes by execution: %w", result.Error)
	}

	// Convert to slice of pointers
	processPtrs := make([]*entity.Process, len(processes))
	for i := range processes {
		processPtrs[i] = &processes[i]
	}

	return processPtrs, nil
}

// CountActiveProcesses counts all active processes
func (r *processRepository) CountActiveProcesses(ctx context.Context) (int64, error) {
	var count int64

	activeStatuses := []entity.ProcessStatus{
		entity.ProcessStatusStarting,
		entity.ProcessStatusRunning,
	}

	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("status IN ?", activeStatuses).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count active processes: %w", result.Error)
	}

	return count, nil
}

// GetResourceUsageSummary retrieves resource usage summary
func (r *processRepository) GetResourceUsageSummary(ctx context.Context, executionID *uuid.UUID) (*repository.ResourceUsageSummary, error) {
	var summary repository.ResourceUsageSummary

	query := r.db.WithContext(ctx).Model(&entity.Process{})
	if executionID != nil {
		query = query.Where("execution_id = ?", *executionID)
	}

	// Get aggregated resource usage
	var result struct {
		TotalCPU    float64
		AverageCPU  float64
		PeakCPU     float64
		TotalMemory uint64
		AverageMemory uint64
		PeakMemory  uint64
		Count       int64
	}

	err := query.Select(
		"SUM(cpu_usage) as total_cpu",
		"AVG(cpu_usage) as average_cpu",
		"MAX(cpu_usage) as peak_cpu",
		"SUM(memory_usage) as total_memory",
		"AVG(memory_usage) as average_memory",
		"MAX(memory_usage) as peak_memory",
		"COUNT(*) as count",
	).Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage summary: %w", err)
	}

	summary.TotalCPUUsage = result.TotalCPU
	summary.AverageCPUUsage = result.AverageCPU
	summary.PeakCPUUsage = result.PeakCPU
	summary.TotalMemoryUsage = result.TotalMemory
	summary.AverageMemoryUsage = result.AverageMemory
	summary.PeakMemoryUsage = result.PeakMemory
	summary.ProcessCount = result.Count

	return &summary, nil
}

// BulkUpdateStatus updates status for multiple processes
func (r *processRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.ProcessStatus) error {
	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id IN ?", ids).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk update process status: %w", result.Error)
	}

	return nil
}

// BulkDelete deletes multiple processes
func (r *processRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Process{}, "id IN ?", ids)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete processes: %w", result.Error)
	}

	return nil
}

// CleanupOldProcesses removes old processes
func (r *processRepository) CleanupOldProcesses(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity.Process{}, "start_time < ?", olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old processes: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// TerminateProcessesByExecution terminates all processes for an execution
func (r *processRepository) TerminateProcessesByExecution(ctx context.Context, executionID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.Process{}).
		Where("execution_id = ? AND status IN ?", executionID, []entity.ProcessStatus{
			entity.ProcessStatusStarting,
			entity.ProcessStatusRunning,
		}).
		Update("status", entity.ProcessStatusKilled)

	if result.Error != nil {
		return fmt.Errorf("failed to terminate processes by execution: %w", result.Error)
	}

	return nil
}

// ValidateProcessExists checks if a process exists
func (r *processRepository) ValidateProcessExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.Process{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate process existence: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateExecutionExists checks if an execution exists
func (r *processRepository) ValidateExecutionExists(ctx context.Context, executionID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", executionID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate execution existence: %w", result.Error)
	}

	return count > 0, nil
}