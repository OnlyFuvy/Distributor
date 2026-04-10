package model

import (
	"time"

	"gorm.io/gorm"
)

// VerificationCode — код подтверждения email
type VerificationCode struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Email     string         `gorm:"not null;index" json:"email"`
	Code      string         `gorm:"not null;size:6" json:"code"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName задаёт имя таблицы
func (VerificationCode) TableName() string {
	return "verification_codes"
}

// IsExpired — проверяет, истёк ли срок действия кода
func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}
