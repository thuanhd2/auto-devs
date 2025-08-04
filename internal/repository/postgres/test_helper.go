package postgres

import (
	"context"
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
)

// setupTestGormDB creates a test database using GORM and testcontainers
func setupTestGormDB(t *testing.T) (*database.GormDB, func()) {
	ctx := context.Background()

	// Create PostgreSQL test container
	pgContainer, err := pgcontainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		pgcontainer.WithDatabase("testdb"),
		pgcontainer.WithUsername("testuser"),
		pgcontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection details
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create GORM connection
	dsn := "host=" + host + " user=testuser password=testpass dbname=testdb port=" + port.Port() + " sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&entity.Project{}, &entity.Task{}, &entity.TaskStatusHistory{})
	require.NoError(t, err)

	// Create GormDB wrapper
	gormDB := &database.GormDB{DB: db}

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		pgContainer.Terminate(ctx)
	}

	return gormDB, cleanup
}
