package database

import (
	"github.com/auto-devs/auto-devs/internal/entity"
)

// RunMigrations runs all database migrations using GORM AutoMigrate
func RunMigrations(db *GormDB) error {
	// AutoMigrate will create tables, foreign keys, constraints, and indexes
	// based on the struct tags and relationships defined in the entities
	return db.AutoMigrate(
		&entity.Project{},
		&entity.Task{},
	)
}
