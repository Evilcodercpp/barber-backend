package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"barber-backend/internal/model"
	"barber-backend/internal/notify"
	"barber-backend/internal/repository"
	"barber-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	aptSvc          *service.AppointmentService
	svcSvc          *service.ServiceService
	dateSvc         *service.AvailableDateService
	clientSvc       *service.ClientService
	supplySvc       *service.SupplyService
	waitlistSvc     *service.WaitlistService
	svcSupplyRepo   *repository.ServiceSupplyRepository
	aptRepo         *repository.AppointmentRepository
	reviewRepo      *repository.ReviewRepository
	profileRepo     *repository.MasterProfileRepository
	educationRepo   *repository.MasterEducationRepository
	portfolioRepo   *repository.MasterPortfolioRepository
	notifier        *notify.Notifier
}

func NewHandler(
	aptSvc *service.AppointmentService,
	svcSvc *service.ServiceService,
	dateSvc *service.AvailableDateService,
	clientSvc *service.ClientService,
	supplySvc *service.SupplyService,
	waitlistSvc *service.WaitlistService,
	svcSupplyRepo *repository.ServiceSupplyRepository,
	aptRepo *repository.AppointmentRepository,
	reviewRepo *repository.ReviewRepository,
	profileRepo *repository.MasterProfileRepository,
	educationRepo *repository.MasterEducationRepository,
	portfolioRepo *repository.MasterPortfolioRepository,
	notifier *notify.Notifier,
) *Handler {
	return &Handler{
		aptSvc:        aptSvc,
		svcSvc:        svcSvc,
		dateSvc:       dateSvc,
		clientSvc:     clientSvc,
		supplySvc:     supplySvc,
		waitlistSvc:   waitlistSvc,
		svcSupplyRepo: svcSupplyRepo,
		aptRepo:       aptRepo,
		reviewRepo:    reviewRepo,
		profileRepo:   profileRepo,
		educationRepo: educationRepo,
		portfolioRepo: portfolioRepo,
		notifier:      notifier,
	}
}

// clientComment — возвращает комментарий мастера о клиенте из базы клиентов
func (h *Handler) clientComment(telegram, phone string) string {
	c := h.clientSvc.FindByContact(telegram, phone)
	if c == nil {
		return ""
	}
	return c.Comment
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api")

	api.POST("/appointments", h.CreateAppointment)
	api.GET("/appointments", h.GetAppointmentsByDate)
	api.GET("/appointments/range", h.GetAppointmentsByRange)
	api.GET("/appointments/slots", h.GetBookedSlots)
	api.GET("/appointments/available-slots", h.GetAvailableSlots)
	api.GET("/appointments/all", h.GetAllAppointments)
	api.GET("/appointments/by-contact", h.GetByContact)
	api.GET("/appointments/unpaid", h.GetUnpaid)
	api.PATCH("/appointments/:id", h.UpdateAppointment)
	api.PATCH("/appointments/:id/late", h.SetLate)
	api.DELETE("/appointments/:id", h.DeleteAppointment)

	api.GET("/finance", h.GetFinance)

	api.GET("/waitlist", h.GetWaitlist)
	api.POST("/waitlist", h.CreateWaitlistEntry)
	api.PATCH("/waitlist/:id", h.UpdateWaitlistStatus)
	api.DELETE("/waitlist/:id", h.DeleteWaitlistEntry)
	api.GET("/waitlist/count", h.GetWaitlistCount)

	api.GET("/services", h.GetServices)
	api.POST("/services", h.CreateService)
	api.PUT("/services/:id", h.UpdateService)
	api.DELETE("/services/:id", h.DeleteService)
	api.GET("/services/:id/supplies", h.GetServiceSupplies)
	api.POST("/services/:id/supplies", h.AddServiceSupply)
	api.PATCH("/services/:id/supplies/:sid", h.UpdateServiceSupply)
	api.DELETE("/services/:id/supplies/:sid", h.DeleteServiceSupply)

	api.POST("/upload", h.UploadFile)

	api.GET("/dates", h.GetAvailableDates)
	api.GET("/dates/range", h.GetAvailableDatesByRange)
	api.GET("/dates/check", h.CheckDateAvailable)
	api.POST("/dates", h.AddAvailableDate)
	api.DELETE("/dates/:date", h.RemoveAvailableDate)
	api.POST("/dates/close", h.CloseDate)
	api.POST("/dates/open", h.OpenDate)
	api.PATCH("/dates/:date", h.UpdateDateHours)

	api.GET("/clients", h.GetClients)
	api.POST("/clients", h.CreateClient)
	api.GET("/clients/:id/card", h.GetClientCard)
	api.PUT("/clients/:id", h.UpdateClient)
	api.DELETE("/clients/:id", h.DeleteClient)

	api.GET("/supplies", h.GetSupplies)
	api.GET("/supplies/search", h.SearchSupplies)
	api.GET("/supplies/inventory", h.GetInventory)
	api.GET("/supplies/:type", h.GetSuppliesByType)
	api.POST("/supplies", h.CreateSupply)
	api.PUT("/supplies/:id", h.UpdateSupply)
	api.DELETE("/supplies/:id", h.DeleteSupply)
	api.POST("/supplies/:id/restock", h.RestockSupply)

	api.POST("/admin/reseed-services", h.ReseedServices)

	// Master profile (public read, master write)
	api.GET("/profile", h.GetProfile)
	api.PUT("/profile", h.UpdateProfile)
	api.GET("/profile/education", h.GetEducation)
	api.POST("/profile/education", h.CreateEducation)
	api.DELETE("/profile/education/:id", h.DeleteEducation)
	api.GET("/profile/portfolio", h.GetPortfolio)
	api.POST("/profile/portfolio", h.CreatePortfolioItem)
	api.PATCH("/profile/portfolio/:id", h.UpdatePortfolioItem)
	api.DELETE("/profile/portfolio/:id", h.DeletePortfolioItem)

	// Reviews
	api.GET("/reviews", h.GetReviews)
	api.POST("/reviews/check", h.CheckReviewEligibility)
	api.POST("/reviews", h.SubmitReview)
	api.GET("/reviews/all", h.GetAllReviews)
	api.PATCH("/reviews/:id", h.ApproveReview)
	api.DELETE("/reviews/:id", h.DeleteReview)
}

