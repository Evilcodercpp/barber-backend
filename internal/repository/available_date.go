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
	err := r.db.Where("date >= ? AND date <= ?", startDate, endDate).Order("date ASC").Find(&dates).Error
	return dates, err
}

func (r *AvailableDateRepository) IsAvailable(date string) (bool, error) {
	var d model.AvailableDate
	err := r.db.Where("date = ? AND closed = false", date).First(&d).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}

func (r *AvailableDateRepository) CloseDate(date string) error {
	return r.db.Model(&model.AvailableDate{}).Where("date = ?", date).Update("closed", true).Error
}

func (r *AvailableDateRepository) OpenDate(date string) error {
	return r.db.Model(&model.AvailableDate{}).Where("date = ?", date).Update("closed", false).Error
}

func (r *AvailableDateRepository) UpdateHours(date, workStart, workEnd string) error {
	return r.db.Model(&model.AvailableDate{}).Where("date = ?", date).Updates(map[string]interface{}{
		"work_start": workStart,
		"work_end":   workEnd,
	}).Error
}
