package handler

import (
	"music-app/internal/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// JWTSecret — секретный ключ для подписи токенов (в продакшене хранить в .env!)
var JWTSecret = []byte("your-secret-key-change-in-production")

// AuthHandler — обработчик авторизации
type AuthHandler struct {
	DB *gorm.DB
}

// Claims — структура данных внутри JWT токена
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// RegisterRequest — данные для регистрации
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse — ответ сервера при регистрации
type RegisterResponse struct {
	Message string      `json:"message"`
	User    interface{} `json:"user,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginRequest — данные для входа
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse — ответ сервера при входе
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Register — обработчик регистрации нового пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// 1. Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// 2. Проверка, существует ли пользователь с таким email или username
	var existingUser model.User
	if err := h.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Error: "Database error",
		})
		return
	}

	if existingUser.ID != 0 {
		c.JSON(http.StatusConflict, RegisterResponse{
			Error: "User with this email or username already exists",
		})
		return
	}

	// 3. Хэширование пароля (никогда не храним пароли в открытом виде!)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// 4. Создание нового пользователя
	user := model.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Error: "Failed to create user",
		})
		return
	}

	// 5. Успешный ответ (без пароля!)
	c.JSON(http.StatusCreated, RegisterResponse{
		Message: "User registered successfully",
		User: gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// Login — обработчик входа пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// 1. Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// 2. Поиск пользователя по email
	var user model.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, LoginResponse{
				Error: "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Error: "Database error",
		})
		return
	}

	// 3. Проверка пароля (сравниваем введённый с хэшем в БД)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Error: "Invalid email or password",
		})
		return
	}

	// 4. Генерация JWT токена
	token, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// 5. Успешный ответ с токеном
	c.JSON(http.StatusOK, LoginResponse{
		Message: "Login successful",
		Token:   token,
	})
}

// GenerateJWT — генерирует JWT токен для пользователя
func GenerateJWT(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "music-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}