// ==================== Appointments ====================

func (h *Handler) CreateAppointment(c echo.Context) error {
	var req model.CreateAppointmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	apt, err := h.aptSvc.CreateAppointment(req)
	if err != nil {
		var ce *service.ConflictError
		if errors.As(err, &ce) {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error":    err.Error(),
				"conflict": ce.ConflictingApt,
			})
		}
		return c.JSON(http.StatusConflict, m(err.Error()))
	}

	// Определяем: новый клиент или постоянный
	allByContact, _ := h.aptSvc.GetByContact(apt.Telegram, apt.Phone)
	isNew := len(allByContact) <= 1
	comment := h.clientComment(apt.Telegram, apt.Phone)
	if apt.Time == "по договорённости" {
		go h.notifier.NotifyIndividualRequest(apt, isNew, comment)
	} else {
		go h.notifier.NotifyNewBooking(apt, isNew, comment)
	}

	return c.JSON(http.StatusCreated, apt)
}

func (h *Handler) GetAppointmentsByDate(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	data, err := h.aptSvc.GetByDate(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAppointmentsByRange(c echo.Context) error {
	start, end := c.QueryParam("start"), c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, m("start и end обязательны"))
	}
	data, err := h.aptSvc.GetByDateRange(start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAvailableSlots(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	duration := 60
	if d, err := strconv.Atoi(c.QueryParam("duration")); err == nil && d > 0 {
		duration = d
	}
	result, err := h.aptSvc.GetAvailableSlots(date, duration)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetBookedSlots(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	slots, err := h.aptSvc.GetBookedSlots(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, slots)
}

func (h *Handler) GetAllAppointments(c echo.Context) error {
	data, err := h.aptSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) UpdateAppointment(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var req model.UpdateAppointmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}

	// Сохраняем состояние до обновления для уведомления
	oldApt, _ := h.aptSvc.GetByID(id)

	apt, err := h.aptSvc.UpdateAppointment(id, req)
	if err != nil {
		var ce *service.ConflictError
		if errors.As(err, &ce) {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error":    err.Error(),
				"conflict": ce.ConflictingApt,
			})
		}
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}

	// Отправляем уведомление при смене статуса или даты/времени
	if oldApt != nil {
		comment := h.clientComment(apt.Telegram, apt.Phone)
		statusChanged := apt.Status != oldApt.Status
		dateOrTimeChanged := apt.Date != oldApt.Date || apt.Time != oldApt.Time

		if statusChanged {
			switch apt.Status {
			case "cancelled":
				go h.notifier.NotifyCancelled(apt, comment)
			case "completed":
				go h.notifier.NotifyCompleted(apt)
			case "rescheduled":
				go h.notifier.NotifyRescheduled(apt, oldApt.Date, oldApt.Time, comment)
			case "late":
				go h.notifier.NotifyLate(apt, apt.LateMin)
			}
		} else if dateOrTimeChanged && apt.Status == "rescheduled" {
			go h.notifier.NotifyRescheduled(apt, oldApt.Date, oldApt.Time, comment)
		}
	}

	return c.JSON(http.StatusOK, apt)
}

