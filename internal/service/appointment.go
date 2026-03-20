package service

import (
	"errors"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type AppointmentService struct {
	repo     *repository.AppointmentRepository
	dateRepo *repository.AvailableDateRepository
}

func NewAppointmentService(repo *repository.AppointmentRepository, dateRepo *repository.AvailableDateRepository) *AppointmentService {
	return &AppointmentService{repo: repo, dateRepo: dateRepo}
}

func (s *AppointmentService) CreateAppointment(req model.CreateAppointmentRequest) (*model.Appointment, error) {
	if req.ClientName == "" || req.Service == "" || req.Date == "" || req.Time == "" {
		return nil, errors.New("все поля обязательны")
	}

	if req.Telegram == "" && req.Phone == "" {
		return nil, errors.New("укажите Telegram или телефон")
	}

	available, err := s.dateRepo.IsAvailable(req.Date)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("мастер не работает в этот день")
	}

	bookedSlots, err := s.repo.GetBookedSlots(req.Date)
	if err != nil {
		return nil, err
	}
	for _, slot := range bookedSlots {
		if slot == req.Time {
			return nil, errors.New("этот слот уже занят")
		}
	}

	apt := &model.Appointment{
		ClientName: req.ClientName,
		Telegram:   req.Telegram,
		Phone:      req.Phone,
		Service:    req.Service,
		Date:       req.Date,
		Time:       req.Time,
	}

	if err := s.repo.Create(apt); err != nil {
		return nil, err
	}
	return apt, nil
}

func (s *AppointmentService) GetByDate(date string) ([]model.Appointment, error) {
	return s.repo.GetByDate(date)
}

func (s *AppointmentService) GetByDateRange(startDate, endDate string) ([]model.Appointment, error) {
	return s.repo.GetByDateRange(startDate, endDate)
}

func (s *AppointmentService) GetBookedSlots(date string) ([]string, error) {
	return s.repo.GetBookedSlots(date)
}

func (s *AppointmentService) GetAll() ([]model.Appointment, error) {
	return s.repo.GetAll()
}

func (s *AppointmentService) Delete(id uint) error {
	return s.repo.Delete(id)
}
