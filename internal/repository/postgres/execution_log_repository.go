package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type executionLogRepository struct {
	db *database.GormDB
}

// NewExecutionLogRepository creates a new PostgreSQL execution log repository
func NewExecutionLogRepository(db *database.GormDB) repository.ExecutionLogRepository {
	return &executionLogRepository{
		db: db,
	}
}

// Create creates a new execution log
func (r *executionLogRepository) Create(ctx context.Context, log *entity.ExecutionLog) error {
	// Generate UUID if not provided
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create execution log: %w", result.Error)
	}

	return nil
}

// GetByID retrieves an execution log by ID
func (r *executionLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ExecutionLog, error) {
	var log entity.ExecutionLog

	result := r.db.WithContext(ctx).First(&log, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution log not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get execution log: %w", result.Error)
	}

	return &log, nil
}

// GetByExecutionID retrieves all logs for a specific execution
func (r *executionLogRepository) GetByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get execution logs: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetByProcessID retrieves all logs for a specific process
func (r *executionLogRepository) GetByProcessID(ctx context.Context, processID uuid.UUID) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("process_id = ?", processID).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get process logs: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// Delete deletes an execution log by ID (soft delete)
func (r *executionLogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.ExecutionLog{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete execution log: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("execution log not found with id %s", id)
	}

	return nil
}

// BatchCreate creates multiple execution logs in a single transaction
func (r *executionLogRepository) BatchCreate(ctx context.Context, logs []*entity.ExecutionLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Set default values for logs
	for _, log := range logs {
		if log.ID == uuid.Nil {
			log.ID = uuid.New()
		}
		if log.Timestamp.IsZero() {
			log.Timestamp = time.Now()
		}
	}

	// Process in batches to avoid memory issues
	batchSize := 100 // Default batch size
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}

		batch := logs[i:end]
		result := r.db.WithContext(ctx).CreateInBatches(batch, batchSize)
		if result.Error != nil {
			return fmt.Errorf("failed to batch create execution logs: %w", result.Error)
		}
	}

	return nil
}

// BatchInsertOrUpdate inserts or updates logs
func (r *executionLogRepository) BatchInsertOrUpdate(ctx context.Context, logs []*entity.ExecutionLog) error {
	if len(logs) == 0 {
		return nil
	}

	for _, log := range logs {
		if err := r.insertOrUpdateLog(ctx, log); err != nil {
			return fmt.Errorf("failed to insert/update log: %w", err)
		}
	}

	return nil
}

// insertOrUpdateLog handles a single log insert or update
func (r *executionLogRepository) insertOrUpdateLog(ctx context.Context, log *entity.ExecutionLog) error {
	// Check if log exists based on execution_id and line
	var existingLog entity.ExecutionLog
	result := r.db.WithContext(ctx).Where("execution_id = ? AND line = ?", log.ExecutionID, log.Line).First(&existingLog)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Log doesn't exist, create new one
			if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
				return fmt.Errorf("failed to create execution log: %w", err)
			}
		} else {
			// Database error
			return fmt.Errorf("failed to check existing log: %w", result.Error)
		}
	} else {
		// Log exists, update it
		// Preserve the original ID and created_at
		updateData := map[string]interface{}{
			"message":   log.Message,
			"log_level": log.Level,
			"source":    log.Source,
			"metadata":  log.Metadata,
			"timestamp": log.Timestamp,
		}

		if err := r.db.WithContext(ctx).Model(&existingLog).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update execution log: %w", err)
		}
	}

	return nil
}