func (h *Handler) SetLate(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var body struct {
		LateMinutes int  `json:"late_minutes"`
		ShiftTime   bool `json:"shift_time"`
	}
	if err := c.Bind(&body); err != nil || body.LateMinutes <= 0 {
		return c.JSON(http.StatusBadRequest, m("late_minutes обязателен и должен быть > 0"))
	}
	apt, err := h.aptSvc.SetLate(id, body.LateMinutes, body.ShiftTime)
	if err != nil {
		var ce *service.ConflictError
		if errors.As(err, &ce) {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error":    err.Error(),
				"conflict": ce.ConflictingApt,
			})
		}
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}

	go h.notifier.NotifyLate(apt, body.LateMinutes)

	return c.JSON(http.StatusOK, apt)
}

func (h *Handler) GetByContact(c echo.Context) error {
	telegram := c.QueryParam("telegram")
	phone := c.QueryParam("phone")
	data, err := h.aptSvc.GetByContact(telegram, phone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetFinance(c echo.Context) error {
	start, end := c.QueryParam("start"), c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, m("start и end обязательны"))
	}
	mode := c.QueryParam("mode") // "accrual" (default) or "cash"
	summary, err := h.aptSvc.GetFinanceSummary(start, end, mode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetUnpaid(c echo.Context) error {
	apts, err := h.aptSvc.GetUnpaid()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, apts)
}

func (h *Handler) DeleteAppointment(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.aptSvc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

// ==================== Services ====================

func (h *Handler) ReseedServices(c echo.Context) error {
	if err := h.svcSvc.SeedDefaults(); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "услуги обновлены"})
}

func (h *Handler) GetServices(c echo.Context) error {
	data, err := h.svcSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CreateService(c echo.Context) error {
	var req model.CreateServiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	svc, err := h.svcSvc.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, svc)
}

func (h *Handler) UpdateService(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var svc model.Service
	if err := c.Bind(&svc); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	svc.ID = id
	if err := h.svcSvc.Update(&svc); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, svc)
}

func (h *Handler) DeleteService(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.svcSvc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

// ==================== Service Supplies ====================

func (h *Handler) GetServiceSupplies(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	data, err := h.svcSupplyRepo.GetByServiceID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) AddServiceSupply(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var body struct {
		SupplyID uint    `json:"supply_id"`
		Quantity float64 `json:"quantity"`
	}
	if err := c.Bind(&body); err != nil || body.SupplyID == 0 {
		return c.JSON(http.StatusBadRequest, m("supply_id и quantity обязательны"))
	}
	ss := &model.ServiceSupply{ServiceID: id, SupplyID: body.SupplyID, Quantity: body.Quantity}
	if err := h.svcSupplyRepo.Create(ss); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, ss)
}

