package model

import (
	"time"

	"gorm.io/gorm"
)

// Track представляет музыкальный трек пользователя
type Track struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`   // Связь с пользователем
	Title        string         `gorm:"not null" json:"title"`           // Название трека
	FilePath     string         `gorm:"not null" json:"file_path"`       // Путь к файлу на диске
	FileName     string         `gorm:"not null" json:"file_name"`       // Оригинальное имя файла
	FileSize     int64          `gorm:"not null" json:"file_size"`       // Размер файла в байтах
	Duration     float64        `json:"duration"`                        // Длительность в секундах (заполняется позже)
	WaveformData []float64      `gorm:"type:jsonb" json:"waveform_data"` // Данные волны (Фурье) в формате JSON
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Связь с пользователем (для GORM)
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName явно задаёт имя таблицы в БД
func (Track) TableName() string {
	return "tracks"
}
