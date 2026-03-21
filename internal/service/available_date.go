package service

import (
	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type AvailableDateService struct {
	repo *repository.AvailableDateRepository
}

func NewAvailableDateService(repo *repository.AvailableDateRepository) *AvailableDateService {
	return &AvailableDateService{repo: repo}
}

func (s *AvailableDateService) Add(date string) error {
	return s.repo.Add(date)
}

func (s *AvailableDateService) Remove(date string) error {
	return s.repo.Remove(date)
}

func (s *AvailableDateService) GetAll() ([]model.AvailableDate, error) {
	return s.repo.GetAll()
}

func (s *AvailableDateService) GetByRange(startDate, endDate string) ([]model.AvailableDate, error) {
	return s.repo.GetByRange(startDate, endDate)
}

func (s *AvailableDateService) IsAvailable(date string) (bool, error) {
	return s.repo.IsAvailable(date)
}

func (s *AvailableDateService) CloseDate(date string) error {
	return s.repo.CloseDate(date)
}

func (s *AvailableDateService) OpenDate(date string) error {
	return s.repo.OpenDate(date)
}

func (s *AvailableDateService) UpdateHours(date, workStart, workEnd string) error {
	return s.repo.UpdateHours(date, workStart, workEnd)
}