func (h *Handler) UpdateServiceSupply(c echo.Context) error {
	sid := parseID(c.Param("sid"))
	if sid == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var body struct {
		Quantity float64 `json:"quantity"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	if err := h.svcSupplyRepo.UpdateQuantity(sid, body.Quantity); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"id": sid, "quantity": body.Quantity})
}

func (h *Handler) DeleteServiceSupply(c echo.Context) error {
	sid := parseID(c.Param("sid"))
	if sid == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.svcSupplyRepo.Delete(sid); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

func (h *Handler) UploadFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, m("Нет файла"))
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка открытия"))
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка чтения файла"))
	}

	imgbbKey := os.Getenv("IMGBB_API_KEY")
	if imgbbKey == "" {
		return h.saveLocalFile(c, data, filepath.Ext(file.Filename))
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("key", imgbbKey)
	_ = writer.WriteField("image", encoded)
	ext := filepath.Ext(file.Filename)
	_ = writer.WriteField("name", fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	writer.Close()

	resp, err := http.Post("https://api.imgbb.com/1/upload", writer.FormDataContentType(), &body)
	if err != nil {
		return h.saveLocalFile(c, data, ext)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Success {
		// ImgBB failed — save locally as fallback
		return h.saveLocalFile(c, data, ext)
	}

	return c.JSON(http.StatusOK, map[string]string{"url": result.Data.URL})
}

func (h *Handler) saveLocalFile(c echo.Context, data []byte, ext string) error {
	dir := "/tmp/uploads"
	_ = os.MkdirAll(dir, 0755)
	fname := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(dir, fname), data, 0644); err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка сохранения файла"))
	}
	baseURL := os.Getenv("APP_BASE_URL")
	return c.JSON(http.StatusOK, map[string]string{"url": baseURL + "/uploads/" + fname})
}

// ==================== Dates ====================

func (h *Handler) GetAvailableDates(c echo.Context) error {
	data, err := h.dateSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAvailableDatesByRange(c echo.Context) error {
	start, end := c.QueryParam("start"), c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, m("start и end обязательны"))
	}
	data, err := h.dateSvc.GetByRange(start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CheckDateAvailable(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	available, err := h.dateSvc.IsAvailable(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]bool{"available": available})
}

func (h *Handler) AddAvailableDate(c echo.Context) error {
	var body struct{ Date string `json:"date"` }
	if err := c.Bind(&body); err != nil || body.Date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	if err := h.dateSvc.Add(body.Date); err != nil {
		return c.JSON(http.StatusConflict, m("Дата уже добавлена"))
	}
	return c.JSON(http.StatusCreated, map[string]string{"date": body.Date})
}

func (h *Handler) RemoveAvailableDate(c echo.Context) error {
	date := c.Param("date")
	if err := h.dateSvc.Remove(date); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Дата удалена"))
}

func (h *Handler) UpdateDateHours(c echo.Context) error {
	date := c.Param("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	var body struct {
		WorkStart string `json:"work_start"`
		WorkEnd   string `json:"work_end"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	if err := h.dateSvc.UpdateHours(date, body.WorkStart, body.WorkEnd); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]string{"date": date, "work_start": body.WorkStart, "work_end": body.WorkEnd})
}

func (h *Handler) CloseDate(c echo.Context) error {
	var body struct{ Date string `json:"date"` }
	if err := c.Bind(&body); err != nil || body.Date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	if err := h.dateSvc.CloseDate(body.Date); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("День закрыт"))
}

func (h *Handler) OpenDate(c echo.Context) error {
	var body struct{ Date string `json:"date"` }
	if err := c.Bind(&body); err != nil || body.Date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	if err := h.dateSvc.OpenDate(body.Date); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("День открыт"))
}

// ==================== Clients ====================

func (h *Handler) GetClients(c echo.Context) error {
	data, err := h.clientSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CreateClient(c echo.Context) error {
	var req model.CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	client, err := h.clientSvc.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, client)
}

func (h *Handler) UpdateClient(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var client model.Client
	if err := c.Bind(&client); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	client.ID = id
	if err := h.clientSvc.Update(&client); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, client)
}

func (h *Handler) DeleteClient(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.clientSvc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

// ==================== Supplies ====================

func (h *Handler) GetSupplies(c echo.Context) error {
	data, err := h.supplySvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetSuppliesByType(c echo.Context) error {
	t := c.Param("type")
	data, err := h.supplySvc.GetByType(t)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CreateSupply(c echo.Context) error {
	var req model.CreateSupplyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	supply, err := h.supplySvc.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, supply)
}

func (h *Handler) UpdateSupply(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var supply model.Supply
	if err := c.Bind(&supply); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	supply.ID = id
	if err := h.supplySvc.Update(&supply); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, supply)
}

func (h *Handler) DeleteSupply(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.supplySvc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

// SearchSupplies — GET /api/supplies/search?q=...
func (h *Handler) SearchSupplies(c echo.Context) error {
	q := c.QueryParam("q")
	data, err := h.supplySvc.Search(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, data)
}

// GetInventory — GET /api/supplies/inventory
func (h *Handler) GetInventory(c echo.Context) error {
	summary, err := h.supplySvc.GetInventory()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, summary)
}

// RestockSupply — POST /api/supplies/:id/restock
// Body: {"quantity": 500}
func (h *Handler) RestockSupply(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var body struct {
		Quantity float64 `json:"quantity"`
	}
	if err := c.Bind(&body); err != nil || body.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, m("quantity обязателен и должен быть > 0"))
	}
	supply, err := h.supplySvc.Restock(id, body.Quantity)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusOK, supply)
}

