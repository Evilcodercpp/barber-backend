package repository

import (
	"barber-backend/internal/model"

	"gorm.io/gorm"
)

type AvailableDateRepository struct {
	db *gorm.DB
}

func NewAvailableDateRepository(db *gorm.DB) *AvailableDateRepository {
	return &AvailableDateRepository{db: db}
}

func (r *AvailableDateRepository) Add(date string) error {
	return r.db.Create(&model.AvailableDate{Date: date}).Error
}

func (r *AvailableDateRepository) Remove(date string) error {
	return r.db.Where("date = ?", date).Delete(&model.AvailableDate{}).Error
}

func (r *AvailableDateRepository) GetAll() ([]model.AvailableDate, error) {
	var dates []model.AvailableDate
	err := r.db.Order("date ASC").Find(&dates).Error
	return dates, err
}

func (r *AvailableDateRepository) GetByRange(startDate, endDate string) ([]model.AvailableDate, error) {
	var dates []model.AvailableDate
	err := r.db.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC").Find(&dates).Error
	return dates, err
}

func (r *AvailableDateRepository) IsAvailable(date string) (bool, error) {
	var count int64
	err := r.db.Model(&model.AvailableDate{}).Where("date = ?", date).Count(&count).Error
	return count > 0, err
}
