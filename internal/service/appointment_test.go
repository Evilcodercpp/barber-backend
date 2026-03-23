package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"barber-backend/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock-реализации интерфейсов репозиториев
// ─────────────────────────────────────────────────────────────────────────────

type mockAppointmentRepo struct {
	appointments map[uint]*model.Appointment
	nextID       uint
	createErr    error
	updateErr    error
}

func newMockAptRepo() *mockAppointmentRepo {
	return &mockAppointmentRepo{appointments: map[uint]*model.Appointment{}, nextID: 1}
}

func (m *mockAppointmentRepo) Create(apt *model.Appointment) error {
	if m.createErr != nil {
		return m.createErr
	}
	apt.ID = m.nextID
	m.nextID++
	cp := *apt
	m.appointments[apt.ID] = &cp
	return nil
}

func (m *mockAppointmentRepo) GetByID(id uint) (*model.Appointment, error) {
	a, ok := m.appointments[id]
	if !ok {
		return nil, errors.New("не найдено")
	}
	cp := *a
	return &cp, nil
}

func (m *mockAppointmentRepo) Update(apt *model.Appointment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *apt
	m.appointments[apt.ID] = &cp
	return nil
}

func (m *mockAppointmentRepo) GetByDate(date string) ([]model.Appointment, error) {
	var result []model.Appointment
	for _, a := range m.appointments {
		if a.Date == date {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAppointmentRepo) GetByDateRange(start, end string) ([]model.Appointment, error) {
	var result []model.Appointment
	for _, a := range m.appointments {
		if a.Date >= start && a.Date <= end {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAppointmentRepo) GetCompletedByDateRange(start, end string) ([]model.Appointment, error) {
	var result []model.Appointment
	for _, a := range m.appointments {
		if a.Date >= start && a.Date <= end && a.Status == "completed" {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAppointmentRepo) GetAll() ([]model.Appointment, error) {
	var result []model.Appointment
	for _, a := range m.appointments {
		result = append(result, *a)
	}
	return result, nil
}

func (m *mockAppointmentRepo) GetByContact(telegram, phone string) ([]model.Appointment, error) {
	var result []model.Appointment
	for _, a := range m.appointments {
		if (telegram != "" && a.Telegram == telegram) || (phone != "" && a.Phone == phone) {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAppointmentRepo) Delete(id uint) error {
	delete(m.appointments, id)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────

type mockDateRepo struct {
	dates map[string]*model.AvailableDate
}

func newMockDateRepo() *mockDateRepo {
	return &mockDateRepo{dates: map[string]*model.AvailableDate{}}
}

func (m *mockDateRepo) addDate(date string, closed bool, start, end string) {
	m.dates[date] = &model.AvailableDate{Date: date, Closed: closed, WorkStart: start, WorkEnd: end}
}

func (m *mockDateRepo) IsAvailable(date string) (bool, error) {
	d, ok := m.dates[date]
	if !ok || d.Closed {
		return false, nil
	}
	return true, nil
}

func (m *mockDateRepo) GetByDate(date string) (*model.AvailableDate, error) {
	d, ok := m.dates[date]
	if !ok {
		return nil, errors.New("не найдено")
	}
	cp := *d
	return &cp, nil
}

// ─────────────────────────────────────────────────────────────────────────────

type mockClientRepo struct {
	calls []string
}

func (m *mockClientRepo) FindOrCreate(name, telegram, phone string) {
	m.calls = append(m.calls, fmt.Sprintf("%s|%s|%s", name, telegram, phone))
}

// ─────────────────────────────────────────────────────────────────────────────

type mockServiceRepo struct {
	services map[string]*model.Service
}

func newMockServiceRepo() *mockServiceRepo {
	return &mockServiceRepo{services: map[string]*model.Service{}}
}

func (m *mockServiceRepo) GetByName(name string) (*model.Service, error) {
	s, ok := m.services[name]
	if !ok {
		return nil, errors.New("услуга не найдена")
	}
	cp := *s
	return &cp, nil
}

// ─────────────────────────────────────────────────────────────────────────────

type mockServiceSupplyRepo struct {
	templates map[uint][]model.ServiceSupply
}

func newMockSSRepo() *mockServiceSupplyRepo {
	return &mockServiceSupplyRepo{templates: map[uint][]model.ServiceSupply{}}
}

func (m *mockServiceSupplyRepo) GetByServiceIDRaw(serviceID uint) ([]model.ServiceSupply, error) {
	return m.templates[serviceID], nil
}

// ─────────────────────────────────────────────────────────────────────────────

type mockSupplyRepo struct {
	supplies   map[uint]*model.Supply
	deductions []struct {
		id  uint
		qty float64
	}
}

func newMockSupplyRepo() *mockSupplyRepo {
	return &mockSupplyRepo{supplies: map[uint]*model.Supply{}}
}

func (m *mockSupplyRepo) addSupply(id uint, qtyGrams, totalCost, stock float64) {
	m.supplies[id] = &model.Supply{
		ID:            id,
		QuantityGrams: qtyGrams,
		TotalCost:     totalCost,
		Quantity:      stock,
	}
}

func (m *mockSupplyRepo) GetByID(id uint) (*model.Supply, error) {
	s, ok := m.supplies[id]
	if !ok {
		return nil, errors.New("расходник не найден")
	}
	cp := *s
	return &cp, nil
}

func (m *mockSupplyRepo) DeductQuantity(id uint, qty float64) error {
	if s, ok := m.supplies[id]; ok {
		s.Quantity -= qty
		if s.Quantity < 0 {
			s.Quantity = 0
		}
	}
	m.deductions = append(m.deductions, struct {
		id  uint
		qty float64
	}{id, qty})
	return nil
}

func (m *mockSupplyRepo) AddQuantity(id uint, qty float64) error {
	if s, ok := m.supplies[id]; ok {
		s.Quantity += qty
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Вспомогательный конструктор сервиса с mock-репозиториями
// ─────────────────────────────────────────────────────────────────────────────

type testDeps struct {
	aptRepo     *mockAppointmentRepo
	dateRepo    *mockDateRepo
	clientRepo  *mockClientRepo
	svcRepo     *mockServiceRepo
	ssRepo      *mockServiceSupplyRepo
	supplyRepo  *mockSupplyRepo
	svc         *AppointmentService
}

func newTestService() *testDeps {
	d := &testDeps{
		aptRepo:    newMockAptRepo(),
		dateRepo:   newMockDateRepo(),
		clientRepo: &mockClientRepo{},
		svcRepo:    newMockServiceRepo(),
		ssRepo:     newMockSSRepo(),
		supplyRepo: newMockSupplyRepo(),
	}
	d.svc = &AppointmentService{
		repo:          d.aptRepo,
		dateRepo:      d.dateRepo,
		clientRepo:    d.clientRepo,
		svcRepo:       d.svcRepo,
		svcSupplyRepo: d.ssRepo,
		supplyRepo:    d.supplyRepo,
	}
	return d
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: timeToMinutes (чистая функция)
// ─────────────────────────────────────────────────────────────────────────────

func TestTimeToMinutes(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"10:00", 600},
		{"10:30", 630},
		{"00:00", 0},
		{"23:59", 1439},
		{"19:00", 1140},
		// Краш-тесты / невалидный ввод
		{"", 0},        // пустая строка
		{"abc", 0},     // нет двоеточия
		{"1030", 0},    // без двоеточия
		{"25:99", 1599}, // за пределами суток — функция не обрезает, просто парсит
		{":30", 30},    // пропущены часы
		{"10:", 600},   // пропущены минуты
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := timeToMinutes(tc.input)
			if got != tc.want {
				t.Errorf("timeToMinutes(%q) = %d, хотим %d", tc.input, got, tc.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: ParsePrice (экспортируемая чистая функция)
// ─────────────────────────────────────────────────────────────────────────────

func TestParsePrice(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"4 500 ₽", 4500},
		{"12 000 ₽", 12000},
		{"1000", 1000},
		{"от 3 000 ₽", 3000},
		{"", 0},
		{"abc", 0},
		{"бесплатно", 0},
		// Несколько групп цифр объединяются
		{"8 500 – 12 500 ₽", 850012500}, // все цифры склеиваются
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := ParsePrice(tc.input)
			if got != tc.want {
				t.Errorf("ParsePrice(%q) = %d, хотим %d", tc.input, got, tc.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: CreateAppointment — валидация
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateAppointment_Validation(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-01", false, "10:00", "19:00")

	base := model.CreateAppointmentRequest{
		ClientName:  "Мария",
		Telegram:    "@maria",
		Phone:       "",
		Service:     "Стрижка",
		Date:        "2026-04-01",
		Time:        "10:00",
		DurationMin: 60,
		Price:       3000,
	}

	t.Run("успешное создание", func(t *testing.T) {
		_, err := d.svc.CreateAppointment(base)
		if err != nil {
			t.Fatalf("ожидаем успех, получили: %v", err)
		}
	})

	t.Run("пустое имя", func(t *testing.T) {
		req := base
		req.ClientName = ""
		_, err := d.svc.CreateAppointment(req)
		if err == nil {
			t.Fatal("должна быть ошибка при пустом имени")
		}
	})

	t.Run("пустая услуга", func(t *testing.T) {
		req := base
		req.Service = ""
		_, err := d.svc.CreateAppointment(req)
		if err == nil {
			t.Fatal("должна быть ошибка при пустой услуге")
		}
	})

	t.Run("пустая дата", func(t *testing.T) {
		req := base
		req.Date = ""
		_, err := d.svc.CreateAppointment(req)
		if err == nil {
			t.Fatal("должна быть ошибка при пустой дате")
		}
	})

	t.Run("пустое время", func(t *testing.T) {
		req := base
		req.Time = ""
		_, err := d.svc.CreateAppointment(req)
		if err == nil {
			t.Fatal("должна быть ошибка при пустом времени")
		}
	})

	t.Run("нет ни telegram ни phone", func(t *testing.T) {
		req := base
		req.Telegram = ""
		req.Phone = ""
		_, err := d.svc.CreateAppointment(req)
		if err == nil {
			t.Fatal("должна быть ошибка если нет контакта")
		}
	})

	t.Run("только телефон — ок", func(t *testing.T) {
		req := base
		req.Telegram = ""
		req.Phone = "+79161234567"
		req.Time = "11:00" // другой слот чтобы не конфликтовать
		_, err := d.svc.CreateAppointment(req)
		if err != nil {
			t.Fatalf("только телефон должен быть допустим: %v", err)
		}
	})
}

func TestCreateAppointment_DayUnavailable(t *testing.T) {
	d := newTestService()
	// День не добавлен в расписание

	req := model.CreateAppointmentRequest{
		ClientName: "Анна", Telegram: "@anna",
		Service: "Стрижка", Date: "2026-04-02", Time: "10:00", DurationMin: 60,
	}
	_, err := d.svc.CreateAppointment(req)
	if err == nil {
		t.Fatal("ожидаем ошибку для недоступного дня")
	}
	if !strings.Contains(err.Error(), "не работает") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

func TestCreateAppointment_ClosedDay(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-03", true, "10:00", "19:00") // выходной

	req := model.CreateAppointmentRequest{
		ClientName: "Ольга", Telegram: "@olga",
		Service: "Стрижка", Date: "2026-04-03", Time: "10:00", DurationMin: 60,
	}
	_, err := d.svc.CreateAppointment(req)
	if err == nil {
		t.Fatal("ожидаем ошибку для закрытого дня")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: CreateAppointment — конфликты времени
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateAppointment_ConflictDetection(t *testing.T) {
	date := "2026-04-05"

	setup := func() *testDeps {
		d := newTestService()
		d.dateRepo.addDate(date, false, "10:00", "19:00")
		// Занимаем 10:00–11:00
		d.aptRepo.appointments[1] = &model.Appointment{
			ID: 1, Date: date, Time: "10:00", DurationMin: 60, Status: "active",
		}
		d.aptRepo.nextID = 2
		return d
	}

	t.Run("прямое пересечение", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "10:00", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		var ce *ConflictError
		if !errors.As(err, &ce) {
			t.Fatalf("ожидаем ConflictError, получили: %v", err)
		}
	})

	t.Run("частичное перекрытие (начало внутри занятого)", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "10:30", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		var ce *ConflictError
		if !errors.As(err, &ce) {
			t.Fatalf("ожидаем ConflictError, получили: %v", err)
		}
	})

	t.Run("частичное перекрытие (конец заходит в занятое)", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "09:30", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		var ce *ConflictError
		if !errors.As(err, &ce) {
			t.Fatalf("ожидаем ConflictError, получили: %v", err)
		}
	})

	t.Run("сразу после занятого — ок", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "11:00", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		if err != nil {
			t.Fatalf("слот после занятого должен быть доступен: %v", err)
		}
	})

	t.Run("сразу до занятого — ок", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "09:00", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		if err != nil {
			t.Fatalf("слот до занятого должен быть доступен: %v", err)
		}
	})

	t.Run("отменённая запись не блокирует слот", func(t *testing.T) {
		d := setup()
		d.aptRepo.appointments[1].Status = "cancelled"
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "10:00", DurationMin: 60,
		}
		_, err := d.svc.CreateAppointment(req)
		if err != nil {
			t.Fatalf("отменённая запись не должна блокировать: %v", err)
		}
	})

	t.Run("нулевая длительность → по умолчанию 60 мин", func(t *testing.T) {
		d := setup()
		req := model.CreateAppointmentRequest{
			ClientName: "Ира", Telegram: "@ira",
			Service: "Стрижка", Date: date, Time: "10:30", DurationMin: 0,
		}
		_, err := d.svc.CreateAppointment(req)
		var ce *ConflictError
		if !errors.As(err, &ce) {
			t.Fatalf("DurationMin=0 должен defaultиться к 60 и конфликтовать: %v", err)
		}
	})
}

func TestCreateAppointment_IndividualTime(t *testing.T) {
	d := newTestService()
	// День НЕ добавлен в расписание — "по договорённости" должно всё равно работать

	req := model.CreateAppointmentRequest{
		ClientName: "Катя", Telegram: "@katya",
		Service: "Окрашивание", Date: "2026-05-01", Time: "по договорённости",
		DurationMin: 180,
	}
	apt, err := d.svc.CreateAppointment(req)
	if err != nil {
		t.Fatalf("'по договорённости' должно игнорировать расписание: %v", err)
	}
	if apt.Status != "active" {
		t.Errorf("статус должен быть active, получили %q", apt.Status)
	}
}

func TestCreateAppointment_AutoCreatesClient(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-07", false, "10:00", "19:00")

	req := model.CreateAppointmentRequest{
		ClientName: "Света", Telegram: "@sveta",
		Service: "Стрижка", Date: "2026-04-07", Time: "10:00", DurationMin: 60,
	}
	_, err := d.svc.CreateAppointment(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.clientRepo.calls) == 0 {
		t.Error("FindOrCreate должна была быть вызвана")
	}
	if !strings.Contains(d.clientRepo.calls[0], "Света") {
		t.Errorf("неверные аргументы FindOrCreate: %v", d.clientRepo.calls)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: UpdateAppointment
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateAppointment_InvalidStatus(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active"}

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "flying"})
	if err == nil {
		t.Fatal("неверный статус должен возвращать ошибку")
	}
}

func TestUpdateAppointment_NotFound(t *testing.T) {
	d := newTestService()
	_, err := d.svc.UpdateAppointment(999, model.UpdateAppointmentRequest{Status: "completed"})
	if err == nil {
		t.Fatal("несуществующий ID должен возвращать ошибку")
	}
}

func TestUpdateAppointment_DateChangeBecomesRescheduled(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active"}

	apt, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Date: "2026-04-11", Time: "10:00"})
	if err != nil {
		t.Fatal(err)
	}
	if apt.Status != "rescheduled" {
		t.Errorf("перенос должен давать статус rescheduled, получили %q", apt.Status)
	}
}

func TestUpdateAppointment_DateChangeKeepsManualStatus(t *testing.T) {
	// Если статус явно задан через req.Status — он приоритетнее автоматического rescheduled
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active"}

	apt, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{
		Status: "cancelled",
		Date:   "2026-04-11",
		Time:   "10:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Status обновляется сначала (cancelled), затем Date меняет только если Status == "active"
	if apt.Status != "cancelled" {
		t.Errorf("статус cancelled должен сохраниться, получили %q", apt.Status)
	}
}

func TestUpdateAppointment_ConflictOnReschedule(t *testing.T) {
	d := newTestService()
	// Две записи: ID=1 и ID=2, ID=2 хотим перенести на время ID=1
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active"}
	d.aptRepo.appointments[2] = &model.Appointment{ID: 2, Date: "2026-04-10", Time: "12:00", DurationMin: 60, Status: "active"}
	d.aptRepo.nextID = 3

	_, err := d.svc.UpdateAppointment(2, model.UpdateAppointmentRequest{Time: "10:00"})
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("ожидаем ConflictError при переносе на занятое время, получили: %v", err)
	}
}

func TestUpdateAppointment_NoSelfConflict(t *testing.T) {
	// Запись не должна конфликтовать сама с собой при обновлении без смены времени
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active"}

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Time: "10:00", DurationMin: 60})
	if err != nil {
		t.Fatalf("запись не должна конфликтовать сама с собой: %v", err)
	}
}

func TestUpdateAppointment_SupplyDeductionOnCompletion(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60,
		Status: "active", Service: "Окрашивание корней",
	}
	// Услуга с шаблоном расходников
	d.svcRepo.services["Окрашивание корней"] = &model.Service{ID: 10, Name: "Окрашивание корней"}
	d.ssRepo.templates[10] = []model.ServiceSupply{
		{SupplyID: 5, Quantity: 50},
		{SupplyID: 7, Quantity: 20},
	}
	d.supplyRepo.addSupply(5, 0, 0, 200)
	d.supplyRepo.addSupply(7, 0, 0, 100)

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "completed"})
	if err != nil {
		t.Fatal(err)
	}
	if d.supplyRepo.supplies[5].Quantity != 150 {
		t.Errorf("расходник 5: ожидаем 150, получили %v", d.supplyRepo.supplies[5].Quantity)
	}
	if d.supplyRepo.supplies[7].Quantity != 80 {
		t.Errorf("расходник 7: ожидаем 80, получили %v", d.supplyRepo.supplies[7].Quantity)
	}
}

func TestUpdateAppointment_NoDoubleDeduction(t *testing.T) {
	// Если запись уже completed — повторное сохранение не должно списывать расходники снова
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60,
		Status: "completed", Service: "Стрижка",
	}
	d.svcRepo.services["Стрижка"] = &model.Service{ID: 20, Name: "Стрижка"}
	d.ssRepo.templates[20] = []model.ServiceSupply{{SupplyID: 3, Quantity: 10}}
	d.supplyRepo.addSupply(3, 0, 0, 100)

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "completed", Comment: "обновлён"})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.supplyRepo.deductions) > 0 {
		t.Error("при повторном completed не должно быть списания расходников")
	}
}

func TestUpdateAppointment_ManualSuppliesOverrideTemplate(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60,
		Status: "active", Service: "Окрашивание корней",
		SuppliesUsed: `[{"supply_id":9,"quantity":30}]`,
	}
	d.svcRepo.services["Окрашивание корней"] = &model.Service{ID: 10, Name: "Окрашивание корней"}
	d.ssRepo.templates[10] = []model.ServiceSupply{{SupplyID: 5, Quantity: 50}}
	d.supplyRepo.addSupply(5, 0, 0, 200)
	d.supplyRepo.addSupply(9, 0, 0, 100)

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "completed"})
	if err != nil {
		t.Fatal(err)
	}
	// Должен быть списан только расходник 9 (ручной), не шаблонный 5
	if d.supplyRepo.supplies[5].Quantity != 200 {
		t.Error("шаблонный расходник не должен был быть списан")
	}
	if d.supplyRepo.supplies[9].Quantity != 70 {
		t.Errorf("ручной расходник 9: ожидаем 70, получили %v", d.supplyRepo.supplies[9].Quantity)
	}
}