// ==================== Waitlist ====================

func (h *Handler) GetWaitlist(c echo.Context) error {
	entries, err := h.waitlistSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, entries)
}

func (h *Handler) CreateWaitlistEntry(c echo.Context) error {
	var req model.WaitlistEntry
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	entry, err := h.waitlistSvc.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, entry)
}

func (h *Handler) UpdateWaitlistStatus(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&body); err != nil || body.Status == "" {
		return c.JSON(http.StatusBadRequest, m("status обязателен"))
	}
	if err := h.waitlistSvc.UpdateStatus(id, body.Status); err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"id": id, "status": body.Status})
}

func (h *Handler) DeleteWaitlistEntry(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.waitlistSvc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

// GetWaitlistCount — GET /api/waitlist/count?date=2026-04-01
func (h *Handler) GetWaitlistCount(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, m("date обязателен"))
	}
	count, err := h.waitlistSvc.CountWaiting(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]int64{"count": count})
}

// ==================== Client Card ====================

func (h *Handler) GetClientCard(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	client, err := h.clientSvc.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, m("Клиент не найден"))
	}

	// Получаем историю визитов по контакту
	apts, _ := h.aptSvc.GetByContact(client.Telegram, client.Phone)
	if len(apts) == 0 && client.Name != "" {
		apts, _ = h.aptRepo.GetByClientName(client.Name)
	}

	card := model.ClientCard{Client: *client}
	svcCount := map[string]int{}
	for _, a := range apts {
		if a.Status != "completed" {
			continue
		}
		card.TotalVisits++
		card.TotalSpent += a.Price + a.Tips
		if card.LastVisit == "" || a.Date > card.LastVisit {
			card.LastVisit = a.Date
		}
		svcCount[a.Service]++
	}
	if card.TotalVisits > 0 {
		card.AverageCheck = card.TotalSpent / card.TotalVisits
	}
	for svc, cnt := range svcCount {
		if cnt > svcCount[card.FavoriteService] {
			card.FavoriteService = svc
		}
	}
	return c.JSON(http.StatusOK, card)
}

// ==================== Reviews ====================