// GetLogsBatch retrieves logs in batches for pagination
func (r *executionLogRepository) GetLogsBatch(ctx context.Context, executionID uuid.UUID, limit, offset int) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	query := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get logs batch: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetByLevel retrieves logs by level
func (r *executionLogRepository) GetByLevel(ctx context.Context, executionID uuid.UUID, level entity.LogLevel) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("execution_id = ? AND level = ?", executionID, level).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get logs by level: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetByLevels retrieves logs by multiple levels
func (r *executionLogRepository) GetByLevels(ctx context.Context, executionID uuid.UUID, levels []entity.LogLevel) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("execution_id = ? AND level IN ?", executionID, levels).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get logs by levels: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetBySource retrieves logs by source
func (r *executionLogRepository) GetBySource(ctx context.Context, executionID uuid.UUID, source string) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("execution_id = ? AND source = ?", executionID, source).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get logs by source: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetByDateRange retrieves logs within a date range
func (r *executionLogRepository) GetByDateRange(ctx context.Context, executionID uuid.UUID, startDate, endDate time.Time) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	result := r.db.WithContext(ctx).Where("execution_id = ? AND timestamp BETWEEN ? AND ?", executionID, startDate, endDate).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get logs by date range: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetRecentLogs retrieves recent logs with limit
func (r *executionLogRepository) GetRecentLogs(ctx context.Context, executionID uuid.UUID, limit int) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	query := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// SearchLogs searches logs by message content
func (r *executionLogRepository) SearchLogs(ctx context.Context, executionID uuid.UUID, searchTerm string) ([]*entity.ExecutionLog, error) {
	var logs []entity.ExecutionLog

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	result := r.db.WithContext(ctx).Where("execution_id = ? AND LOWER(message) LIKE ?", executionID, searchPattern).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search logs: %w", result.Error)
	}

	// Convert to slice of pointers
	logPtrs := make([]*entity.ExecutionLog, len(logs))
	for i := range logs {
		logPtrs[i] = &logs[i]
	}

	return logPtrs, nil
}

// GetLogStats retrieves log statistics
func (r *executionLogRepository) GetLogStats(ctx context.Context, executionID uuid.UUID) (*repository.LogStats, error) {
	var stats repository.LogStats

	query := r.db.WithContext(ctx).Model(&entity.ExecutionLog{}).Where("execution_id = ?", executionID)

	// Count total logs
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total logs: %w", err)
	}
	stats.TotalLogs = totalCount

	// Count logs by level
	logsByLevel := make(map[entity.LogLevel]int64)
	var levelCounts []struct {
		Level entity.LogLevel
		Count int64
	}

	if err := query.Select("level, COUNT(*) as count").Group("level").Scan(&levelCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get logs by level: %w", err)
	}

	for _, lc := range levelCounts {
		logsByLevel[lc.Level] = lc.Count
	}
	stats.LogsByLevel = logsByLevel

	// Count logs by source
	logsBySource := make(map[string]int64)
	var sourceCounts []struct {
		Source string
		Count  int64
	}

	if err := query.Select("source, COUNT(*) as count").Group("source").Scan(&sourceCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get logs by source: %w", err)
	}

	for _, sc := range sourceCounts {
		logsBySource[sc.Source] = sc.Count
	}
	stats.LogsBySource = logsBySource

	// Count error and warning logs
	stats.ErrorCount = logsByLevel[entity.LogLevelError]
	stats.WarningCount = logsByLevel[entity.LogLevelWarn]

	// Get first and last log times
	var firstTime, lastTime time.Time
	if err := query.Select("MIN(timestamp)").Scan(&firstTime).Error; err != nil {
		return nil, fmt.Errorf("failed to get first log time: %w", err)
	}
	if !firstTime.IsZero() {
		stats.FirstLogTime = &firstTime
	}

	if err := query.Select("MAX(timestamp)").Scan(&lastTime).Error; err != nil {
		return nil, fmt.Errorf("failed to get last log time: %w", err)
	}
	if !lastTime.IsZero() {
		stats.LastLogTime = &lastTime
	}

	// Get recent error logs
	var recentErrorLogs []entity.ExecutionLog
	if err := query.Where("level = ?", entity.LogLevelError).Order("timestamp DESC").Limit(5).Find(&recentErrorLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent error logs: %w", err)
	}

	recentErrorPtrs := make([]*entity.ExecutionLog, len(recentErrorLogs))
	for i := range recentErrorLogs {
		recentErrorPtrs[i] = &recentErrorLogs[i]
	}
	stats.RecentErrorLogs = recentErrorPtrs

	// Calculate log size (approximate)
	var totalSize int64
	if err := query.Select("SUM(LENGTH(message))").Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate log size: %w", err)
	}
	stats.LogSizeBytes = totalSize

	return &stats, nil
}

// GetErrorLogs retrieves error logs with limit
func (r *executionLogRepository) GetErrorLogs(ctx context.Context, executionID uuid.UUID, limit int) ([]*entity.ExecutionLog, error) {
	return r.GetByLevel(ctx, executionID, entity.LogLevelError)
}

