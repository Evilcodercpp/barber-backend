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
		Description: req.Description,
		Photos:      req.Photos,
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
	defaults := []model.Service{
		{Name: "Окрашивание корней", Duration: "~90 мин", DurationMin: 90, Price: "4 500 ₽", Category: "color", SortOrder: 1,
			Description: "Точечная работа с отросшими корнями: насыщенный, стойкий цвет без ущерба для длины. Результат — однородный тон от корней до кончиков."},
		{Name: "Окрашивание корней + Блики", Duration: "~210 мин", DurationMin: 210, Price: "6 000 ₽", Category: "color", SortOrder: 2,
			Description: "Коррекция корней в сочетании с точечными бликами — цвет выглядит объёмным и многомерным, как после отдыха на солнце."},
		{Name: "Классическое окрашивание S/M", Duration: "~140 мин", DurationMin: 140, Price: "6 000 ₽", Category: "color", SortOrder: 3,
			Description: "Равномерное окрашивание по всей длине для волос до плеч. Стойкий пигмент, бережная формула, живой блеск."},
		{Name: "Классическое окрашивание L", Duration: "~150 мин", DurationMin: 150, Price: "7 000 ₽", Category: "color", SortOrder: 4,
			Description: "Полное окрашивание для длинных волос. Глубокий, насыщенный цвет с сохранением структуры и мягкости волоса."},
		{Name: "Экстра блонд S/M", Duration: "~180 мин", DurationMin: 180, Price: "7 000 ₽", Category: "color", SortOrder: 5,
			Description: "Максимальное осветление до чистого блонда для волос до плеч. Профессиональные составы с ухаживающими компонентами."},
		{Name: "Экстра блонд L", Duration: "~210 мин", DurationMin: 210, Price: "8 000 ₽", Category: "color", SortOrder: 6,
			Description: "Экстремальное осветление длинных волос с минимальным повреждением. Для тех, кто стремится к лёгкости платинового тона."},
		{Name: "Шатуш", Duration: "~120 мин", DurationMin: 120, Price: "5 000 ₽", Category: "color", SortOrder: 7,
			Description: "Техника ручного окрашивания, создающая плавный переход от тёмных корней к светлым кончикам. Естественно, модно, без резких границ."},
		{Name: "Трендовое окрашивание S/M", Duration: "индивидуально", DurationMin: 180, Price: "от 8 500 ₽", Category: "color", SortOrder: 8,
			Description: "Индивидуальный подход: контрастные акценты, мягкие растяжки и актуальные техники для волос до плеч. Цвет разрабатывается с нуля — под вас."},
		{Name: "Трендовое окрашивание L", Duration: "индивидуально", DurationMin: 210, Price: "от 10 000 ₽", Category: "color", SortOrder: 9,
			Description: "Авторская работа с длинными волосами: многоуровневый цвет, современные техники и неповторимый результат. Каждый оттенок — история."},
		{Name: "Тотальная перезагрузка цвета", Duration: "индивидуально", DurationMin: 240, Price: "от 10 500 ₽", Category: "color", SortOrder: 10,
			Description: "Комплексная трансформация: от смены базы до коррекции оттенка. Идеально, если вы хотите кардинально изменить цвет и начать с чистого листа."},
		{Name: "Индивидуальное окрашивание / Air Touch", Duration: "индивидуально", DurationMin: 240, Price: "от 12 500 ₽", Category: "color", SortOrder: 11,
			Description: "Премиальная техника балаяж + Air Touch: цвет, созданный вручную, с максимально естественным переходом и воздушной текстурой. Штучная работа."},
		{Name: "Стрижка с укладкой", Duration: "~60 мин", DurationMin: 60, Price: "3 000 ₽", Category: "cut", SortOrder: 12,
			Description: "Авторская стрижка с профессиональной укладкой. Форма, которая подчёркивает черты лица и долго сохраняет стиль."},
		{Name: "Мужская стрижка", Duration: "~80 мин", DurationMin: 80, Price: "2 000 ₽", Category: "cut", SortOrder: 13,
			Description: "Чёткие линии, аккуратная форма, финальная укладка. Всё, что нужно для уверенного и опрятного образа."},
		{Name: "Укладка", Duration: "~60 мин", DurationMin: 60, Price: "2 300 ₽", Category: "cut", SortOrder: 14,
			Description: "Профессиональная укладка любой сложности: объёмные волны, гладкая стрижка или естественный стайлинг — выбор за вами."},
		{Name: "Окантовка к любой услуге", Duration: "индивидуально", DurationMin: 30, Price: "1 000 ₽", Category: "cut", SortOrder: 15,
			Description: "Чёткая окантовка шеи и висков в дополнение к любой основной услуге. Финальный штрих, который завершает идеальный образ."},
	}

	for i := range defaults {
		_, err := s.repo.GetByName(defaults[i].Name)
		if err != nil {
			// не найдена — создаём
			if err := s.repo.Create(&defaults[i]); err != nil {
				return err
			}
		}
		// найдена — не трогаем, чтобы не затирать изменения мастера
	}
	return nil
}
