package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type AppointmentService struct {
	repo          appointmentRepo
	dateRepo      availableDateRepo
	clientRepo    clientRepo
	svcRepo       serviceRepo
	svcSupplyRepo serviceSupplyRepo
	supplyRepo    supplyRepo
}

func NewAppointmentService(
	repo *repository.AppointmentRepository,
	dateRepo *repository.AvailableDateRepository,
	clientRepo *repository.ClientRepository,
	svcRepo *repository.ServiceRepository,
	svcSupplyRepo *repository.ServiceSupplyRepository,
	supplyRepo *repository.SupplyRepository,
) *AppointmentService {
	return &AppointmentService{
		repo:          repo,
		dateRepo:      dateRepo,
		clientRepo:    clientRepo,
		svcRepo:       svcRepo,
		svcSupplyRepo: svcSupplyRepo,
		supplyRepo:    supplyRepo,
	}
}

// ConflictError возвращается при пересечении записей — содержит конфликтующую запись
type ConflictError struct {
	ConflictingApt *model.Appointment
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("пересечение с записью в %s (%s)", e.ConflictingApt.Time, e.ConflictingApt.Service)
}

func (s *AppointmentService) CreateAppointment(req model.CreateAppointmentRequest) (*model.Appointment, error) {
	if req.ClientName == "" || req.Service == "" || req.Date == "" || req.Time == "" {
		return nil, errors.New("все поля обязательны")
	}
	if req.Telegram == "" && req.Phone == "" {
		return nil, errors.New("укажите Telegram или телефон")
	}

	durationMin := req.DurationMin

	// Individual-time bookings skip availability and conflict checks
	if req.Time != "по договорённости" {
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
				conflict := apt
				return nil, &ConflictError{ConflictingApt: &conflict}
			}
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

	wasCompleted := apt.Status == "completed"

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
	if req.Service != "" {
		apt.Service = req.Service
	}
	if req.DurationMin > 0 {
		apt.DurationMin = req.DurationMin
	}
	if req.Price > 0 {
		apt.Price = req.Price
	}
	apt.Tips = req.Tips
	apt.Rent = req.Rent
	apt.LateMin = req.LateMin
	if req.SuppliesUsed != "" {
		apt.SuppliesUsed = req.SuppliesUsed
	}
	if req.Comment != "" {
		apt.Comment = req.Comment
	}
	apt.MasterComment = req.MasterComment
	if req.ActualEndTime != "" {
		apt.ActualEndTime = req.ActualEndTime
	}
	if req.PaymentStatus != "" {
		apt.PaymentStatus = req.PaymentStatus
	}
	if req.PaymentDate != "" {
		apt.PaymentDate = req.PaymentDate
	}
	if req.PaidAmount > 0 {
		apt.PaidAmount = req.PaidAmount
	}
	if req.PaymentMethod != "" {
		apt.PaymentMethod = req.PaymentMethod
	}

	// Проверка конфликтов при изменении времени, даты или длительности
	if req.Time != "" || req.Date != "" || req.DurationMin > 0 {
		existing, err := s.repo.GetByDate(apt.Date)
		if err == nil {
			newStart := timeToMinutes(apt.Time)
			newDur := apt.DurationMin
			if newDur <= 0 {
				newDur = 60
			}
			newEnd := newStart + newDur
			for _, other := range existing {
				if other.ID == apt.ID || other.Status == "cancelled" {
					continue
				}
				otherStart := timeToMinutes(other.Time)
				otherDur := other.DurationMin
				if otherDur <= 0 {
					otherDur = 60
				}
				otherEnd := otherStart + otherDur
				if newStart < otherEnd && newEnd > otherStart {
					conflict := other
					return nil, &ConflictError{ConflictingApt: &conflict}
				}
			}
		}
	}

	// Рассчитать и сохранить себестоимость расходников при первом завершении
	if !wasCompleted && apt.Status == "completed" {
		apt.SupplyCost = s.computeSupplyCost(apt)
	}

	if err := s.repo.Update(apt); err != nil {
		return nil, err
	}

	// Списать расходники со склада при первом завершении
	if !wasCompleted && apt.Status == "completed" {
		s.deductSupplies(apt)
	}

	return apt, nil
}

