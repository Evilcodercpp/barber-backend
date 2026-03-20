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
