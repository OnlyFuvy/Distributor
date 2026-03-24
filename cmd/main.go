package main

import (
	"log"
	"music-app/internal/config"
	"music-app/internal/database"
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Загрузка конфигурации из .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Подключение к базе данных
	db, err := database.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// 3. Инициализация веб-сервера (Gin)
	r := gin.Default()

	// 4. Простой маршрут для проверки работы
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok", 
			"db": "connected",
		})
	})

	// 5. Запуск сервера на указанном порту
	log.Printf("🚀 Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Подавляем предупреждение о неиспользуемой переменной (пока)
	_ = db
}