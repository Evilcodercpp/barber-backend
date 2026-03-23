package service

import "barber-backend/internal/model"

// appointmentRepo — минимальный интерфейс для репозитория записей
type appointmentRepo interface {
	Create(apt *model.Appointment) error
	GetByID(id uint) (*model.Appointment, error)
	Update(apt *model.Appointment) error
	GetByDate(date string) ([]model.Appointment, error)
	GetByDateRange(startDate, endDate string) ([]model.Appointment, error)
	GetCompletedByDateRange(startDate, endDate string) ([]model.Appointment, error)
	GetAll() ([]model.Appointment, error)
	GetByContact(telegram, phone string) ([]model.Appointment, error)
	Delete(id uint) error
}

// availableDateRepo — интерфейс для репозитория доступных дат
type availableDateRepo interface {
	IsAvailable(date string) (bool, error)
	GetByDate(date string) (*model.AvailableDate, error)
}

// clientRepo — интерфейс для репозитория клиентов
type clientRepo interface {
	FindOrCreate(name, telegram, phone string)
}

// serviceRepo — интерфейс для репозитория услуг
type serviceRepo interface {
	GetByName(name string) (*model.Service, error)
}

// serviceSupplyRepo — интерфейс для шаблонов расходников услуг
type serviceSupplyRepo interface {
	GetByServiceIDRaw(serviceID uint) ([]model.ServiceSupply, error)
}

// supplyRepo — интерфейс для репозитория расходников
type supplyRepo interface {
	GetByID(id uint) (*model.Supply, error)
	DeductQuantity(id uint, qty float64) error
	AddQuantity(id uint, qty float64) error
}
