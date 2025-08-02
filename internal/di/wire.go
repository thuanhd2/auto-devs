//go:build wireinject
// +build wireinject

package di

import (
	"github.com/auto-devs/auto-devs/config"
	"github.com/auto-devs/auto-devs/pkg/database"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the entire application
var ProviderSet = wire.NewSet(
	config.Load,
	ProvideDatabase,
	// Add repository providers here when database is implemented
	// Add usecase providers here
	// Add handler providers here
)

// InitializeApp builds the entire dependency tree
func InitializeApp() (*App, error) {
	wire.Build(
		ProviderSet,
		NewApp,
	)
	return &App{}, nil
}

// App represents the initialized application with all dependencies
type App struct {
	Config *config.Config
	DB     *database.DB
	// Add other dependencies here as needed
}

// NewApp creates a new App instance
func NewApp(cfg *config.Config, db *database.DB) *App {
	return &App{
		Config: cfg,
		DB:     db,
	}
}

// ProvideDatabase provides a database connection
func ProvideDatabase(cfg *config.Config) (*database.DB, error) {
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	return database.NewConnection(dbConfig)
}