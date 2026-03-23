package repository

import (
	"time"

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

// GetByClientName ищет записи по имени клиента (для карточки)
func (r *AppointmentRepository) GetByClientName(name string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("client_name = ?", name).Order("date DESC, time ASC").Find(&appointments).Error
	return appointments, err
}

// GetUnpaid возвращает все записи с payment_status = unpaid или partial
func (r *AppointmentRepository) GetUnpaid() ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where("payment_status IN ?", []string{"unpaid", "partial"}).
		Order("payment_date ASC, date ASC").Find(&appointments).Error
	return appointments, err
}

// GetCompletedByPaymentDate возвращает завершённые записи по диапазону даты оплаты
func (r *AppointmentRepository) GetCompletedByPaymentDate(startDate, endDate string) ([]model.Appointment, error) {
	var appointments []model.Appointment
	err := r.db.Where(
		"payment_status = 'paid' AND COALESCE(NULLIF(payment_date,''), date) >= ? AND COALESCE(NULLIF(payment_date,''), date) <= ?",
		startDate, endDate,
	).Order("date ASC, time ASC").Find(&appointments).Error
	return appointments, err
}

// GetForReminder возвращает кандидатов на напоминание за ближайшие 3 дня
// (финальная фильтрация по окну 20-28ч делается в Go)
func (r *AppointmentRepository) GetForReminder() ([]model.Appointment, error) {
	now := time.Now()
	dates := []string{
		now.Format("2006-01-02"),
		now.AddDate(0, 0, 1).Format("2006-01-02"),
		now.AddDate(0, 0, 2).Format("2006-01-02"),
	}
	var apts []model.Appointment
	err := r.db.Where(
		"date IN ? AND status IN ? AND telegram != '' AND reminder_sent = false",
		dates, []string{"active", "rescheduled"},
	).Find(&apts).Error
	return apts, err
}

// MarkReminderSent помечает запись как отправленную
func (r *AppointmentRepository) MarkReminderSent(id uint) error {
	return r.db.Model(&model.Appointment{}).Where("id = ?", id).Update("reminder_sent", true).Error
}
