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

// FindOrCreate — добавляет клиента в базу если его ещё нет (по telegram или phone)
func (r *ClientRepository) FindOrCreate(name, telegram, phone string) {
	var existing model.Client
	query := r.db
	if telegram != "" {
		query = query.Where("telegram = ?", telegram)
	} else if phone != "" {
		query = query.Where("phone = ?", phone)
	} else {
		return
	}
	if err := query.First(&existing).Error; err == nil {
		return // уже существует
	}
	r.db.Create(&model.Client{Name: name, Telegram: telegram, Phone: phone})
}
