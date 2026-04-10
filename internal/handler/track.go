package handler

import (
	"encoding/json"
	"fmt"
	"music-app/internal/model"
	"music-app/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

// TrackResponse — структура трека для API ответа (с правильным типом waveform_data)
type TrackResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	Title        string    `json:"title"`
	FilePath     string    `json:"file_path"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	Duration     float64   `json:"duration"`
	WaveformData []float64 `json:"waveform_data"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetTracksResponse — ответ сервера со списком треков
type GetTracksResponse struct {
	Message string          `json:"message"`
	Tracks  []TrackResponse `json:"tracks,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// DeleteTrackResponse — ответ сервера при удалении трека
type DeleteTrackResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
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

	// 3. ✅ Конвертируем waveform_data для каждого трека
	//    Создаём новый слайс с правильными данными для API
	var responseTracks []TrackResponse
	for _, track := range tracks {
		// Конвертируем datatypes.JSON в []float64
		var waveform []float64
		if track.WaveformData != nil {
			json.Unmarshal(track.WaveformData, &waveform)
		}

		responseTracks = append(responseTracks, TrackResponse{
			ID:           track.ID,
			UserID:       track.UserID,
			Title:        track.Title,
			FilePath:     track.FilePath,
			FileName:     track.FileName,
			FileSize:     track.FileSize,
			Duration:     track.Duration,
			WaveformData: waveform, // ← Теперь массив, а не JSON
			CreatedAt:    track.CreatedAt,
			UpdatedAt:    track.UpdatedAt,
		})
	}

	// 4. Успешный ответ
	c.JSON(http.StatusOK, GetTracksResponse{
		Message: "Tracks retrieved successfully",
		Tracks:  responseTracks,
	})
}

// DeleteTrack — обработчик удаления трека
func (h *TrackHandler) DeleteTrack(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста (из middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, DeleteTrackResponse{
			Error: "User not authenticated",
		})
		return
	}

	// 2. Получаем ID трека из URL параметра
	trackIDStr := c.Param("id")
	var trackID uint
	if _, err := fmt.Sscanf(trackIDStr, "%d", &trackID); err != nil {
		c.JSON(http.StatusBadRequest, DeleteTrackResponse{
			Error: "Invalid track ID",
		})
		return
	}

	// 3. Находим трек (проверяем, что он принадлежит этому пользователю)
	var track model.Track
	if err := h.DB.Where("id = ? AND user_id = ?", trackID, userID).First(&track).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, DeleteTrackResponse{
				Error: "Track not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, DeleteTrackResponse{
			Error: "Database error",
		})
		return
	}

	// 4. Удаляем файл с диска
	if err := os.Remove(track.FilePath); err != nil {
		// Логируем ошибку, но не прерываем удаление из БД
		// log.Printf("Failed to delete file %s: %v", track.FilePath, err)
	}

	// 5. Удаляем запись из базы данных
	if err := h.DB.Delete(&track).Error; err != nil {
		c.JSON(http.StatusInternalServerError, DeleteTrackResponse{
			Error: "Failed to delete track",
		})
		return
	}

	// 6. Успешный ответ
	c.JSON(http.StatusOK, DeleteTrackResponse{
		Message: "Track deleted successfully",
	})
}
