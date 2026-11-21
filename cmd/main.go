package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"music-app/internal/config"
	"music-app/internal/database"
	"music-app/internal/handler"
	"music-app/internal/middleware"
	"music-app/internal/model"
	"music-app/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	//  1. Загрузка .env файла 
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found, using environment variables")
	}

	// 2. Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 3. Подключение к базе данных
	db, err := database.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// 4. Авто-миграция таблиц
	err = db.AutoMigrate(&model.User{}, &model.Track{}, &model.VerificationCode{})
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("✅ Database tables created successfully")

	// 5.  Оптимизация для продакшена 
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 6. Инициализация веб-сервера
	r := gin.Default()

	// 7. Создаём обработчики
	authHandler := &handler.AuthHandler{DB: db}
	trackHandler := &handler.TrackHandler{DB: db}

	// 8. Сервис верификации email 
	emailService := &service.EmailService{
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUser:     cfg.SMTPUser,
		SMTPPassword: cfg.SMTPPassword,
		FromEmail:    cfg.FromEmail,
	}

	verificationHandler := &handler.VerificationHandler{
		DB:           db,
		EmailService: emailService,
	}

	//  9. Раздача статических файлов 
	r.Static("/css", "frontend/css")
	r.Static("/js", "frontend/js")
	r.Static("/uploads", "uploads")
	r.StaticFile("/", "frontend/index.html")
	r.StaticFile("/levels", "frontend/levels/index.html")

	// 10. Публичные маршруты API
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	// 11. Маршруты авторизации и верификации
	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/verify/send", verificationHandler.SendVerificationCode)
		api.POST("/verify/confirm", verificationHandler.VerifyEmail)
	}

	// 12. Защищённые маршруты
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/tracks/upload", trackHandler.UploadTrack)
		protected.GET("/tracks", trackHandler.GetTracks)
		protected.DELETE("/tracks/:id", trackHandler.DeleteTrack)
	}

	// 13. Настройка HTTP сервера для многопоточности 
	srv := &http.Server{
		Addr:         "0.0.0.0:" + cfg.ServerPort, //  Слушаем все сетевые интерфейсы
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	//  14. Запуск сервера в отдельной горутине 
	go func() {
		log.Printf("🚀 Server starting on %s", srv.Addr)
		log.Printf("🌐 Frontend available at http://localhost:%s", cfg.ServerPort)

		if cfg.SMTPHost != "" && cfg.SMTPPort > 0 {
			log.Printf("✉️  Email verification enabled (SMTP: %s:%d)", cfg.SMTPHost, cfg.SMTPPort)
		} else {
			log.Println("⚠️  SMTP not configured — email verification will not work")
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	//  15. Graceful shutdown (корректное завершение работы) 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
