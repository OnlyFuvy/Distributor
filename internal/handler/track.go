package handler

import (
	"encoding/json"
	"music-app/internal/model"
	"music-app/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TrackHandler — обработчик для работы с треками
type TrackHandler struct {
	DB *gorm.DB
}

// UploadTrackResponse — ответ сервера при загрузке трека
type UploadTrackResponse struct {
	Message string       `json:"message"`
	Track   *model.Track `json:"track,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GetTracksResponse — ответ сервера со списком треков
type GetTracksResponse struct {
	Message string        `json:"message"`
	Tracks  []model.Track `json:"tracks,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// UploadTrack — обработчик загрузки музыкального трека
func (h *TrackHandler) UploadTrack(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста (из middleware)
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

	// 6. 🔥 ВЫЗЫВАЕМ PYTHON-СКРИПТ ДЛЯ АНАЛИЗА АУДИО 🔥
	analysis, err := utils.AnalyzeAudioFile(filePath)
	if err != nil {
		// Если анализ не удался — логируем ошибку, но не прерываем загрузку
		// Трек всё равно сохранится, просто без данных волны
		_ = err // Подавляем предупреждение "unused variable"
		// В будущем: log.Printf("Audio analysis failed: %v", err)
	}

	// 7. Создаём запись в базе данных
	track := model.Track{
		UserID:   userID.(uint),
		Title:    file.Filename,
		FileName: file.Filename,
		FilePath: filePath,
		FileSize: file.Size,
	}

	// 8. Если анализ успешен — добавляем данные волны в трек
	if analysis != nil && analysis.Success {
		track.Duration = analysis.Duration

		// ✅ Конвертируем []float64 в datatypes.JSON через json.Marshal
		jsonBytes, err := json.Marshal(analysis.WaveformData)
		if err == nil {
			track.WaveformData = datatypes.JSON(jsonBytes)
		}
	}

	if err := h.DB.Create(&track).Error; err != nil {
		// Если ошибка БД — удаляем файл с диска
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, UploadTrackResponse{
			Error: "Failed to save track to database",
		})
		return
	}

	// 9. Успешный ответ
	c.JSON(http.StatusCreated, UploadTrackResponse{
		Message: "Track uploaded successfully",
		Track:   &track,
	})
}

// GetTracks — обработчик получения списка треков пользователя
func (h *TrackHandler) GetTracks(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста (из middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, GetTracksResponse{
			Error: "User not authenticated",
		})
		return
	}

	// 2. Ищем все треки этого пользователя
	var tracks []model.Track
	if err := h.DB.Where("user_id = ?", userID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GetTracksResponse{
			Error: "Failed to fetch tracks",
		})
		return
	}

	// 3. Успешный ответ
	c.JSON(http.StatusOK, GetTracksResponse{
		Message: "Tracks retrieved successfully",
		Tracks:  tracks,
	})
}
