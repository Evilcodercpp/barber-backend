package service

import (
	"errors"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type ServiceService struct {
	repo *repository.ServiceRepository
}

func NewServiceService(repo *repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) Create(req model.CreateServiceRequest) (*model.Service, error) {
	if req.Name == "" || req.Price == "" {
		return nil, errors.New("название и цена обязательны")
	}

	svc := &model.Service{
		Name:     req.Name,
		Duration: req.Duration,
		Price:    req.Price,
		Category: req.Category,
	}

	if err := s.repo.Create(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) GetAll() ([]model.Service, error) {
	return s.repo.GetAll()
}

func (s *ServiceService) Update(svc *model.Service) error {
	return s.repo.Update(svc)
}

func (s *ServiceService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ServiceService) SeedDefaults() error {
	count, err := s.repo.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaults := []model.Service{
		{Name: "Окрашивание корней", Duration: "90 мин", Price: "4 500 ₽", Category: "color", SortOrder: 1},
		{Name: "Классическое окрашивание S/M", Duration: "120 мин", Price: "6 000 ₽", Category: "color", SortOrder: 2},
		{Name: "Классическое окрашивание L", Duration: "150 мин", Price: "7 000 ₽", Category: "color", SortOrder: 3},
		{Name: "Экстра блонд S/M", Duration: "180 мин", Price: "7 000 ₽", Category: "color", SortOrder: 4},
		{Name: "Экстра блонд L", Duration: "210 мин", Price: "8 000 ₽", Category: "color", SortOrder: 5},
		{Name: "Трендовое окрашивание S/M", Duration: "180 мин", Price: "8 500 ₽", Category: "color", SortOrder: 6},
		{Name: "Трендовое окрашивание L", Duration: "210 мин", Price: "10 000 ₽", Category: "color", SortOrder: 7},
		{Name: "Тотальная перезагрузка цвета", Duration: "240 мин", Price: "10 500 ₽", Category: "color", SortOrder: 8},
		{Name: "Air Touch", Duration: "240 мин", Price: "12 500 ₽", Category: "color", SortOrder: 9},
		{Name: "Стрижка с укладкой", Duration: "60 мин", Price: "3 000 ₽", Category: "cut", SortOrder: 10},
		{Name: "Мужская стрижка", Duration: "40 мин", Price: "2 000 ₽", Category: "cut", SortOrder: 11},
		{Name: "Укладка по форме", Duration: "40 мин", Price: "2 300 ₽", Category: "cut", SortOrder: 12},
	}

	for i := range defaults {
		if err := s.repo.Create(&defaults[i]); err != nil {
			return err
		}
	}
	return nil
}
