package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"barber-backend/internal/bot"
	"barber-backend/internal/handler"
	"barber-backend/internal/model"
	"barber-backend/internal/notify"
	"barber-backend/internal/repository"
	"barber-backend/internal/service"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_NAME", "barber"),
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	if err := db.AutoMigrate(
		&model.Appointment{},
		&model.Service{},
		&model.AvailableDate{},
		&model.Client{},
		&model.Supply{},
		&model.ServiceSupply{},
		&model.TelegramUser{},
		&model.WaitlistEntry{},
		&model.Review{},
		&model.MasterProfile{},
		&model.MasterEducation{},
		&model.MasterPortfolio{},
	); err != nil {
		log.Fatal("Ошибка миграции:", err)
	}
	log.Println("БД подключена, миграция выполнена")

	// Repositories
	aptRepo := repository.NewAppointmentRepository(db)
	svcRepo := repository.NewServiceRepository(db)
	dateRepo := repository.NewAvailableDateRepository(db)
	clientRepo := repository.NewClientRepository(db)
	supplyRepo := repository.NewSupplyRepository(db)
	svcSupplyRepo := repository.NewServiceSupplyRepository(db)
	tgUserRepo := repository.NewTelegramUserRepository(db)
	waitlistRepo := repository.NewWaitlistRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	profileRepo := repository.NewMasterProfileRepository(db)
	educationRepo := repository.NewMasterEducationRepository(db)
	portfolioRepo := repository.NewMasterPortfolioRepository(db)

	// Services
	aptSvc := service.NewAppointmentService(aptRepo, dateRepo, clientRepo, svcRepo, svcSupplyRepo, supplyRepo)
	svcSvc := service.NewServiceService(svcRepo)
	dateSvc := service.NewAvailableDateService(dateRepo)
	clientSvc := service.NewClientService(clientRepo)
	supplySvc := service.NewSupplyService(supplyRepo)
	waitlistSvc := service.NewWaitlistService(waitlistRepo)

	// Seed
	if err := svcSvc.SeedDefaults(); err != nil {
		log.Println("Ошибка seed:", err)
	}

	// Notifier
	tgNotifier := notify.NewNotifier()
	if tgNotifier.Enabled() {
		log.Println("Telegram уведомления: включены")
	} else {
		log.Println("Telegram уведомления: выключены (TELEGRAM_BOT_TOKEN или TELEGRAM_MASTER_CHAT_ID не заданы)")
	}

	// Telegram bot (long-polling)
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken != "" {
		tgBot := bot.New(botToken, tgUserRepo)
		tgBot.Start(context.Background())
		log.Println("Telegram бот: запущен (long-polling)")
	}

	// Reminder cron — каждый час проверяет записи на завтра
	go func() {
		for {
			sendReminders(aptRepo, tgUserRepo, tgNotifier)
			time.Sleep(1 * time.Hour)
		}
	}()

	// Daily summary cron — каждый день в 19:30 по МСК отправляет сводку на следующий день
	go func() {
		moscowLoc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			moscowLoc = time.FixedZone("MSK", 3*60*60)
		}
		for {
			now := time.Now().In(moscowLoc)
			next := time.Date(now.Year(), now.Month(), now.Day(), 19, 30, 0, 0, moscowLoc)
			if !now.Before(next) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(next.Sub(now))
			sendDailySummary(aptRepo, clientRepo, supplyRepo, svcRepo, svcSupplyRepo, tgNotifier)
		}
	}()

	// Cleanup cron — раз в сутки удаляет просроченные записи листа ожидания
	go func() {
		for {
			today := time.Now().Format("2006-01-02")
			if err := waitlistRepo.DeleteExpired(today); err != nil {
				log.Printf("[waitlist cleanup] ошибка: %v", err)
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	// Handler
	h := handler.NewHandler(aptSvc, svcSvc, dateSvc, clientSvc, supplySvc, waitlistSvc, svcSupplyRepo, aptRepo, reviewRepo, profileRepo, educationRepo, portfolioRepo, tgNotifier)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	h.RegisterRoutes(e)
	e.Static("/uploads", "/tmp/uploads")

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	port := getEnv("PORT", "8080")
	log.Printf("Сервер запущен на :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func sendDailySummary(
	aptRepo *repository.AppointmentRepository,
	clientRepo *repository.ClientRepository,
	supplyRepo *repository.SupplyRepository,
	svcRepo *repository.ServiceRepository,
	svcSupplyRepo *repository.ServiceSupplyRepository,
	notifier *notify.Notifier,
) {
	if !notifier.Enabled() {
		return
	}

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	allApts, err := aptRepo.GetByDate(tomorrow)
	if err != nil {
		log.Printf("[daily-summary] ошибка запроса записей: %v", err)
		return
	}

	// Только активные записи
	var apts []model.Appointment
	for _, a := range allApts {
		if a.Status == "active" || a.Status == "rescheduled" {
			apts = append(apts, a)
		}
	}

	// Все расходники с низким остатком
	globalLowStock, err := supplyRepo.GetLowStock()
	if err != nil {
		log.Printf("[daily-summary] ошибка запроса расходников: %v", err)
	}

	// Индекс расходников с низким остатком: supply_id → Supply
	lowStockByID := make(map[uint]model.Supply)
	for _, s := range globalLowStock {
		lowStockByID[s.ID] = s
	}

	var items []notify.DailySummaryItem
	for _, apt := range apts {
		item := notify.DailySummaryItem{Apt: apt}

		// Карточка клиента
		client := clientRepo.FindByContact(apt.Telegram, apt.Phone)
		if client == nil && apt.ClientName != "" {
			// Попытка найти по имени в прошлых записях не нужна — просто nil
		}
		item.ClientCard = client

		// Прошлые визиты мастера (master_comment из завершённых записей этого клиента)
		var pastApts []model.Appointment
		if apt.Telegram != "" || apt.Phone != "" {
			pastApts, _ = aptRepo.GetByContact(apt.Telegram, apt.Phone)
		}
		for _, pa := range pastApts {
			if pa.ID == apt.ID || pa.Status != "completed" || pa.MasterComment == "" {
				continue
			}
			comment := fmt.Sprintf("%s: %s", pa.Date, pa.MasterComment)
			item.PastComments = append(item.PastComments, comment)
			if len(item.PastComments) >= 3 {
				break
			}
		}

		// Расходники под услугу с низким остатком
		svc, err := svcRepo.GetByName(apt.Service)
		if err == nil && svc != nil {
			svcSupplies, _ := svcSupplyRepo.GetByServiceIDRaw(svc.ID)
			for _, ss := range svcSupplies {
				if s, ok := lowStockByID[ss.SupplyID]; ok {
					item.LowSupplies = append(item.LowSupplies, s)
				}
			}
		}

		items = append(items, item)
	}

	notifier.NotifyDailySummary(tomorrow, items, globalLowStock)
	log.Printf("[daily-summary] отправлена сводка на %s (%d записей)", tomorrow, len(items))
}

func sendReminders(aptRepo *repository.AppointmentRepository, userRepo *repository.TelegramUserRepository, notifier *notify.Notifier) {
	apts, err := aptRepo.GetForReminder()
	if err != nil {
		log.Printf("[reminder] ошибка запроса: %v", err)
		return
	}

	now := time.Now()
	sent := 0
	for _, apt := range apts {
		// пропускаем записи "по договорённости" — у них нет конкретного времени
		if apt.Time == "по договорённости" {
			continue
		}
		// парсим дату+время записи
		aptTime, err := time.ParseInLocation("2006-01-02 15:04", apt.Date+" "+apt.Time, now.Location())
		if err != nil {
			continue
		}
		diff := aptTime.Sub(now)
		// отправляем только если до записи от 20 до 28 часов
		if diff < 20*time.Hour || diff > 28*time.Hour {
			continue
		}

		username := strings.ToLower(strings.TrimPrefix(apt.Telegram, "@"))
		u, err := userRepo.GetByUsername(username)
		if err != nil || u == nil {
			continue // клиент не подписан на бота
		}
		notifier.NotifyClientReminder(u.ChatID, &apt)
		if err := aptRepo.MarkReminderSent(apt.ID); err != nil {
			log.Printf("[reminder] ошибка отметки apt %d: %v", apt.ID, err)
		}
		sent++
	}
	if sent > 0 {
		log.Printf("[reminder] отправлено %d напоминаний", sent)
	}
}
