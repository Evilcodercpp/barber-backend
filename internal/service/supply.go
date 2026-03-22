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

func computeCostPerUnit(supply *model.Supply) {
	if supply.QuantityGrams > 0 && supply.TotalCost > 0 {
		supply.CostPerUnit = supply.TotalCost / supply.QuantityGrams
	}
}

func (s *SupplyService) Create(req model.CreateSupplyRequest) (*model.Supply, error) {
	if req.Name == "" || req.Brand == "" || req.Type == "" {
		return nil, errors.New("тип, бренд и название обязательны")
	}
	if req.Type != "paint" && req.Type != "material" {
		return nil, errors.New("тип должен быть paint или material")
	}
	unit := req.Unit
	if unit == "" {
		unit = "gram"
	}
	supply := &model.Supply{
		Type:          req.Type,
		Brand:         req.Brand,
		Name:          req.Name,
		Quantity:      req.Quantity,
		Price:         req.Price,
		Unit:          unit,
		QuantityGrams: req.QuantityGrams,
		TotalCost:     req.TotalCost,
		Comment:       req.Comment,
		Color:         req.Color,
	}
	if err := s.repo.Create(supply); err != nil {
		return nil, err
	}
	computeCostPerUnit(supply)
	return supply, nil
}

func (s *SupplyService) GetByType(supplyType string) ([]model.Supply, error) {
	supplies, err := s.repo.GetByType(supplyType)
	if err != nil {
		return nil, err
	}
	for i := range supplies {
		computeCostPerUnit(&supplies[i])
	}
	return supplies, nil
}

func (s *SupplyService) GetAll() ([]model.Supply, error) {
	supplies, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	for i := range supplies {
		computeCostPerUnit(&supplies[i])
	}
	return supplies, nil
}

func (s *SupplyService) Update(supply *model.Supply) error {
	err := s.repo.Update(supply)
	if err != nil {
		return err
	}
	computeCostPerUnit(supply)
	return nil
}

func (s *SupplyService) Delete(id uint) error {
	return s.repo.Delete(id)
}
