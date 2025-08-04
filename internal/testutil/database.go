package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDBConfig holds configuration for test database setup
type TestDBConfig struct {
	// Database name
	Database string
	// Username for database connection
	Username string
	// Password for database connection
	Password string
	// Enable GORM debug logging
	EnableLogging bool
	// Custom migration entities
	MigrationEntities []interface{}
}

// DefaultTestDBConfig returns default configuration for test database
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Database:          "testdb",
		Username:          "testuser",
		Password:          "testpass",
		EnableLogging:     false,
		MigrationEntities: getDefaultEntities(),
	}
}

// TestContainer wraps testcontainers functionality for database testing
type TestContainer struct {
	Container testcontainers.Container
	DB        *database.GormDB
	GormDB    *gorm.DB
	Config    *TestDBConfig
}

// SetupTestDB creates a test database using GORM and testcontainers
func SetupTestDB(t *testing.T, config ...*TestDBConfig) (*TestContainer, func()) {
	cfg := DefaultTestDBConfig()
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	ctx := context.Background()

	// Create PostgreSQL test container
	pgContainer, err := pgcontainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		pgcontainer.WithDatabase(cfg.Database),
		pgcontainer.WithUsername(cfg.Username),
		pgcontainer.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection details
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Build DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, cfg.Username, cfg.Password, cfg.Database, port.Port())

	// Configure GORM
	gormConfig := &gorm.Config{}
	if cfg.EnableLogging {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	// Create GORM connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(cfg.MigrationEntities...)
	require.NoError(t, err)

	// Create GormDB wrapper
	gormDB := &database.GormDB{DB: db}

	testContainer := &TestContainer{
		Container: pgContainer,
		DB:        gormDB,
		GormDB:    db,
		Config:    cfg,
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return testContainer, cleanup
}

// getDefaultEntities returns the default entities for migration
func getDefaultEntities() []interface{} {
	return []interface{}{
		&entity.Project{},
		&entity.Task{},
		&entity.AuditLog{},
	}
}

// TruncateTables truncates all tables in the test database
func (tc *TestContainer) TruncateTables(t *testing.T) {
	tables := []string{"audit_logs", "tasks", "projects"}
	
	for _, table := range tables {
		err := tc.GormDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error
		require.NoError(t, err)
	}
}

// GetConnectionString returns the database connection string for the test container
func (tc *TestContainer) GetConnectionString(t *testing.T) string {
	ctx := context.Background()
	
	host, err := tc.Container.Host(ctx)
	require.NoError(t, err)

	port, err := tc.Container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, tc.Config.Username, tc.Config.Password, tc.Config.Database, port.Port())
}

// ExecuteInTransaction executes a function within a database transaction that gets rolled back
func (tc *TestContainer) ExecuteInTransaction(t *testing.T, fn func(*gorm.DB)) {
	tx := tc.GormDB.Begin()
	require.NoError(t, tx.Error)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		tx.Rollback()
	}()

	fn(tx)
}

// WithTestTransaction is a helper that creates a test database and executes the test in a transaction
func WithTestTransaction(t *testing.T, testFn func(*testing.T, *gorm.DB)) {
	container, cleanup := SetupTestDB(t)
	defer cleanup()

	container.ExecuteInTransaction(t, func(tx *gorm.DB) {
		testFn(t, tx)
	})
}