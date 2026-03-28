package repository

import (
	"barber-backend/internal/model"
	"gorm.io/gorm"
)

type MasterProfileRepository struct {
	db *gorm.DB
}

func NewMasterProfileRepository(db *gorm.DB) *MasterProfileRepository {
	return &MasterProfileRepository{db: db}
}

func (r *MasterProfileRepository) Get() (*model.MasterProfile, error) {
	var p model.MasterProfile
	if err := r.db.First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *MasterProfileRepository) Upsert(req model.UpdateProfileRequest) (*model.MasterProfile, error) {
	var p model.MasterProfile
	r.db.First(&p)
	p.Bio = req.Bio
	p.ExperienceYears = req.ExperienceYears
	p.PhotoURL = req.PhotoURL
	if err := r.db.Save(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// Education

type MasterEducationRepository struct {
	db *gorm.DB
}

func NewMasterEducationRepository(db *gorm.DB) *MasterEducationRepository {
	return &MasterEducationRepository{db: db}
}

func (r *MasterEducationRepository) GetAll() ([]model.MasterEducation, error) {
	var items []model.MasterEducation
	if err := r.db.Order("year desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MasterEducationRepository) Create(req model.CreateEducationRequest) (*model.MasterEducation, error) {
	item := model.MasterEducation{
		Title:    req.Title,
		Year:     req.Year,
		Type:     req.Type,
		ImageURL: req.ImageURL,
	}
	if item.Type == "" {
		item.Type = "education"
	}
	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MasterEducationRepository) Delete(id uint) error {
	return r.db.Delete(&model.MasterEducation{}, id).Error
}

// Portfolio

type MasterPortfolioRepository struct {
	db *gorm.DB
}

func NewMasterPortfolioRepository(db *gorm.DB) *MasterPortfolioRepository {
	return &MasterPortfolioRepository{db: db}
}

func (r *MasterPortfolioRepository) GetAll() ([]model.MasterPortfolio, error) {
	var items []model.MasterPortfolio
	if err := r.db.Order("sort_order asc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MasterPortfolioRepository) Create(req model.CreatePortfolioRequest) (*model.MasterPortfolio, error) {
	item := model.MasterPortfolio{
		PhotoURL:  req.PhotoURL,
		Caption:   req.Caption,
		SortOrder: req.SortOrder,
	}
	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MasterPortfolioRepository) Update(id uint, caption string) (*model.MasterPortfolio, error) {
	var item model.MasterPortfolio
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	item.Caption = caption
	if err := r.db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MasterPortfolioRepository) Delete(id uint) error {
	return r.db.Delete(&model.MasterPortfolio{}, id).Error
}
