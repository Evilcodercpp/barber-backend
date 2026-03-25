package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *model.Review) error {
	return r.db.Create(review).Error
}

func (r *ReviewRepository) GetApproved() ([]model.Review, error) {
	var reviews []model.Review
	err := r.db.Where("approved = true").Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepository) GetAll() ([]model.Review, error) {
	var reviews []model.Review
	err := r.db.Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepository) SetApproved(id uint, approved bool) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).Update("approved", approved).Error
}

func (r *ReviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.Review{}, id).Error
}

func (r *ReviewRepository) ExistsForAppointment(appointmentID uint) bool {
	var count int64
	r.db.Model(&model.Review{}).Where("appointment_id = ?", appointmentID).Count(&count)
	return count > 0
}

// GetCompletedByPhone возвращает завершённые записи клиента по номеру телефона
func (r *ReviewRepository) GetCompletedByPhone(phone string) ([]model.Appointment, error) {
	var apts []model.Appointment
	err := r.db.Where("phone = ? AND status = ?", phone, "completed").
		Order("date DESC").Find(&apts).Error
	return apts, err
}
