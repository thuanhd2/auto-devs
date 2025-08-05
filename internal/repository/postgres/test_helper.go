package postgres

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/repository"
	"github.com/auto-devs/auto-devs/pkg/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // postgres driver
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *database.GormDB

func newDbMigrator() pgtestdb.Migrator {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../../")

	folderPath := filepath.Join(projectRoot, "migrations")
	gm := golangmigrator.New(folderPath)

	return gm
}

func SetupTestDB(t *testing.T) *database.GormDB {
	// Get the absolute path to the project root directory
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../../")

	// Load .env.test from project root
	godotenv.Load(filepath.Join(projectRoot, ".env.test"))
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	role := pgtestdb.DefaultRole()
	role.Capabilities = "SUPERUSER"
	// database.InitDB()
	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       dbHost,
		Port:       dbPort,
		User:       dbUser,
		Password:   dbPassword,
		Database:   dbName,
		Options:    "sslmode=disable",
		TestRole:   &role,
	}
	migrator := newDbMigrator()
	sqlDb := pgtestdb.New(t, pgxConf, migrator)
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDb,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	testDB = &database.GormDB{
		DB: db,
	}
	return testDB
}

func TeardownTestDB() error {
	return nil
}

// Helper functions for creating test data
func CreateTestProject(t *testing.T, projectRepo repository.ProjectRepository, ctx context.Context) *entity.Project {
	project := &entity.Project{
		Name:          "Test Project",
		Description:   "Test Description",
		RepositoryURL: "https://github.com/test/repo.git",
	}
	err := projectRepo.Create(ctx, project)
	require.NoError(t, err)
	return project
}

func CreateTestTask(t *testing.T, taskRepo repository.TaskRepository, projectID uuid.UUID, ctx context.Context) *entity.Task {
	task := &entity.Task{
		ProjectID:   projectID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.TaskStatusTODO,
	}

	err := taskRepo.Create(ctx, task)
	require.NoError(t, err)
	return task
}
