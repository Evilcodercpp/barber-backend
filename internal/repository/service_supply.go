package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type ServiceSupplyRepository struct {
	db *gorm.DB
}

func NewServiceSupplyRepository(db *gorm.DB) *ServiceSupplyRepository {
	return &ServiceSupplyRepository{db: db}
}

func (r *ServiceSupplyRepository) GetByServiceID(serviceID uint) ([]model.ServiceSupplyWithInfo, error) {
	var result []model.ServiceSupplyWithInfo
	err := r.db.Table("service_supplies ss").
		Select("ss.id, ss.service_id, ss.supply_id, ss.quantity, s.brand as supply_brand, s.name as supply_name, s.type as supply_type").
		Joins("LEFT JOIN supplies s ON s.id = ss.supply_id").
		Where("ss.service_id = ?", serviceID).
		Order("ss.id ASC").
		Scan(&result).Error
	return result, err
}

func (r *ServiceSupplyRepository) GetByServiceIDRaw(serviceID uint) ([]model.ServiceSupply, error) {
	var result []model.ServiceSupply
	err := r.db.Where("service_id = ?", serviceID).Find(&result).Error
	return result, err
}

func (r *ServiceSupplyRepository) Create(ss *model.ServiceSupply) error {
	return r.db.Create(ss).Error
}

func (r *ServiceSupplyRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServiceSupply{}, id).Error
}

func (r *ServiceSupplyRepository) DeleteByServiceID(serviceID uint) error {
	return r.db.Where("service_id = ?", serviceID).Delete(&model.ServiceSupply{}).Error
}

func (r *ServiceSupplyRepository) UpdateQuantity(id uint, quantity int) error {
	return r.db.Model(&model.ServiceSupply{}).Where("id = ?", id).Update("quantity", quantity).Error
}