// CheckReviewEligibility — проверяет по телефону или Telegram, есть ли завершённые записи, на которые ещё не оставлен отзыв
func (h *Handler) CheckReviewEligibility(c echo.Context) error {
	var body struct {
		Phone    string `json:"phone"`
		Telegram string `json:"telegram"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	if body.Phone == "" && body.Telegram == "" {
		return c.JSON(http.StatusBadRequest, m("телефон или telegram обязателен"))
	}
	seen := map[uint]bool{}
	var apts []model.Appointment
	if body.Phone != "" {
		if byPhone, err := h.reviewRepo.GetCompletedByPhone(body.Phone); err == nil {
			for _, a := range byPhone {
				if !seen[a.ID] {
					seen[a.ID] = true
					apts = append(apts, a)
				}
			}
		}
	}
	if body.Telegram != "" {
		if byTg, err := h.reviewRepo.GetCompletedByTelegram(body.Telegram); err == nil {
			for _, a := range byTg {
				if !seen[a.ID] {
					seen[a.ID] = true
					apts = append(apts, a)
				}
			}
		}
	}
	eligible := []model.EligibleAppointment{}
	for _, a := range apts {
		if !h.reviewRepo.ExistsForAppointment(a.ID) {
			eligible = append(eligible, model.EligibleAppointment{
				ID:      a.ID,
				Service: a.Service,
				Date:    a.Date,
			})
		}
	}
	return c.JSON(http.StatusOK, eligible)
}

// SubmitReview — клиент отправляет отзыв
func (h *Handler) SubmitReview(c echo.Context) error {
	var req model.SubmitReviewRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	if req.Rating < 1 || req.Rating > 5 {
		return c.JSON(http.StatusBadRequest, m("Рейтинг от 1 до 5"))
	}
	if req.AppointmentID == 0 {
		return c.JSON(http.StatusBadRequest, m("appointment_id обязателен"))
	}
	if h.reviewRepo.ExistsForAppointment(req.AppointmentID) {
		return c.JSON(http.StatusConflict, m("Отзыв на эту запись уже оставлен"))
	}
	// проверяем, что запись действительно завершена и принадлежит этому контакту
	apt, err := h.aptRepo.GetByID(req.AppointmentID)
	if err != nil || apt.Status != "completed" {
		return c.JSON(http.StatusForbidden, m("Нет доступа к этой записи"))
	}
	phoneMatch := req.Phone != "" && normalizePhone(apt.Phone) == normalizePhone(req.Phone)
	tgMatch := req.Telegram != "" && normalizeTelegram(apt.Telegram) == normalizeTelegram(req.Telegram)
	if !phoneMatch && !tgMatch {
		return c.JSON(http.StatusForbidden, m("Нет доступа к этой записи"))
	}
	review := &model.Review{
		AppointmentID: req.AppointmentID,
		ServiceName:   apt.Service,
		Rating:        req.Rating,
		Text:          req.Text,
		Photos:        req.Photos,
		ClientName:    apt.ClientName,
		Phone:         req.Phone,
		Approved:      false,
	}
	if err := h.reviewRepo.Create(review); err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка сохранения отзыва"))
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "Отзыв отправлен на модерацию"})
}

// GetReviews — публичный список одобренных отзывов (без личных данных)
func (h *Handler) GetReviews(c echo.Context) error {
	reviews, err := h.reviewRepo.GetApproved()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	public := make([]model.PublicReview, 0, len(reviews))
	for _, r := range reviews {
		public = append(public, model.PublicReview{
			ID:          r.ID,
			ClientName:  r.ClientName,
			ServiceName: r.ServiceName,
			Rating:      r.Rating,
			Text:        r.Text,
			Photos:      r.Photos,
			CreatedAt:   r.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, public)
}

// GetAllReviews — все отзывы для мастера (с личными данными)
func (h *Handler) GetAllReviews(c echo.Context) error {
	reviews, err := h.reviewRepo.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, reviews)
}

// ApproveReview — мастер одобряет или отклоняет отзыв
func (h *Handler) ApproveReview(c echo.Context) error {
	id := parseID(c.Param("id"))
	var body struct {
		Approved bool `json:"approved"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	if err := h.reviewRepo.SetApproved(id, body.Approved); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]bool{"approved": body.Approved})
}

// DeleteReview — мастер удаляет отзыв
func (h *Handler) DeleteReview(c echo.Context) error {
	id := parseID(c.Param("id"))
	if err := h.reviewRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалён"))
}

// Helpers

func m(msg string) map[string]string { return map[string]string{"error": msg} }

func normalizePhone(phone string) string {
	digits := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits += string(c)
		}
	}
	if len(digits) >= 10 {
		return digits[len(digits)-10:]
	}
	return digits
}

func normalizeTelegram(t string) string {
	if len(t) > 0 && t[0] == '@' {
		t = t[1:]
	}
	result := ""
	for _, c := range t {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		}
	}
	return result
}

func parseID(s string) uint {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}

// ==================== Master Profile ====================

func (h *Handler) GetProfile(c echo.Context) error {
	profile, err := h.profileRepo.Get()
	if err != nil {
		// Профиль ещё не создан — возвращаем пустой
		return c.JSON(http.StatusOK, model.MasterProfile{})
	}
	return c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	var req model.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	profile, err := h.profileRepo.Upsert(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, profile)
}

func (h *Handler) GetEducation(c echo.Context) error {
	items, err := h.educationRepo.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	if items == nil {
		items = []model.MasterEducation{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateEducation(c echo.Context) error {
	var req model.CreateEducationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	item, err := h.educationRepo.Create(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) DeleteEducation(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.educationRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}

func (h *Handler) GetPortfolio(c echo.Context) error {
	items, err := h.portfolioRepo.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	if items == nil {
		items = []model.MasterPortfolio{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) CreatePortfolioItem(c echo.Context) error {
	var req model.CreatePortfolioRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	item, err := h.portfolioRepo.Create(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdatePortfolioItem(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	var req model.UpdatePortfolioRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	item, err := h.portfolioRepo.Update(id, req.Caption)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, item)
}

func (h *Handler) DeletePortfolioItem(c echo.Context) error {
	id := parseID(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, m("Неверный ID"))
	}
	if err := h.portfolioRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, m("Удалено"))
}