// resolveUsedItems возвращает список расходников для записи:
// приоритет — ручной список из apt.SuppliesUsed, иначе шаблон услуги.
func (s *AppointmentService) resolveUsedItems(apt *model.Appointment) []model.SupplyUsedItem {
	if apt.SuppliesUsed != "" {
		var items []model.SupplyUsedItem
		if err := json.Unmarshal([]byte(apt.SuppliesUsed), &items); err == nil {
			var result []model.SupplyUsedItem
			for _, it := range items {
				if it.SupplyID > 0 && it.Quantity > 0 {
					result = append(result, it)
				}
			}
			return result
		}
	}

	svc, err := s.svcRepo.GetByName(apt.Service)
	if err != nil {
		return nil
	}
	template, err := s.svcSupplyRepo.GetByServiceIDRaw(svc.ID)
	if err != nil {
		return nil
	}
	var result []model.SupplyUsedItem
	for _, t := range template {
		if t.Quantity > 0 {
			result = append(result, model.SupplyUsedItem{SupplyID: t.SupplyID, Quantity: t.Quantity})
		}
	}
	return result
}

// computeSupplyCost рассчитывает суммарную себестоимость расходников в рублях.
func (s *AppointmentService) computeSupplyCost(apt *model.Appointment) int {
	items := s.resolveUsedItems(apt)
	var total float64
	for _, it := range items {
		supply, err := s.supplyRepo.GetByID(it.SupplyID)
		if err != nil {
			continue
		}
		if supply.QuantityGrams > 0 && supply.TotalCost > 0 {
			costPerUnit := supply.TotalCost / supply.QuantityGrams
			total += costPerUnit * it.Quantity
		}
	}
	return int(total + 0.5) // округление
}

// deductSupplies списывает расходники со склада после завершения записи.
func (s *AppointmentService) deductSupplies(apt *model.Appointment) {
	for _, it := range s.resolveUsedItems(apt) {
		s.supplyRepo.DeductQuantity(it.SupplyID, it.Quantity) //nolint:errcheck — ошибка логируется в repo
	}
}

// AvailableSlotsResponse — список доступных слотов для клиента
type AvailableSlotsResponse struct {
	Slots     []string `json:"slots"`
	WorkStart string   `json:"work_start"`
	WorkEnd   string   `json:"work_end"`
}

// GetAvailableSlots возвращает доступные временные слоты с шагом 30 минут,
// учитывая рабочее время мастера и уже занятые интервалы.
func (s *AppointmentService) GetAvailableSlots(date string, durationMin int) (*AvailableSlotsResponse, error) {
	if durationMin <= 0 {
		durationMin = 60
	}

	workStart, workEnd := "10:00", "19:00"

	dateInfo, err := s.dateRepo.GetByDate(date)
	if err == nil {
		workStart = dateInfo.WorkStart
		workEnd = dateInfo.WorkEnd
		if dateInfo.Closed {
			return &AvailableSlotsResponse{Slots: []string{}, WorkStart: workStart, WorkEnd: workEnd}, nil
		}
	}

	existing, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, err
	}

	type interval struct{ start, end int }
	var occupied []interval
	for _, apt := range existing {
		if apt.Status == "cancelled" {
			continue
		}
		aptStart := timeToMinutes(apt.Time)
		aptDur := apt.DurationMin
		if aptDur <= 0 {
			aptDur = 60
		}
		occupied = append(occupied, interval{aptStart, aptStart + aptDur})
	}

	wsMin := timeToMinutes(workStart)
	weMin := timeToMinutes(workEnd)

	var slots []string
	for m := wsMin; m < weMin; m += 30 {
		slotEnd := m + durationMin
		if slotEnd > weMin {
			break
		}
		avail := true
		for _, o := range occupied {
			if m < o.end && slotEnd > o.start {
				avail = false
				break
			}
		}
		if avail {
			h := m / 60
			min := m % 60
			slots = append(slots, fmt.Sprintf("%02d:%02d", h, min))
		}
	}

	if slots == nil {
		slots = []string{}
	}
	return &AvailableSlotsResponse{Slots: slots, WorkStart: workStart, WorkEnd: workEnd}, nil
}

