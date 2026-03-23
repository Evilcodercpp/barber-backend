package service

import (
	"errors"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type WaitlistService struct {
	repo *repository.WaitlistRepository
}

func NewWaitlistService(repo *repository.WaitlistRepository) *WaitlistService {
	return &WaitlistService{repo: repo}
}

func (s *WaitlistService) Create(req model.WaitlistEntry) (*model.WaitlistEntry, error) {
	if req.ClientName == "" || req.Date == "" {
		return nil, errors.New("имя клиента и дата обязательны")
	}
	req.Status = "waiting"
	if err := s.repo.Create(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *WaitlistService) GetAll() ([]model.WaitlistEntry, error) {
	return s.repo.GetAll()
}

func (s *WaitlistService) GetByDate(date string) ([]model.WaitlistEntry, error) {
	return s.repo.GetByDate(date)
}

func (s *WaitlistService) CountWaiting(date string) (int64, error) {
	return s.repo.CountWaiting(date)
}

func (s *WaitlistService) UpdateStatus(id uint, status string) error {
	valid := map[string]bool{"waiting": true, "notified": true, "booked": true, "declined": true}
	if !valid[status] {
		return errors.New("неверный статус")
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *WaitlistService) Delete(id uint) error {
	return s.repo.Delete(id)
}
