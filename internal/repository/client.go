package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(c *model.Client) error {
	return r.db.Create(c).Error
}

func (r *ClientRepository) GetAll() ([]model.Client, error) {
	var clients []model.Client
	err := r.db.Order("created_at DESC").Find(&clients).Error
	return clients, err
}

func (r *ClientRepository) Update(c *model.Client) error {
	return r.db.Save(c).Error
}

func (r *ClientRepository) Delete(id uint) error {
	return r.db.Delete(&model.Client{}, id).Error
}

// FindOrCreate — добавляет клиента в базу, если нет — или дополняет существующего новыми данными
func (r *ClientRepository) FindOrCreate(name, telegram, phone string) {
	var existing model.Client
	found := false

	// Ищем по Telegram
	if telegram != "" {
		if err := r.db.Where("telegram = ?", telegram).First(&existing).Error; err == nil {
			found = true
		}
	}

	// Если не нашли по Telegram — ищем по телефону
	if !found && phone != "" {
		if err := r.db.Where("phone = ?", phone).First(&existing).Error; err == nil {
			found = true
		}
	}

	if found {
		// Дополняем существующую запись новыми данными (не перезаписываем старые)
		changed := false
		if existing.Telegram == "" && telegram != "" {
			existing.Telegram = telegram
			changed = true
		}
		if existing.Phone == "" && phone != "" {
			existing.Phone = phone
			changed = true
		}
		if changed {
			r.db.Save(&existing)
		}
		return
	}

	// Создаём нового клиента
	r.db.Create(&model.Client{Name: name, Telegram: telegram, Phone: phone})
}
