package service

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type AppointmentService struct {
	repo       *repository.AppointmentRepository
	dateRepo   *repository.AvailableDateRepository
	clientRepo *repository.ClientRepository
}

func NewAppointmentService(
	repo *repository.AppointmentRepository,
	dateRepo *repository.AvailableDateRepository,
	clientRepo *repository.ClientRepository,
) *AppointmentService {
	return &AppointmentService{repo: repo, dateRepo: dateRepo, clientRepo: clientRepo}
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
		if apt.Status == "cancelled" {
			continue
		}
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
		Status:      "active",
		Price:       req.Price,
	}

	if err := s.repo.Create(apt); err != nil {
		return nil, err
	}

	// Автоматически добавить клиента в базу
	s.clientRepo.FindOrCreate(req.ClientName, req.Telegram, req.Phone)

	return apt, nil
}

func (s *AppointmentService) UpdateAppointment(id uint, req model.UpdateAppointmentRequest) (*model.Appointment, error) {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("запись не найдена")
	}

	if req.Status != "" {
		validStatuses := map[string]bool{
			"active": true, "rescheduled": true, "cancelled": true,
			"late": true, "completed": true,
		}
		if !validStatuses[req.Status] {
			return nil, errors.New("неверный статус")
		}
		apt.Status = req.Status
	}
	if req.Date != "" {
		apt.Date = req.Date
		if apt.Status == "active" {
			apt.Status = "rescheduled"
		}
	}
	if req.Time != "" {
		apt.Time = req.Time
	}
	if req.Price > 0 {
		apt.Price = req.Price
	}
	apt.Tips = req.Tips
	apt.Rent = req.Rent

	if err := s.repo.Update(apt); err != nil {
		return nil, err
	}
	return apt, nil
}

func (s *AppointmentService) GetBookedSlots(date string) ([]string, error) {
	existing, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, err
	}

	bookedSet := make(map[string]bool)
	allSlots := generateSlots()

	for _, apt := range existing {
		if apt.Status == "cancelled" {
			continue
		}
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

func (s *AppointmentService) GetByContact(telegram, phone string) ([]model.Appointment, error) {
	return s.repo.GetByContact(telegram, phone)
}

func (s *AppointmentService) GetFinanceSummary(startDate, endDate string) (*model.FinanceSummary, error) {
	appointments, err := s.repo.GetCompletedByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary := &model.FinanceSummary{Appointments: appointments}
	for _, apt := range appointments {
		summary.TotalRevenue += apt.Price
		summary.TotalTips += apt.Tips
		summary.TotalRent += apt.Rent
	}
	summary.Profit = summary.TotalRevenue + summary.TotalTips - summary.TotalRent

	return summary, nil
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

// ParsePrice parses "4 500 ₽" -> 4500
func ParsePrice(priceStr string) int {
	re := regexp.MustCompile(`\d+`)
	nums := re.FindAllString(priceStr, -1)
	if len(nums) == 0 {
		return 0
	}
	result, _ := strconv.Atoi(strings.Join(nums, ""))
	return result
}
