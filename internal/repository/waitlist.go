package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type WaitlistRepository struct {
	db *gorm.DB
}

func NewWaitlistRepository(db *gorm.DB) *WaitlistRepository {
	return &WaitlistRepository{db: db}
}

func (r *WaitlistRepository) Create(e *model.WaitlistEntry) error {
	return r.db.Create(e).Error
}

func (r *WaitlistRepository) GetAll() ([]model.WaitlistEntry, error) {
	var entries []model.WaitlistEntry
	err := r.db.Order("date ASC, created_at ASC").Find(&entries).Error
	return entries, err
}

func (r *WaitlistRepository) GetByDate(date string) ([]model.WaitlistEntry, error) {
	var entries []model.WaitlistEntry
	err := r.db.Where("date = ? AND status = 'waiting'", date).
		Order("created_at ASC").Find(&entries).Error
	return entries, err
}

func (r *WaitlistRepository) CountWaiting(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.WaitlistEntry{}).
		Where("date = ? AND status = 'waiting'", date).Count(&count).Error
	return count, err
}

func (r *WaitlistRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.WaitlistEntry{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *WaitlistRepository) Delete(id uint) error {
	return r.db.Delete(&model.WaitlistEntry{}, id).Error
}

// DeleteExpired удаляет записи листа ожидания, дата которых уже прошла (статус waiting/notified)
func (r *WaitlistRepository) DeleteExpired(today string) error {
	return r.db.Where("date < ? AND status IN ('waiting','notified')", today).
		Delete(&model.WaitlistEntry{}).Error
}
