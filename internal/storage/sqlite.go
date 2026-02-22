package storage

import (
	"log"
	"path/filepath"

	"ai-companion-cli-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents the database connection
type DB struct {
	*gorm.DB
	dbPath string
}

// NewDB initialize database connection given path
func NewDB(dbPath string) *DB {
	newLogger := logger.Default.LogMode(logger.Silent)

	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &DB{
		DB:     gormDB,
		dbPath: dbPath,
	}
}

// Initialize creates tables if they don't exist
func (db *DB) Initialize() error {
	return db.AutoMigrate(
		&models.CharacterProfile{},
		&models.ChatMessage{},
		&models.SessionState{},
		&models.RelationshipState{},
		&models.MemoryFact{},
		&models.MemorySummary{},
		&models.CharacterEmotionState{},
	)
}

// GetDBPath returns the underlying file path
func (db *DB) GetDBPath() string {
	return filepath.Clean(db.dbPath)
}
