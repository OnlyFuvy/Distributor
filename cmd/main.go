package main

import (
	"log"
	"music-app/internal/config"
	"music-app/internal/database"
	"music-app/internal/handler"
	"music-app/internal/middleware"
	"music-app/internal/model"
	"net/http"
	"path/filepath"

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

	// 3. Авто-миграция: Создаём таблицы в БД автоматически
	err = db.AutoMigrate(&model.User{}, &model.Track{})
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("✅ Database tables created successfully")

	// 4. Инициализация веб-сервера (Gin)
	r := gin.Default()

	// 5. Создаём обработчики
	authHandler := &handler.AuthHandler{DB: db}
	trackHandler := &handler.TrackHandler{DB: db}

	// 🔥 РАЗДАЧА СТАТИЧЕСКИХ ФАЙЛОВ (фронтенд + uploads) 🔥
	r.Static("/css", filepath.Join("frontend", "css"))
	r.Static("/js", filepath.Join("frontend", "js"))
	r.Static("/uploads", filepath.Join("uploads")) // ← Добавьте эту строку!
	r.StaticFile("/", filepath.Join("frontend", "index.html"))

	// 7. Публичные маршруты API
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	// 8. Маршруты авторизации (доступны всем)
	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register) // Регистрация
		api.POST("/login", authHandler.Login)       // Вход
	}

	// 9. Защищённые маршруты (требуется JWT токен)
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/tracks/upload", trackHandler.UploadTrack) // Загрузка трека
		protected.GET("/tracks", trackHandler.GetTracks)           // Получить свои треки
	}

	// 10. Запуск сервера на указанном порту
	log.Printf("🚀 Server starting on port %s", cfg.ServerPort)
	log.Printf("🌐 Frontend available at http://localhost:%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
