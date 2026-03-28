package model

import "time"

type Appointment struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ClientName   string    `json:"client_name" gorm:"not null"`
	Telegram     string    `json:"telegram"`
	Phone        string    `json:"phone"`
	Service      string    `json:"service" gorm:"not null"`
	DurationMin  int       `json:"duration_min" gorm:"default:60"`
	Date         string    `json:"date" gorm:"not null;index"`
	Time         string    `json:"time" gorm:"not null"`
	Status       string    `json:"status" gorm:"default:'active'"`
	Price        int       `json:"price" gorm:"default:0"`
	Tips         int       `json:"tips" gorm:"default:0"`
	Rent         int       `json:"rent" gorm:"default:0"`
	LateMin      int       `json:"late_min" gorm:"default:0"`
	SuppliesUsed  string    `json:"supplies_used" gorm:"type:text"` // JSON: []SupplyUsedItem
	SupplyCost    int       `json:"supply_cost" gorm:"default:0"`   // себестоимость расходников (руб), рассчитывается при завершении
	Comment       string    `json:"comment" gorm:"type:text"`
	MasterComment  string    `json:"master_comment" gorm:"type:text"`   // заметка мастера после визита
	ActualEndTime  string    `json:"actual_end_time"`                   // фактическое время окончания
	ReminderSent   bool      `json:"reminder_sent" gorm:"default:false"`
	PaymentStatus string    `json:"payment_status" gorm:"default:'paid'"` // paid / unpaid / partial
	PaymentDate   string    `json:"payment_date"`                          // дата оплаты (если отличается от date)
	PaidAmount    int       `json:"paid_amount" gorm:"default:0"`          // оплачено по факту (для partial)
	PaymentMethod string    `json:"payment_method" gorm:"default:'cash'"`  // cash / card / transfer
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SupplyUsedItem — одна позиция расходника, использованного в записи.
// Хранится в поле SuppliesUsed как JSON-массив.
type SupplyUsedItem struct {
	SupplyID uint    `json:"supply_id"`
	Quantity float64 `json:"quantity"` // в тех же единицах, что Supply.Unit (граммы или штуки)
}

type CreateAppointmentRequest struct {
	ClientName  string `json:"client_name"`
	Telegram    string `json:"telegram"`
	Phone       string `json:"phone"`
	Service     string `json:"service"`
	DurationMin int    `json:"duration_min"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Price       int    `json:"price"`
	Comment     string `json:"comment"`
}

type UpdateAppointmentRequest struct {
	Date          string `json:"date"`
	Time          string `json:"time"`
	Service       string `json:"service"`
	DurationMin   int    `json:"duration_min"`
	Status        string `json:"status"`
	Price         int    `json:"price"`
	Tips          int    `json:"tips"`
	Rent          int    `json:"rent"`
	LateMin       int    `json:"late_min"`
	SuppliesUsed  string `json:"supplies_used"`
	Comment       string `json:"comment"`
	MasterComment string `json:"master_comment"`
	ActualEndTime string `json:"actual_end_time"`
	PaymentStatus string `json:"payment_status"`
	PaymentDate   string `json:"payment_date"`
	PaidAmount    int    `json:"paid_amount"`
	PaymentMethod string `json:"payment_method"`
}

type FinanceSummary struct {
	Appointments    []Appointment `json:"appointments"`
	TotalRevenue    int           `json:"total_revenue"`
	TotalTips       int           `json:"total_tips"`
	TotalRent       int           `json:"total_rent"`
	TotalSupplyCost int           `json:"total_supply_cost"` // суммарная себестоимость расходников
	Profit          int           `json:"profit"`            // Revenue + Tips - Rent - SupplyCost
}

type Service struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	Duration    string `json:"duration" gorm:"not null"`
	DurationMin int    `json:"duration_min" gorm:"default:60"`
	Price       string `json:"price" gorm:"not null"`
	Category    string `json:"category" gorm:"default:'general'"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
	Description string `json:"description" gorm:"type:text"`
	Photos      string `json:"photos" gorm:"type:text"` // JSON array of URLs: ["url1","url2"]
}

