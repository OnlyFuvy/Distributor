package database

import (
	"fmt"
	"music-app/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabaseConnection создаёт подключение к PostgreSQL
func NewDatabaseConnection(cfg *config.Config) (*gorm.DB, error) {
	// ✅ Правильное форматирование: %d для целых чисел (порт)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort, // ← %d ожидает int, а не %s
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