func TestUpdateAppointment_InvalidSuppliesJSON(t *testing.T) {
	// Невалидный JSON в SuppliesUsed → откат к шаблону
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60,
		Status: "active", Service: "Стрижка",
		SuppliesUsed: `INVALID_JSON`,
	}
	d.svcRepo.services["Стрижка"] = &model.Service{ID: 20, Name: "Стрижка"}
	d.ssRepo.templates[20] = []model.ServiceSupply{{SupplyID: 3, Quantity: 10}}
	d.supplyRepo.addSupply(3, 0, 0, 100)

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "completed"})
	if err != nil {
		t.Fatal(err)
	}
	// Невалидный JSON → должен использоваться шаблон
	if d.supplyRepo.supplies[3].Quantity != 90 {
		t.Errorf("при невалидном JSON должен использоваться шаблон, расходник 3: ожидаем 90, получили %v", d.supplyRepo.supplies[3].Quantity)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: GetAvailableSlots
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAvailableSlots_EmptyDay(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-15", false, "10:00", "12:00")

	resp, err := d.svc.GetAvailableSlots("2026-04-15", 60)
	if err != nil {
		t.Fatal(err)
	}
	// 10:00–12:00, шаг 30 мин, длительность 60 мин → слоты: 10:00, 10:30, 11:00
	if len(resp.Slots) != 3 {
		t.Errorf("ожидаем 3 слота, получили %d: %v", len(resp.Slots), resp.Slots)
	}
}

func TestGetAvailableSlots_ClosedDay(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-16", true, "10:00", "19:00") // закрыт

	resp, err := d.svc.GetAvailableSlots("2026-04-16", 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Slots) != 0 {
		t.Errorf("закрытый день должен давать 0 слотов, получили %d", len(resp.Slots))
	}
}

func TestGetAvailableSlots_WithExistingAppointment(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-17", false, "10:00", "13:00")
	// Занято 10:00–11:00
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-17", Time: "10:00", DurationMin: 60, Status: "active",
	}

	resp, err := d.svc.GetAvailableSlots("2026-04-17", 60)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range resp.Slots {
		if s == "10:00" || s == "10:30" {
			t.Errorf("слот %s должен быть занят", s)
		}
	}
	found11 := false
	for _, s := range resp.Slots {
		if s == "11:00" {
			found11 = true
		}
	}
	if !found11 {
		t.Errorf("слот 11:00 должен быть доступен, получили: %v", resp.Slots)
	}
}

