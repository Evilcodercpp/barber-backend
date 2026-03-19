package handler

import (
	"net/http"
	"strconv"

	"barber-backend/internal/model"
	"barber-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	aptSvc  *service.AppointmentService
	svcSvc  *service.ServiceService
	dateSvc *service.AvailableDateService
}

func NewHandler(
	aptSvc *service.AppointmentService,
	svcSvc *service.ServiceService,
	dateSvc *service.AvailableDateService,
) *Handler {
	return &Handler{aptSvc: aptSvc, svcSvc: svcSvc, dateSvc: dateSvc}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api")

	// Appointments
	api.POST("/appointments", h.CreateAppointment)
	api.GET("/appointments", h.GetAppointmentsByDate)
	api.GET("/appointments/range", h.GetAppointmentsByRange)
	api.GET("/appointments/slots", h.GetBookedSlots)
	api.GET("/appointments/all", h.GetAllAppointments)
	api.DELETE("/appointments/:id", h.DeleteAppointment)

	// Services
	api.GET("/services", h.GetServices)
	api.POST("/services", h.CreateService)
	api.PUT("/services/:id", h.UpdateService)
	api.DELETE("/services/:id", h.DeleteService)

	// Available dates
	api.GET("/dates", h.GetAvailableDates)
	api.GET("/dates/range", h.GetAvailableDatesByRange)
	api.GET("/dates/check", h.CheckDateAvailable)
	api.POST("/dates", h.AddAvailableDate)
	api.DELETE("/dates/:date", h.RemoveAvailableDate)
}

// ==================== Appointments ====================

func (h *Handler) CreateAppointment(c echo.Context) error {
	var req model.CreateAppointmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}

	apt, err := h.aptSvc.CreateAppointment(req)
	if err != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, apt)
}

func (h *Handler) GetAppointmentsByDate(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметр date обязателен"})
	}
	data, err := h.aptSvc.GetByDate(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAppointmentsByRange(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметры start и end обязательны"})
	}
	data, err := h.aptSvc.GetByDateRange(start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetBookedSlots(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметр date обязателен"})
	}
	slots, err := h.aptSvc.GetBookedSlots(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, slots)
}

func (h *Handler) GetAllAppointments(c echo.Context) error {
	data, err := h.aptSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) DeleteAppointment(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}
	if err := h.aptSvc.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Удалено"})
}

// ==================== Services ====================

func (h *Handler) GetServices(c echo.Context) error {
	data, err := h.svcSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CreateService(c echo.Context) error {
	var req model.CreateServiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}
	svc, err := h.svcSvc.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, svc)
}

func (h *Handler) UpdateService(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}
	var svc model.Service
	if err := c.Bind(&svc); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}
	svc.ID = uint(id)
	if err := h.svcSvc.Update(&svc); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, svc)
}

func (h *Handler) DeleteService(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}
	if err := h.svcSvc.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Услуга удалена"})
}

// ==================== Available Dates ====================

func (h *Handler) GetAvailableDates(c echo.Context) error {
	data, err := h.dateSvc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetAvailableDatesByRange(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметры start и end обязательны"})
	}
	data, err := h.dateSvc.GetByRange(start, end)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CheckDateAvailable(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметр date обязателен"})
	}
	available, err := h.dateSvc.IsAvailable(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"available": available})
}

func (h *Handler) AddAvailableDate(c echo.Context) error {
	var body struct {
		Date string `json:"date"`
	}
	if err := c.Bind(&body); err != nil || body.Date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Параметр date обязателен"})
	}
	if err := h.dateSvc.Add(body.Date); err != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Дата уже добавлена"})
	}
	return c.JSON(http.StatusCreated, map[string]string{"date": body.Date})
}

func (h *Handler) RemoveAvailableDate(c echo.Context) error {
	date := c.Param("date")
	if err := h.dateSvc.Remove(date); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Дата удалена"})
}