// OccupiedSlot — занятый интервал времени
type OccupiedSlot struct {
	Time        string `json:"time"`
	DurationMin int    `json:"duration_min"`
}

func (s *AppointmentService) GetBookedSlots(date string) ([]OccupiedSlot, error) {
	existing, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, err
	}
	var result []OccupiedSlot
	for _, apt := range existing {
		if apt.Status == "cancelled" {
			continue
		}
		dur := apt.DurationMin
		if dur <= 0 {
			dur = 60
		}
		result = append(result, OccupiedSlot{Time: apt.Time, DurationMin: dur})
	}
	return result, nil
}

func (s *AppointmentService) GetByContact(telegram, phone string) ([]model.Appointment, error) {
	return s.repo.GetByContact(telegram, phone)
}

func (s *AppointmentService) GetFinanceSummary(startDate, endDate, mode string) (*model.FinanceSummary, error) {
	var appointments []model.Appointment
	var err error
	if mode == "cash" {
		appointments, err = s.repo.GetCompletedByPaymentDate(startDate, endDate)
	} else {
		appointments, err = s.repo.GetCompletedByDateRange(startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	summary := &model.FinanceSummary{Appointments: appointments}
	for _, apt := range appointments {
		summary.TotalRevenue += apt.Price
		summary.TotalTips += apt.Tips
		summary.TotalRent += apt.Rent
		summary.TotalSupplyCost += apt.SupplyCost
	}
	summary.Profit = summary.TotalRevenue + summary.TotalTips - summary.TotalRent - summary.TotalSupplyCost

	return summary, nil
}

// GetUnpaid возвращает все неоплаченные и частично оплаченные записи
func (s *AppointmentService) GetUnpaid() ([]model.Appointment, error) {
	return s.repo.GetUnpaid()
}

func (s *AppointmentService) GetByID(id uint) (*model.Appointment, error) {
	return s.repo.GetByID(id)
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

// SetLate отмечает опоздание клиента. При shiftTime=true сдвигает время начала и проверяет конфликты.
func (s *AppointmentService) SetLate(id uint, lateMin int, shiftTime bool) (*model.Appointment, error) {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("запись не найдена")
	}

	apt.LateMin = lateMin
	apt.Status = "late"

	if shiftTime {
		newMinutes := timeToMinutes(apt.Time) + lateMin
		apt.Time = fmt.Sprintf("%02d:%02d", newMinutes/60, newMinutes%60)

		existing, err := s.repo.GetByDate(apt.Date)
		if err == nil {
			newStart := timeToMinutes(apt.Time)
			newEnd := newStart + apt.DurationMin
			if newEnd <= newStart {
				newEnd = newStart + 60
			}
			for _, other := range existing {
				if other.ID == apt.ID || other.Status == "cancelled" {
					continue
				}
				otherStart := timeToMinutes(other.Time)
				otherDur := other.DurationMin
				if otherDur <= 0 {
					otherDur = 60
				}
				if newStart < otherStart+otherDur && newEnd > otherStart {
					conflict := other
					return nil, &ConflictError{ConflictingApt: &conflict}
				}
			}
		}
	}

	if err := s.repo.Update(apt); err != nil {
		return nil, err
	}
	return apt, nil
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
