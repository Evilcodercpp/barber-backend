package repository

import (
	"barber-backend/internal/model"

	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(apt *model.Appointment) error {
	return r.db.Create(apt).Error
}

func (r *AppointmentRepository) GetByDate(date string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("date = ?", date).Order("time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetByDateRange(startDate, endDate string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC, time ASC").
		Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetBookedSlots(date string) ([]string, error) {
	var times []string
	err := r.db.Model(&model.Appointment{}).
		Where("date = ?", date).
		Pluck("time", &times).Error
	return times, err
}

func (r *AppointmentRepository) GetAll() ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Order("date DESC, time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Appointment{}, id).Error
}
