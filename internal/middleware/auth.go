package middleware

import (
	"music-app/internal/handler"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth — middleware для проверки JWT токена
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 2. Проверяем формат: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. Парсим и валидируем токен
		claims := &handler.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return handler.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 4. Сохраняем данные пользователя в контекст запроса
		//    Теперь любой обработчик может получить userID через c.Get("userID")
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		// 5. Передаём управление следующему обработчику
		c.Next()
	}
}
