package service

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

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

	// Проверка пересечения с существующими записями
	existing, err := s.repo.GetByDate(req.Date)
	if err != nil {
		return nil, err
	}

	durationMin := req.DurationMin
	if durationMin <= 0 {
		durationMin = 60
	}

	newStart := timeToMinutes(req.Time)
	newEnd := newStart + durationMin

	for _, apt := range existing {
		aptStart := timeToMinutes(apt.Time)
		aptDur := apt.DurationMin
		if aptDur <= 0 {
			aptDur = 60
		}
		aptEnd := aptStart + aptDur

		if newStart < aptEnd && newEnd > aptStart {
			return nil, fmt.Errorf("пересечение с записью в %s (%s)", apt.Time, apt.Service)
		}
	}

	apt := &model.Appointment{
		ClientName:  req.ClientName,
		Telegram:    req.Telegram,
		Phone:       req.Phone,
		Service:     req.Service,
		DurationMin: durationMin,
		Date:        req.Date,
		Time:        req.Time,
	}

	if err := s.repo.Create(apt); err != nil {
		return nil, err
	}
	return apt, nil
}

// GetBookedSlots возвращает все слоты, занятые с учётом длительности
func (s *AppointmentService) GetBookedSlots(date string) ([]string, error) {
	existing, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, err
	}

	bookedSet := make(map[string]bool)
	allSlots := generateSlots()

	for _, apt := range existing {
		aptStart := timeToMinutes(apt.Time)
		aptDur := apt.DurationMin
		if aptDur <= 0 {
			aptDur = 60
		}
		aptEnd := aptStart + aptDur

		for _, slot := range allSlots {
			slotMin := timeToMinutes(slot)
			if slotMin >= aptStart && slotMin < aptEnd {
				bookedSet[slot] = true
			}
		}
	}

	var booked []string
	for slot := range bookedSet {
		booked = append(booked, slot)
	}
	sort.Strings(booked)
	return booked, nil
}

func (s *AppointmentService) GetByDate(date string) ([]model.Appointment, error) {
	return s.repo.GetByDate(date)
}

func (s *AppointmentService) GetByDateRange(startDate, endDate string) ([]model.Appointment, error) {
	return s.repo.GetByDateRange(startDate, endDate)
}

func (s *AppointmentService) GetAll() ([]model.Appointment, error) {
	return s.repo.GetAll()
}

func (s *AppointmentService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// Вспомогательные функции

func timeToMinutes(t string) int {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m
}

func generateSlots() []string {
	var slots []string
	for h := 9; h <= 20; h++ {
		slots = append(slots, fmt.Sprintf("%02d:00", h))
		if h < 20 {
			slots = append(slots, fmt.Sprintf("%02d:30", h))
		}
	}
	return slots
}