type CreateServiceRequest struct {
	Name        string `json:"name"`
	Duration    string `json:"duration"`
	DurationMin int    `json:"duration_min"`
	Price       string `json:"price"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Photos      string `json:"photos"`
}

// ServiceSupply — шаблон расходников для услуги
type ServiceSupply struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	ServiceID uint    `json:"service_id" gorm:"not null;index"`
	SupplyID  uint    `json:"supply_id" gorm:"not null"`
	Quantity  float64 `json:"quantity" gorm:"type:real;default:0"`
}

type AvailableDate struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Date      string `json:"date" gorm:"not null;uniqueIndex"`
	Closed    bool   `json:"closed" gorm:"default:false"`
	WorkStart string `json:"work_start" gorm:"default:'10:00'"`
	WorkEnd   string `json:"work_end" gorm:"default:'19:00'"`
}

// Client — база клиентов мастера
type Client struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Telegram     string    `json:"telegram"`
	Phone        string    `json:"phone"`
	Comment      string    `json:"comment"`
	HairType     string    `json:"hair_type"`
	ColorFormula string    `json:"color_formula" gorm:"type:text"` // JSON: [{brand,name,shade,percent}]
	Allergies    string    `json:"allergies"`
	BirthDate    string    `json:"birth_date"`
	Tags         string    `json:"tags"`   // JSON array of strings
	Source       string    `json:"source"` // откуда пришёл
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CreateClientRequest struct {
	Name         string `json:"name"`
	Telegram     string `json:"telegram"`
	Phone        string `json:"phone"`
	Comment      string `json:"comment"`
	HairType     string `json:"hair_type"`
	ColorFormula string `json:"color_formula"`
	Allergies    string `json:"allergies"`
	BirthDate    string `json:"birth_date"`
	Tags         string `json:"tags"`
	Source       string `json:"source"`
}

// ClientCard — карточка клиента с вычисляемой статистикой
type ClientCard struct {
	Client
	TotalVisits     int    `json:"total_visits"`
	TotalSpent      int    `json:"total_spent"`
	LastVisit       string `json:"last_visit"`
	AverageCheck    int    `json:"average_check"`
	FavoriteService string `json:"favorite_service"`
}

// WaitlistEntry — запись в листе ожидания (только для мастера)
type WaitlistEntry struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ClientName string    `json:"client_name" gorm:"not null"`
	Telegram   string    `json:"telegram"`
	Phone      string    `json:"phone"`
	Date       string    `json:"date" gorm:"not null;index"`
	Time       string    `json:"time"` // желаемое время, "" = любое
	Service    string    `json:"service"`
	Comment    string    `json:"comment" gorm:"type:text"`
	Status     string    `json:"status" gorm:"default:'waiting'"` // waiting/notified/booked/declined
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Supply — расходники (краски и материалы)
type Supply struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	Type          string  `json:"type" gorm:"not null;index"`          // "paint" или "material"
	Brand         string  `json:"brand" gorm:"not null"`
	Name          string  `json:"name" gorm:"not null"`
	Quantity      float64 `json:"quantity" gorm:"type:real;default:0"` // текущий остаток (г/шт)
	MinQuantity   float64 `json:"min_quantity" gorm:"type:real;default:0"` // порог предупреждения о нехватке
	Unit          string  `json:"unit" gorm:"default:'gram'"`          // "gram" или "piece"
	// PurchaseQty — количество в упаковке при последней закупке (г/шт)
	QuantityGrams float64 `json:"quantity_grams" gorm:"type:real;default:0"`
	TotalCost     float64 `json:"total_cost" gorm:"type:real;default:0"`   // стоимость последней закупки (руб)
	CostPerUnit   float64 `json:"cost_per_unit" gorm:"-"`                  // руб/г или руб/шт (вычисляется)
	LowStock      bool    `json:"low_stock" gorm:"-"`                      // true если остаток ≤ min_quantity (вычисляется)
	Price         string  `json:"price"`                                   // legacy
	Comment       string  `json:"comment"`
	Color         string  `json:"color"`
}

type CreateSupplyRequest struct {
	Type          string  `json:"type"`
	Brand         string  `json:"brand"`
	Name          string  `json:"name"`
	Quantity      float64 `json:"quantity"`
	MinQuantity   float64 `json:"min_quantity"`
	Price         string  `json:"price"`
	Unit          string  `json:"unit"`
	QuantityGrams float64 `json:"quantity_grams"`
	TotalCost     float64 `json:"total_cost"`
	Comment       string  `json:"comment"`
	Color         string  `json:"color"`
}

// ServiceSupplyWithInfo — расходник шаблона с деталями
type ServiceSupplyWithInfo struct {
	ServiceSupply
	SupplyBrand string `json:"supply_brand"`
	SupplyName  string `json:"supply_name"`
	SupplyType  string `json:"supply_type"`
}

// Review — отзыв клиента о завершённой услуге
type Review struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AppointmentID uint      `json:"appointment_id" gorm:"not null;uniqueIndex"`
	ServiceName   string    `json:"service_name"`
	Rating        int       `json:"rating" gorm:"not null"`
	Text          string    `json:"text" gorm:"type:text"`
	Photos        string    `json:"photos" gorm:"type:text"` // JSON array of URLs
	ClientName    string    `json:"client_name"`             // хранится, но не отдаётся публично
	Phone         string    `json:"phone"`                   // хранится, но не отдаётся публично
	Approved      bool      `json:"approved" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// PublicReview — отзыв без личных данных клиента (для клиентского сайта)
type PublicReview struct {
	ID          uint      `json:"id"`
	ClientName  string    `json:"client_name"` // только имя (без контактов)
	ServiceName string    `json:"service_name"`
	Rating      int       `json:"rating"`
	Text        string    `json:"text"`
	Photos      string    `json:"photos"`
	CreatedAt   time.Time `json:"created_at"`
}

type SubmitReviewRequest struct {
	AppointmentID uint   `json:"appointment_id"`
	Phone         string `json:"phone"`
	Telegram      string `json:"telegram"`
	Rating        int    `json:"rating"`
	Text          string `json:"text"`
	Photos        string `json:"photos"` // JSON array
}

// EligibleAppointment — завершённая запись, на которую можно оставить отзыв
type EligibleAppointment struct {
	ID      uint   `json:"id"`
	Service string `json:"service"`
	Date    string `json:"date"`
}

// MasterProfile — профиль мастера (одна строка в таблице)
type MasterProfile struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	Bio             string `json:"bio" gorm:"type:text"`
	ExperienceYears int    `json:"experience_years"`
	PhotoURL        string `json:"photo_url"`
}

// MasterEducation — запись об образовании или сертификате
type MasterEducation struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Title    string `json:"title" gorm:"not null"`
	Year     int    `json:"year"`
	Type     string `json:"type" gorm:"default:'education'"` // education / certificate
	ImageURL string `json:"image_url"`
}

// MasterPortfolio — фото в портфолио
type MasterPortfolio struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	PhotoURL  string `json:"photo_url" gorm:"not null"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
}

type UpdateProfileRequest struct {
	Bio             string `json:"bio"`
	ExperienceYears int    `json:"experience_years"`
	PhotoURL        string `json:"photo_url"`
}

type CreateEducationRequest struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
}

type CreatePortfolioRequest struct {
	PhotoURL  string `json:"photo_url"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order"`
}

type UpdatePortfolioRequest struct {
	Caption string `json:"caption"`
}