func TestGetAvailableSlots_CancelledDoesNotBlock(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-18", false, "10:00", "12:00")
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-18", Time: "10:00", DurationMin: 60, Status: "cancelled",
	}

	resp, err := d.svc.GetAvailableSlots("2026-04-18", 60)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range resp.Slots {
		if s == "10:00" {
			found = true
		}
	}
	if !found {
		t.Errorf("отменённая запись не должна блокировать 10:00, слоты: %v", resp.Slots)
	}
}

func TestGetAvailableSlots_DurationTooLong(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-19", false, "10:00", "10:30")

	// Длительность 60 мин, рабочее окно 30 мин → нет слотов
	resp, err := d.svc.GetAvailableSlots("2026-04-19", 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Slots) != 0 {
		t.Errorf("слишком длинная услуга — ожидаем 0 слотов, получили %d: %v", len(resp.Slots), resp.Slots)
	}
}

func TestGetAvailableSlots_DefaultWorkHours(t *testing.T) {
	// Если дата не задана в dateRepo — используются дефолтные часы 10:00–19:00
	d := newTestService()
	// не добавляем дату в mockDateRepo

	resp, err := d.svc.GetAvailableSlots("2026-04-20", 60)
	if err != nil {
		t.Fatal(err)
	}
	// 10:00–19:00, шаг 30 мин, длительность 60 → с 10:00 до 18:00 = 17 слотов
	if len(resp.Slots) != 17 {
		t.Errorf("ожидаем 17 слотов с дефолтными часами, получили %d: %v", len(resp.Slots), resp.Slots)
	}
}

