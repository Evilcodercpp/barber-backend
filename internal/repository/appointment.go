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

func (r *AppointmentRepository) GetByID(id uint) (*model.Appointment, error) {
	var apt model.Appointment
	if err := r.db.First(&apt, id).Error; err != nil {
		return nil, err
	}
	return &apt, nil
}

func (r *AppointmentRepository) Update(apt *model.Appointment) error {
	return r.db.Save(apt).Error
}

func (r *AppointmentRepository) GetByDate(date string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("date = ?", date).Order("time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetByDateRange(startDate, endDate string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC, time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetCompletedByDateRange(startDate, endDate string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("date >= ? AND date <= ? AND status = ?", startDate, endDate, "completed").
		Order("date ASC, time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetAll() ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Order("date DESC, time ASC").Find(&appointments).Error
	return appointments, err
}

func (r *AppointmentRepository) GetByContact(telegram, phone string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	query := r.db.Order("date DESC, time ASC")
	if telegram != "" {
		query = query.Where("telegram = ?", telegram)
	} else if phone != "" {
		query = query.Where("phone = ?", phone)
	} else {
		return appointments, nil
	}
	return appointments, query.Find(&appointments).Error
}

func (r *AppointmentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Appointment{}, id).Error
}

// GetForReminder возвращает записи на указанную дату, которым ещё не отправлено напоминание
func (r *AppointmentRepository) GetForReminder(date string) ([]model.Appointment, error) {
	var apts []model.Appointment
	err := r.db.Where(
		"date = ? AND status IN ? AND telegram != '' AND reminder_sent = false",
		date, []string{"active", "rescheduled"},
	).Find(&apts).Error
	return apts, err
}

// MarkReminderSent помечает запись как отправленную
func (r *AppointmentRepository) MarkReminderSent(id uint) error {
	return r.db.Model(&model.Appointment{}).Where("id = ?", id).Update("reminder_sent", true).Error
}
