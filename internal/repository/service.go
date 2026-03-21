package repository

import (
	"barber-backend/internal/model"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(svc *model.Service) error {
	return r.db.Create(svc).Error
}

func (r *ServiceRepository) GetAll() ([]model.Service, error) {
	var services []model.Service
	err := r.db.Order("sort_order ASC, id ASC").Find(&services).Error
	return services, err
}

func (r *ServiceRepository) Update(svc *model.Service) error {
	return r.db.Save(svc).Error
}

func (r *ServiceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Service{}, id).Error
}

func (r *ServiceRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Service{}).Count(&count).Error
	return count, err
}

func (r *ServiceRepository) GetByName(name string) (*model.Service, error) {
	var svc model.Service
	if err := r.db.Where("name = ?", name).First(&svc).Error; err != nil {
		return nil, err
	}
	return &svc, nil
}
