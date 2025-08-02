//go:build wireinject
// +build wireinject

package di

import (
	"github.com/auto-devs/auto-devs/config"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the entire application
var ProviderSet = wire.NewSet(
	config.Load,
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
	// Add other dependencies here as needed
}

// NewApp creates a new App instance
func NewApp(cfg *config.Config) *App {
	return &App{
		Config: cfg,
	}
}