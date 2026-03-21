package main

import (
	"fmt"
	"log"
	"os"

	"barber-backend/internal/handler"
	"barber-backend/internal/model"
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

	// Handler
	h := handler.NewHandler(aptSvc, svcSvc, dateSvc, clientSvc, supplySvc, svcSupplyRepo)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	h.RegisterRoutes(e)

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
