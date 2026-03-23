package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type SupplyRepository struct {
	db *gorm.DB
}

func NewSupplyRepository(db *gorm.DB) *SupplyRepository {
	return &SupplyRepository{db: db}
}

func (r *SupplyRepository) Create(s *model.Supply) error {
	return r.db.Create(s).Error
}

func (r *SupplyRepository) GetByType(supplyType string) ([]model.Supply, error) {
	var supplies []model.Supply
	err := r.db.Where("type = ?", supplyType).Order("brand ASC, name ASC").Find(&supplies).Error
	return supplies, err
}

func (r *SupplyRepository) GetAll() ([]model.Supply, error) {
	var supplies []model.Supply
	err := r.db.Order("type ASC, brand ASC, name ASC").Find(&supplies).Error
	return supplies, err
}

func (r *SupplyRepository) Update(s *model.Supply) error {
	return r.db.Save(s).Error
}

func (r *SupplyRepository) Delete(id uint) error {
	return r.db.Delete(&model.Supply{}, id).Error
}

func (r *SupplyRepository) GetByID(id uint) (*model.Supply, error) {
	var s model.Supply
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// DeductQuantity вычитает qty из остатка. Остаток не уходит ниже 0.
func (r *SupplyRepository) DeductQuantity(id uint, qty float64) error {
	return r.db.Model(&model.Supply{}).Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("GREATEST(quantity - ?, 0)", qty)).Error
}

// AddQuantity добавляет qty к остатку (пополнение склада).
func (r *SupplyRepository) AddQuantity(id uint, qty float64) error {
	return r.db.Model(&model.Supply{}).Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("quantity + ?", qty)).Error
}

// Search ищет расходники по названию, бренду или цвету (нечувствительно к регистру).
func (r *SupplyRepository) Search(q string) ([]model.Supply, error) {
	var supplies []model.Supply
	pattern := "%" + q + "%"
	err := r.db.Where("brand ILIKE ? OR name ILIKE ? OR color ILIKE ?", pattern, pattern, pattern).
		Order("type ASC, brand ASC, name ASC").Find(&supplies).Error
	return supplies, err
}
