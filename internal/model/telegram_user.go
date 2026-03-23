package model

import "time"

// TelegramUser — клиент, написавший боту /start
type TelegramUser struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ChatID    int64     `json:"chat_id" gorm:"uniqueIndex;not null"`
	Username  string    `json:"username" gorm:"index"` // lowercase, without @
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