func TestGetAvailableSlots_ZeroDuration(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-21", false, "10:00", "12:00")

	// DurationMin=0 должен defaultиться к 60
	resp, err := d.svc.GetAvailableSlots("2026-04-21", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Slots) == 0 {
		t.Error("DurationMin=0 должен defaultиться к 60, слотов не должно быть 0")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: GetFinanceSummary
// ─────────────────────────────────────────────────────────────────────────────

func TestGetFinanceSummary_Empty(t *testing.T) {
	d := newTestService()

	s, err := d.svc.GetFinanceSummary("2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalRevenue != 0 || s.TotalTips != 0 || s.TotalRent != 0 || s.Profit != 0 {
		t.Errorf("пустой период должен давать нули: %+v", s)
	}
}

func TestGetFinanceSummary_Calculation(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-01", Status: "completed", Price: 3000, Tips: 500, Rent: 1000,
	}
	d.aptRepo.appointments[2] = &model.Appointment{
		ID: 2, Date: "2026-04-05", Status: "completed", Price: 5000, Tips: 0, Rent: 1500,
	}
	d.aptRepo.appointments[3] = &model.Appointment{
		ID: 3, Date: "2026-04-03", Status: "active", Price: 9999, Tips: 999, Rent: 999,
	}

	s, err := d.svc.GetFinanceSummary("2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalRevenue != 8000 {
		t.Errorf("выручка: ожидаем 8000, получили %d", s.TotalRevenue)
	}
	if s.TotalTips != 500 {
		t.Errorf("чаевые: ожидаем 500, получили %d", s.TotalTips)
	}
	if s.TotalRent != 2500 {
		t.Errorf("аренда: ожидаем 2500, получили %d", s.TotalRent)
	}
	// Profit = 8000 + 500 - 2500 = 6000
	if s.Profit != 6000 {
		t.Errorf("прибыль: ожидаем 6000, получили %d", s.Profit)
	}
}

func TestGetFinanceSummary_OnlyCompleted(t *testing.T) {
	d := newTestService()
	for i, status := range []string{"active", "cancelled", "rescheduled", "late"} {
		d.aptRepo.appointments[uint(i+1)] = &model.Appointment{
			ID: uint(i + 1), Date: "2026-04-10", Status: status, Price: 1000,
		}
	}

	s, err := d.svc.GetFinanceSummary("2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalRevenue != 0 {
		t.Errorf("только completed должны учитываться в финансах, получили %d", s.TotalRevenue)
	}
}

func TestGetFinanceSummary_NegativeRent(t *testing.T) {
	// Краш-тест: отрицательная аренда (логически невозможна, но прилетит из JSON)
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-01", Status: "completed", Price: 3000, Tips: 0, Rent: -500,
	}
	s, err := d.svc.GetFinanceSummary("2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatal(err)
	}
	// Profit = 3000 + 0 - (-500) = 3500 — система не защищает от этого, просто фиксируем поведение
	if s.Profit != 3500 {
		t.Errorf("прибыль с отриц. арендой: ожидаем 3500, получили %d", s.Profit)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: SetLate
// ─────────────────────────────────────────────────────────────────────────────

func TestSetLate_NotFound(t *testing.T) {
	d := newTestService()
	_, err := d.svc.SetLate(999, 15, false)
	if err == nil {
		t.Fatal("несуществующий ID должен возвращать ошибку")
	}
}

func TestSetLate_NoShift(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-20", Time: "10:00", DurationMin: 60, Status: "active",
	}

	apt, err := d.svc.SetLate(1, 15, false)
	if err != nil {
		t.Fatal(err)
	}
	if apt.Status != "late" {
		t.Errorf("статус должен быть late, получили %q", apt.Status)
	}
	if apt.Time != "10:00" {
		t.Errorf("время не должно меняться при shiftTime=false, получили %q", apt.Time)
	}
	if apt.LateMin != 15 {
		t.Errorf("LateMin должен быть 15, получили %d", apt.LateMin)
	}
}

func TestSetLate_WithShift(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-20", Time: "10:00", DurationMin: 60, Status: "active",
	}

	apt, err := d.svc.SetLate(1, 30, true)
	if err != nil {
		t.Fatal(err)
	}
	if apt.Time != "10:30" {
		t.Errorf("время должно сдвинуться на 30 мин → 10:30, получили %q", apt.Time)
	}
}

