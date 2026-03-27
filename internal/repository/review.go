package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

// normalizePhone оставляет только цифры и берёт последние 10 (для сравнения +7 vs 8 префикса)
func normalizePhone(phone string) string {
	digits := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits += string(c)
		}
	}
	if len(digits) >= 10 {
		return digits[len(digits)-10:]
	}
	return digits
}

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

// GetCompletedByPhone возвращает завершённые записи клиента по номеру телефона (с нормализацией)
func (r *ReviewRepository) GetCompletedByPhone(phone string) ([]model.Appointment, error) {
	var all []model.Appointment
	err := r.db.Where("status = ?", "completed").Order("date DESC").Find(&all).Error
	if err != nil {
		return nil, err
	}
	norm := normalizePhone(phone)
	var apts []model.Appointment
	for _, a := range all {
		if normalizePhone(a.Phone) == norm {
			apts = append(apts, a)
		}
	}
	return apts, nil
}

// GetCompletedByTelegram возвращает завершённые записи клиента по Telegram (без @, без учёта регистра)
func (r *ReviewRepository) GetCompletedByTelegram(telegram string) ([]model.Appointment, error) {
	tg := normalizeTelegram(telegram)
	var all []model.Appointment
	err := r.db.Where("status = ?", "completed").Order("date DESC").Find(&all).Error
	if err != nil {
		return nil, err
	}
	var apts []model.Appointment
	for _, a := range all {
		if normalizeTelegram(a.Telegram) == tg {
			apts = append(apts, a)
		}
	}
	return apts, nil
}

func normalizeTelegram(t string) string {
	if len(t) > 0 && t[0] == '@' {
		t = t[1:]
	}
	result := ""
	for _, c := range t {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			if c >= 'A' && c <= 'Z' {
				result += string(c + 32)
			} else {
				result += string(c)
			}
		}
	}
	return result
}
