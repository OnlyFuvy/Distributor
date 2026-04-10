package model

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Track представляет музыкальный трек пользователя
type Track struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	Title        string         `gorm:"not null" json:"title"`
	FilePath     string         `gorm:"not null" json:"file_path"`
	FileName     string         `gorm:"not null" json:"file_name"`
	FileSize     int64          `gorm:"not null" json:"file_size"`
	Duration     float64        `json:"duration"`
	WaveformData datatypes.JSON `gorm:"type:jsonb" json:"waveform_data"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Связь с пользователем (для GORM)
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// GetWaveformArray — конвертирует JSON в массив float64 для API
func (t *Track) GetWaveformArray() []float64 {
	if t.WaveformData == nil {
		return []float64{}
	}

	var waveform []float64
	if err := json.Unmarshal(t.WaveformData, &waveform); err != nil {
		return []float64{}
	}
	return waveform
}

// TableName явно задаёт имя таблицы в БД
func (Track) TableName() string {
	return "tracks"
}
