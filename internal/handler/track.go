package handler

import (
	"music-app/internal/model"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrackHandler struct {
	DB *gorm.DB
}

// UploadTrackResponse — ответ сервера при загрузке трека
type UploadTrackResponse struct {
	Message string       `json:"message"`
	Track   *model.Track `json:"track,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// UploadTrack — обработчик загрузки музыкального трека
func (h *TrackHandler) UploadTrack(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста (добавим позже в middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, UploadTrackResponse{
			Error: "User not authenticated",
		})
		return
	}

	// 2. Получаем файл из запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadTrackResponse{
			Error: "No file uploaded: " + err.Error(),
		})
		return
	}

	// 3. Проверяем тип файла (только аудио)
	allowedTypes := map[string]bool{
		"audio/mpeg": true, // .mp3
		"audio/wav":  true, // .wav
		"audio/flac": true, // .flac
		"audio/mp4":  true, // .m4a
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, UploadTrackResponse{
			Error: "Invalid file type. Only audio files allowed",
		})
		return
	}

	// 4. Создаём уникальное имя файла
	fileExt := filepath.Ext(file.Filename)
	fileName := strconv.FormatUint(uint64(userID.(uint)), 10) + "_" +
		strconv.FormatInt(file.Size, 10) + fileExt
	filePath := filepath.Join("uploads", fileName)

	// 5. Сохраняем файл на диск
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, UploadTrackResponse{
			Error: "Failed to save file: " + err.Error(),
		})
		return
	}

	// 6. Создаём запись в базе данных
	track := model.Track{
		UserID:   userID.(uint),
		Title:    file.Filename, // Пока берём имя файла как название
		FileName: file.Filename,
		FilePath: filePath,
		FileSize: file.Size,
	}

	if err := h.DB.Create(&track).Error; err != nil {
		// Если ошибка БД — удаляем файл с диска
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, UploadTrackResponse{
			Error: "Failed to save track to database",
		})
		return
	}

	// 7. Успешный ответ
	c.JSON(http.StatusCreated, UploadTrackResponse{
		Message: "Track uploaded successfully",
		Track:   &track,
	})
}
