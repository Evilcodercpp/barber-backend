package repository

import (
	"barber-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TelegramUserRepository struct {
	db *gorm.DB
}

func NewTelegramUserRepository(db *gorm.DB) *TelegramUserRepository {
	return &TelegramUserRepository{db: db}
}

// Upsert сохраняет или обновляет chat_id → username
func (r *TelegramUserRepository) Upsert(chatID int64, username string) error {
	u := model.TelegramUser{ChatID: chatID, Username: username}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "updated_at"}),
	}).Create(&u).Error
}

// GetByUsername ищет пользователя по нику (без @, lowercase)
func (r *TelegramUserRepository) GetByUsername(username string) (*model.TelegramUser, error) {
	var u model.TelegramUser
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
