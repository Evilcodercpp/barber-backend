package handler

import (
	"net/http"
	"strconv"

	"barber-backend/internal/model"
	"barber-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	aptSvc    *service.AppointmentService
	svcSvc    *service.ServiceService
	dateSvc   *service.AvailableDateService
	clientSvc *service.ClientService
	supplySvc *service.SupplyService
}

func NewHandler(
	aptSvc *service.AppointmentService,
	svcSvc *service.ServiceService,
	dateSvc *service.AvailableDateService,
	clientSvc *service.ClientService,
	supplySvc *service.SupplyService,
) *Handler {
	return &Handler{aptSvc: aptSvc, svcSvc: svcSvc, dateSvc: dateSvc, clientSvc: clientSvc, supplySvc: supplySvc}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api")

	api.POST("/appointments", h.CreateAppointment)
	api.GET("/appointments", h.GetAppointmentsByDate)
	api.GET("/appointments/range", h.GetAppointmentsByRange)
	api.GET("/appointments/slots", h.GetBookedSlots)
	api.GET("/appointments/all", h.GetAllAppointments)
	api.DELETE("/appointments/:id", h.DeleteAppointment)

	api.GET("/services", h.GetServices)
	api.POST("/services", h.CreateService)
	api.PUT("/services/:id", h.UpdateService)
	api.DELETE("/services/:id", h.DeleteService)

	api.GET("/dates", h.GetAvailableDates)
	api.GET("/dates/range", h.GetAvailableDatesByRange)
	api.GET("/dates/check", h.CheckDateAvailable)
	api.POST("/dates", h.AddAvailableDate)
	api.DELETE("/dates/:date", h.RemoveAvailableDate)
	api.POST("/dates/close", h.CloseDate)
	api.POST("/dates/open", h.OpenDate)

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
