package model

import (
	"time"

	"gorm.io/gorm"
)

// User представляет пользователя в системе
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // "-" скрывает поле при выводе JSON
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Для "мягкого" удаления (soft delete)
}

// TableName явно задаёт имя таблицы в БД
func (User) TableName() string {
	return "users"
}
