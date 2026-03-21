package model

import "time"

type Appointment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ClientName  string    `json:"client_name" gorm:"not null"`
	Telegram    string    `json:"telegram"`
	Phone       string    `json:"phone"`
	Service     string    `json:"service" gorm:"not null"`
	DurationMin int       `json:"duration_min" gorm:"default:60"`
	Date        string    `json:"date" gorm:"not null;index"`
	Time        string    `json:"time" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'active'"`
	Price       int       `json:"price" gorm:"default:0"`
	Tips        int       `json:"tips" gorm:"default:0"`
	Rent        int       `json:"rent" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
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
}

type UpdateAppointmentRequest struct {
	Date   string `json:"date"`
	Time   string `json:"time"`
	Status string `json:"status"`
	Price  int    `json:"price"`
	Tips   int    `json:"tips"`
	Rent   int    `json:"rent"`
}

type FinanceSummary struct {
	Appointments []Appointment `json:"appointments"`
	TotalRevenue int           `json:"total_revenue"`
	TotalTips    int           `json:"total_tips"`
	TotalRent    int           `json:"total_rent"`
	Profit       int           `json:"profit"`
}

type Service struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	Duration    string `json:"duration" gorm:"not null"`
	DurationMin int    `json:"duration_min" gorm:"default:60"`
	Price       string `json:"price" gorm:"not null"`
	Category    string `json:"category" gorm:"default:'general'"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
}

type CreateServiceRequest struct {
	Name        string `json:"name"`
	Duration    string `json:"duration"`
	DurationMin int    `json:"duration_min"`
	Price       string `json:"price"`
	Category    string `json:"category"`
}

type AvailableDate struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Date   string `json:"date" gorm:"not null;uniqueIndex"`
	Closed bool   `json:"closed" gorm:"default:false"`
}

// Client — база клиентов мастера
type Client struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Telegram  string    `json:"telegram"`
	Phone     string    `json:"phone"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CreateClientRequest struct {
	Name     string `json:"name"`
	Telegram string `json:"telegram"`
	Phone    string `json:"phone"`
	Comment  string `json:"comment"`
}

// Supply — расходники (краски и материалы)
type Supply struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Type     string `json:"type" gorm:"not null;index"` // "paint" или "material"
	Brand    string `json:"brand" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	Quantity int    `json:"quantity" gorm:"default:0"`
	Price    string `json:"price"`
	Comment  string `json:"comment"`
	Color    string `json:"color"`
}

type CreateSupplyRequest struct {
	Type     string `json:"type"`
	Brand    string `json:"brand"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    string `json:"price"`
	Comment  string `json:"comment"`
	Color    string `json:"color"`
}
