package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"barber-backend/internal/model"
	"barber-backend/internal/repository"
	"barber-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	aptSvc        *service.AppointmentService
	svcSvc        *service.ServiceService
	dateSvc       *service.AvailableDateService
	clientSvc     *service.ClientService
	supplySvc     *service.SupplyService
	svcSupplyRepo *repository.ServiceSupplyRepository
}

func NewHandler(
	aptSvc *service.AppointmentService,
	svcSvc *service.ServiceService,
	dateSvc *service.AvailableDateService,
	clientSvc *service.ClientService,
	supplySvc *service.SupplyService,
	svcSupplyRepo *repository.ServiceSupplyRepository,
) *Handler {
	return &Handler{
		aptSvc:        aptSvc,
		svcSvc:        svcSvc,
		dateSvc:       dateSvc,
		clientSvc:     clientSvc,
		supplySvc:     supplySvc,
		svcSupplyRepo: svcSupplyRepo,
	}
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
	api.PATCH("/appointments/:id", h.UpdateAppointment)
	api.DELETE("/appointments/:id", h.DeleteAppointment)

	api.GET("/finance", h.GetFinance)

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
	api.PUT("/clients/:id", h.UpdateClient)
	api.DELETE("/clients/:id", h.DeleteClient)

	api.GET("/supplies", h.GetSupplies)
	api.GET("/supplies/:type", h.GetSuppliesByType)
	api.POST("/supplies", h.CreateSupply)
	api.PUT("/supplies/:id", h.UpdateSupply)
	api.DELETE("/supplies/:id", h.DeleteSupply)
}

// ==================== Appointments ====================

func (h *Handler) CreateAppointment(c echo.Context) error {
	var req model.CreateAppointmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, m("Неверный формат"))
	}
	apt, err := h.aptSvc.CreateAppointment(req)
	if err != nil {
		return c.JSON(http.StatusConflict, m(err.Error()))
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
	apt, err := h.aptSvc.UpdateAppointment(id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m(err.Error()))
	}
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
	summary, err := h.aptSvc.GetFinanceSummary(start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m(err.Error()))
	}
	return c.JSON(http.StatusOK, summary)
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
		SupplyID uint `json:"supply_id"`
		Quantity int  `json:"quantity"`
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
		Quantity int `json:"quantity"`
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

	os.MkdirAll("/tmp/uploads", 0755)
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create("/tmp/uploads/" + filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка сохранения"))
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, m("Ошибка записи"))
	}
	return c.JSON(http.StatusOK, map[string]string{"url": "/uploads/" + filename})
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

// Helpers

func m(msg string) map[string]string { return map[string]string{"error": msg} }

func parseID(s string) uint {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}
