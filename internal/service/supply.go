package service

import (
	"errors"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type SupplyService struct {
	repo *repository.SupplyRepository
}

func NewSupplyService(repo *repository.SupplyRepository) *SupplyService {
	return &SupplyService{repo: repo}
}

func (s *SupplyService) Create(req model.CreateSupplyRequest) (*model.Supply, error) {
	if req.Name == "" || req.Brand == "" || req.Type == "" {
		return nil, errors.New("тип, бренд и название обязательны")
	}
	if req.Type != "paint" && req.Type != "material" {
		return nil, errors.New("тип должен быть paint или material")
	}
	supply := &model.Supply{
		Type:     req.Type,
		Brand:    req.Brand,
		Name:     req.Name,
		Quantity: req.Quantity,
		Price:    req.Price,
		Comment:  req.Comment,
	}
	if err := s.repo.Create(supply); err != nil {
		return nil, err
	}
	return supply, nil
}

func (s *SupplyService) GetByType(supplyType string) ([]model.Supply, error) {
	return s.repo.GetByType(supplyType)
}

func (s *SupplyService) GetAll() ([]model.Supply, error) {
	return s.repo.GetAll()
}

func (s *SupplyService) Update(supply *model.Supply) error {
	return s.repo.Update(supply)
}

func (s *SupplyService) Delete(id uint) error {
	return s.repo.Delete(id)
}
