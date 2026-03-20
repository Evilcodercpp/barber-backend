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
		Name:        req.Name,
		Duration:    req.Duration,
		DurationMin: req.DurationMin,
		Price:       req.Price,
		Category:    req.Category,
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
		{Name: "Окрашивание корней", Duration: "~90 мин", DurationMin: 90, Price: "4 500 ₽", Category: "color", SortOrder: 1},
		{Name: "Окрашивание корней + Блики", Duration: "~210 мин", DurationMin: 210, Price: "6 000 ₽", Category: "color", SortOrder: 2},
		{Name: "Классическое окрашивание S/M", Duration: "~140 мин", DurationMin: 140, Price: "6 000 ₽", Category: "color", SortOrder: 3},
		{Name: "Классическое окрашивание L", Duration: "~150 мин", DurationMin: 150, Price: "7 000 ₽", Category: "color", SortOrder: 4},
		{Name: "Экстра блонд S/M", Duration: "~180 мин", DurationMin: 180, Price: "7 000 ₽", Category: "color", SortOrder: 5},
		{Name: "Экстра блонд L", Duration: "~210 мин", DurationMin: 210, Price: "8 000 ₽", Category: "color", SortOrder: 6},
		{Name: "Шатуш", Duration: "~120 мин", DurationMin: 120, Price: "5 000 ₽", Category: "color", SortOrder: 7},
		{Name: "Трендовое окрашивание S/M", Duration: "индивидуально", DurationMin: 180, Price: "от 8 500 ₽", Category: "color", SortOrder: 8},
		{Name: "Трендовое окрашивание L", Duration: "индивидуально", DurationMin: 210, Price: "от 10 000 ₽", Category: "color", SortOrder: 9},
		{Name: "Тотальная перезагрузка цвета", Duration: "индивидуально", DurationMin: 240, Price: "от 10 500 ₽", Category: "color", SortOrder: 10},
		{Name: "Индивидуальное окрашивание / Air Touch", Duration: "индивидуально", DurationMin: 240, Price: "от 12 500 ₽", Category: "color", SortOrder: 11},
		{Name: "Стрижка с укладкой", Duration: "~60 мин", DurationMin: 60, Price: "3 000 ₽", Category: "cut", SortOrder: 12},
		{Name: "Мужская стрижка", Duration: "~80 мин", DurationMin: 80, Price: "2 000 ₽", Category: "cut", SortOrder: 13},
		{Name: "Укладка", Duration: "~60 мин", DurationMin: 60, Price: "2 300 ₽", Category: "cut", SortOrder: 14},
		{Name: "Окантовка к любой услуге", Duration: "индивидуально", DurationMin: 30, Price: "1 000 ₽", Category: "cut", SortOrder: 15},
	}

	for i := range defaults {
		if err := s.repo.Create(&defaults[i]); err != nil {
			return err
		}
	}
	return nil
}
