package service

import (
	"errors"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
)

type ClientService struct {
	repo *repository.ClientRepository
}

func NewClientService(repo *repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) Create(req model.CreateClientRequest) (*model.Client, error) {
	if req.Name == "" {
		return nil, errors.New("имя обязательно")
	}
	c := &model.Client{
		Name:         req.Name,
		Telegram:     req.Telegram,
		Phone:        req.Phone,
		Comment:      req.Comment,
		HairType:     req.HairType,
		ColorFormula: req.ColorFormula,
		Allergies:    req.Allergies,
		BirthDate:    req.BirthDate,
		Tags:         req.Tags,
		Source:       req.Source,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ClientService) GetByID(id uint) (*model.Client, error) {
	return s.repo.GetByID(id)
}

func (s *ClientService) GetAll() ([]model.Client, error) {
	return s.repo.GetAll()
}

func (s *ClientService) Update(c *model.Client) error {
	return s.repo.Update(c)
}

func (s *ClientService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// FindByContact — возвращает клиента по telegram или телефону (для уведомлений)
func (s *ClientService) FindByContact(telegram, phone string) *model.Client {
	return s.repo.FindByContact(telegram, phone)
}
