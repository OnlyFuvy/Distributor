package config

import (
	"os"
	"strconv"
)

// Config содержит все настройки приложения
type Config struct {
	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Server
	ServerPort string

	// JWT
	JWTSecret string

	// 🔥 SMTP настройки для отправки email 🔥
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
}

// LoadConfig загружает конфигурацию из переменных окружения (.env)
func LoadConfig() (*Config, error) {
	// Если есть файл .env, godotenv уже загрузил его в main()
	// Здесь просто читаем переменные окружения

	// Database
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	// SMTP Port
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "music_app"),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		// 🔥 SMTP
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
	}, nil
}

// getEnv — вспомогательная функция для получения переной окружения с дефолтом
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