// GetLogsByTimeWindow retrieves logs within a time window
func (r *executionLogRepository) GetLogsByTimeWindow(ctx context.Context, executionID uuid.UUID, windowStart, windowEnd time.Time) ([]*entity.ExecutionLog, error) {
	return r.GetByDateRange(ctx, executionID, windowStart, windowEnd)
}

// RotateLogs keeps only the most recent logs up to maxLogs
func (r *executionLogRepository) RotateLogs(ctx context.Context, executionID uuid.UUID, maxLogs int) error {
	if maxLogs <= 0 {
		return fmt.Errorf("maxLogs must be positive")
	}

	// Get total log count
	var totalCount int64
	result := r.db.WithContext(ctx).Model(&entity.ExecutionLog{}).Where("execution_id = ?", executionID).Count(&totalCount)
	if result.Error != nil {
		return fmt.Errorf("failed to count logs: %w", result.Error)
	}

	if totalCount <= int64(maxLogs) {
		return nil // No rotation needed
	}

	// Delete old logs, keeping only the most recent ones
	subquery := r.db.Model(&entity.ExecutionLog{}).
		Select("id").
		Where("execution_id = ?", executionID).
		Order("timestamp DESC").
		Limit(maxLogs)

	result = r.db.WithContext(ctx).Where("execution_id = ? AND id NOT IN (?)", executionID, subquery).Delete(&entity.ExecutionLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to rotate logs: %w", result.Error)
	}

	return nil
}

// CleanupOldLogs removes logs older than the specified time
func (r *executionLogRepository) CleanupOldLogs(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity.ExecutionLog{}, "timestamp < ?", olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// CleanupExecutionLogs cleans up logs for a specific execution, keeping only recent ones
func (r *executionLogRepository) CleanupExecutionLogs(ctx context.Context, executionID uuid.UUID, keepRecent int) (int64, error) {
	if keepRecent <= 0 {
		// Delete all logs for the execution
		result := r.db.WithContext(ctx).Unscoped().Delete(&entity.ExecutionLog{}, "execution_id = ?", executionID)
		if result.Error != nil {
			return 0, fmt.Errorf("failed to cleanup execution logs: %w", result.Error)
		}
		return result.RowsAffected, nil
	}

	// Keep only the most recent logs
	subquery := r.db.Model(&entity.ExecutionLog{}).
		Select("id").
		Where("execution_id = ?", executionID).
		Order("timestamp DESC").
		Limit(keepRecent)

	result := r.db.WithContext(ctx).Unscoped().Where("execution_id = ? AND id NOT IN (?)", executionID, subquery).Delete(&entity.ExecutionLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup execution logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// ArchiveLogs moves old logs to an archive table (implementation would depend on specific requirements)
func (r *executionLogRepository) ArchiveLogs(ctx context.Context, executionID uuid.UUID, olderThan time.Time) (int64, error) {
	// For now, this is a placeholder that does soft delete
	// In a real implementation, you might move logs to an archive table
	result := r.db.WithContext(ctx).Where("execution_id = ? AND timestamp < ?", executionID, olderThan).Delete(&entity.ExecutionLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to archive logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// BulkDelete deletes multiple logs
func (r *executionLogRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.ExecutionLog{}, "id IN ?", ids)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete logs: %w", result.Error)
	}

	return nil
}

// BulkDeleteByExecution deletes all logs for an execution
func (r *executionLogRepository) BulkDeleteByExecution(ctx context.Context, executionID uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&entity.ExecutionLog{}, "execution_id = ?", executionID)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to bulk delete logs by execution: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// BulkDeleteByLevel deletes logs by level older than specified time
func (r *executionLogRepository) BulkDeleteByLevel(ctx context.Context, level entity.LogLevel, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&entity.ExecutionLog{}, "level = ? AND timestamp < ?", level, olderThan)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to bulk delete logs by level: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// ValidateLogExists checks if a log exists
func (r *executionLogRepository) ValidateLogExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.ExecutionLog{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate log existence: %w", result.Error)
	}

	return count > 0, nil
}

// ValidateExecutionExists checks if an execution exists
func (r *executionLogRepository) ValidateExecutionExists(ctx context.Context, executionID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entity.Execution{}).Where("id = ?", executionID).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to validate execution existence: %w", result.Error)
	}

	return count > 0, nil
}