func TestSetLate_WithShiftConflict(t *testing.T) {
	d := newTestService()
	// ID=1 начинается в 10:00, ID=2 в 11:00
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-20", Time: "10:00", DurationMin: 60, Status: "active",
	}
	d.aptRepo.appointments[2] = &model.Appointment{
		ID: 2, Date: "2026-04-20", Time: "11:00", DurationMin: 60, Status: "active",
	}

	// Опоздание 30 мин → 10:30, конец 11:30 → конфликт с ID=2
	_, err := d.svc.SetLate(1, 30, true)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("ожидаем ConflictError при сдвиге в занятое время, получили: %v", err)
	}
}

func TestSetLate_LargeDelay(t *testing.T) {
	// Краш-тест: опоздание больше чем рабочий день
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-20", Time: "10:00", DurationMin: 60, Status: "active",
	}

	// 600 минут опоздания = 10 часов → время станет 20:00
	apt, err := d.svc.SetLate(1, 600, true)
	if err != nil {
		t.Fatal(err) // нет других записей — конфликта нет
	}
	if apt.Time != "20:00" {
		t.Errorf("при 600 мин задержки ожидаем 20:00, получили %q", apt.Time)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: GetByContact
// ─────────────────────────────────────────────────────────────────────────────

func TestGetByContact_ByTelegram(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Telegram: "@alice", Date: "2026-04-01", Status: "active"}
	d.aptRepo.appointments[2] = &model.Appointment{ID: 2, Telegram: "@bob", Date: "2026-04-02", Status: "active"}

	apts, err := d.svc.GetByContact("@alice", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(apts) != 1 || apts[0].Telegram != "@alice" {
		t.Errorf("ожидаем 1 запись для @alice, получили: %v", apts)
	}
}

func TestGetByContact_ByPhone(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Phone: "+79161234567", Date: "2026-04-01", Status: "active"}

	apts, err := d.svc.GetByContact("", "+79161234567")
	if err != nil {
		t.Fatal(err)
	}
	if len(apts) != 1 {
		t.Errorf("ожидаем 1 запись по телефону, получили: %v", apts)
	}
}

