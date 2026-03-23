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

func computeDerived(supply *model.Supply) {
	if supply.QuantityGrams > 0 && supply.TotalCost > 0 {
		supply.CostPerUnit = supply.TotalCost / supply.QuantityGrams
	}
	supply.LowStock = supply.MinQuantity > 0 && supply.Quantity <= supply.MinQuantity
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
		MinQuantity:   req.MinQuantity,
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
	computeDerived(supply)
	return supply, nil
}

func (s *SupplyService) GetByType(supplyType string) ([]model.Supply, error) {
	supplies, err := s.repo.GetByType(supplyType)
	if err != nil {
		return nil, err
	}
	for i := range supplies {
		computeDerived(&supplies[i])
	}
	return supplies, nil
}

func (s *SupplyService) GetAll() ([]model.Supply, error) {
	supplies, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	for i := range supplies {
		computeDerived(&supplies[i])
	}
	return supplies, nil
}

func (s *SupplyService) Update(supply *model.Supply) error {
	if err := s.repo.Update(supply); err != nil {
		return err
	}
	computeDerived(supply)
	return nil
}

func (s *SupplyService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// Search ищет расходники по строке (бренд, название, цвет).
func (s *SupplyService) Search(q string) ([]model.Supply, error) {
	if q == "" {
		return s.GetAll()
	}
	supplies, err := s.repo.Search(q)
	if err != nil {
		return nil, err
	}
	for i := range supplies {
		computeDerived(&supplies[i])
	}
	return supplies, nil
}

// Restock добавляет qty к остатку расходника. Возвращает обновлённый объект.
func (s *SupplyService) Restock(id uint, qty float64) (*model.Supply, error) {
	if qty <= 0 {
		return nil, errors.New("количество должно быть > 0")
	}
	if err := s.repo.AddQuantity(id, qty); err != nil {
		return nil, err
	}
	supply, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	computeDerived(supply)
	return supply, nil
}

// InventorySummary — итоговые данные по складу.
type InventorySummary struct {
	Items          []model.Supply `json:"items"`
	LowStockCount  int            `json:"low_stock_count"`
	TotalStockValue float64       `json:"total_stock_value"` // руб — сумма (quantity * cost_per_unit) по всем позициям
}

// GetInventory возвращает полную инвентаризационную сводку.
func (s *SupplyService) GetInventory() (*InventorySummary, error) {
	supplies, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	summary := &InventorySummary{Items: supplies}
	for _, sup := range supplies {
		if sup.LowStock {
			summary.LowStockCount++
		}
		summary.TotalStockValue += sup.CostPerUnit * sup.Quantity
	}
	return summary, nil
}
