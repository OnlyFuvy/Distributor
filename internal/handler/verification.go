package handler

import (
	"music-app/internal/model"
	"music-app/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VerificationHandler — обработчик верификации email
type VerificationHandler struct {
	DB           *gorm.DB
	EmailService *service.EmailService
}

// SendCodeRequest — запрос на отправку кода
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendCodeResponse — ответ на отправку кода
type SendCodeResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// VerifyCodeRequest — запрос на подтверждение кода
type VerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// VerifyCodeResponse — ответ на подтверждение кода
type VerifyCodeResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SendVerificationCode — отправляет код подтверждения на email
func (h *VerificationHandler) SendVerificationCode(c *gin.Context) {
	var req SendCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendCodeResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// 1. Проверяем, существует ли пользователь с таким email
	var user model.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, SendCodeResponse{
				Error: "User with this email not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, SendCodeResponse{
			Error: "Database error",
		})
		return
	}

	// 2. Проверяем, не верифицирован ли уже пользователь
	if user.IsVerified {
		c.JSON(http.StatusBadRequest, SendCodeResponse{
			Error: "Email already verified",
		})
		return
	}

	// 3. Генерируем код
	code := service.GenerateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	// 4. Сохраняем код в БД (или обновляем существующий)
	var verificationCode model.VerificationCode
	result := h.DB.Where("email = ? AND used = ?", req.Email, false).
		Order("created_at DESC").
		First(&verificationCode)

	if result.Error == nil && !verificationCode.IsExpired() {
		// Обновляем существующий код
		verificationCode.Code = code
		verificationCode.ExpiresAt = expiresAt
		h.DB.Save(&verificationCode)
	} else {
		// Создаём новый код
		verificationCode = model.VerificationCode{
			UserID:    user.ID,
			Email:     req.Email,
			Code:      code,
			ExpiresAt: expiresAt,
			Used:      false,
		}
		h.DB.Create(&verificationCode)
	}

	// 5. Отправляем email
	if err := h.EmailService.SendVerificationEmail(req.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, SendCodeResponse{
			Error: "Failed to send email: " + err.Error(),
		})
		return
	}

	// 6. Успешный ответ
	c.JSON(http.StatusOK, SendCodeResponse{
		Message: "Verification code sent to " + req.Email,
	})
}

// VerifyEmail — подтверждает email по коду
func (h *VerificationHandler) VerifyEmail(c *gin.Context) {
	var req VerifyCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, VerifyCodeResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// 1. Находим код в БД
	var verificationCode model.VerificationCode
	if err := h.DB.Where("email = ? AND code = ? AND used = ?", req.Email, req.Code, false).
		First(&verificationCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, VerifyCodeResponse{
				Error: "Invalid or expired code",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, VerifyCodeResponse{
			Error: "Database error",
		})
		return
	}

	// 2. Проверяем, не истёк ли срок действия
	if verificationCode.IsExpired() {
		c.JSON(http.StatusBadRequest, VerifyCodeResponse{
			Error: "Code has expired",
		})
		return
	}

	// 3. Находим пользователя
	var user model.User
	if err := h.DB.First(&user, verificationCode.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, VerifyCodeResponse{
			Error: "User not found",
		})
		return
	}

	// 4. Помечаем пользователя как верифицированного
	user.IsVerified = true
	h.DB.Save(&user)

	// 5. Помечаем код как использованный
	verificationCode.Used = true
	h.DB.Save(&verificationCode)

	// 6. Генерируем JWT токен для автоматического входа
	token, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, VerifyCodeResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// 7. Успешный ответ с токеном
	c.JSON(http.StatusOK, VerifyCodeResponse{
		Message: "Email verified successfully",
		Token:   token,
	})
}
