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

	// Services
	aptSvc := service.NewAppointmentService(aptRepo, dateRepo, clientRepo, svcRepo, svcSupplyRepo, supplyRepo)
	svcSvc := service.NewServiceService(svcRepo)
	dateSvc := service.NewAvailableDateService(dateRepo)
	clientSvc := service.NewClientService(clientRepo)
	supplySvc := service.NewSupplyService(supplyRepo)

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

	// Handler
	h := handler.NewHandler(aptSvc, svcSvc, dateSvc, clientSvc, supplySvc, svcSupplyRepo, tgNotifier)

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