func TestGetByContact_EmptyQuery(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1, Telegram: "@alice"}

	apts, err := d.svc.GetByContact("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(apts) != 0 {
		t.Errorf("пустой запрос должен возвращать пустой список, получили: %v", apts)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ТЕСТЫ: Delete
// ─────────────────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{ID: 1}

	if err := d.svc.Delete(1); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.aptRepo.appointments[1]; ok {
		t.Error("запись должна была удалиться")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// КРАШ-ТЕСТЫ: экстремальные входные данные
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateAppointment_ExtremeLongStrings(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-25", false, "10:00", "19:00")

	longString := strings.Repeat("А", 10000)
	req := model.CreateAppointmentRequest{
		ClientName: longString, Telegram: "@test",
		Service: longString, Date: "2026-04-25", Time: "10:00", DurationMin: 60,
	}
	// Не должен паниковать — либо создаёт, либо возвращает ошибку БД
	_, _ = d.svc.CreateAppointment(req)
}

func TestCreateAppointment_VeryLargeDuration(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-26", false, "10:00", "19:00")

	req := model.CreateAppointmentRequest{
		ClientName: "Тест", Telegram: "@test",
		Service: "Услуга", Date: "2026-04-26", Time: "10:00", DurationMin: 99999,
	}
	// Не паникует
	_, _ = d.svc.CreateAppointment(req)
}

func TestGetAvailableSlots_NegativeDuration(t *testing.T) {
	d := newTestService()
	d.dateRepo.addDate("2026-04-27", false, "10:00", "19:00")

	// Отрицательная длительность → defaultится к 60
	resp, err := d.svc.GetAvailableSlots("2026-04-27", -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Slots) == 0 {
		t.Error("отрицательная длительность должна defaultиться к 60, слоты не пустые")
	}
}

func TestSetLate_ZeroDelay(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-28", Time: "10:00", DurationMin: 60, Status: "active",
	}

	apt, err := d.svc.SetLate(1, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if apt.Time != "10:00" {
		t.Errorf("нулевое опоздание не должно менять время, получили %q", apt.Time)
	}
}

func TestUpdateAppointment_DBError(t *testing.T) {
	d := newTestService()
	d.aptRepo.appointments[1] = &model.Appointment{
		ID: 1, Date: "2026-04-10", Time: "10:00", DurationMin: 60, Status: "active",
	}
	d.aptRepo.updateErr = errors.New("db connection lost")

	_, err := d.svc.UpdateAppointment(1, model.UpdateAppointmentRequest{Status: "cancelled"})
	if err == nil {
		t.Fatal("ошибка БД должна возвращаться")
	}
}
