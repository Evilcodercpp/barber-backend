package model

import "time"

type Appointment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ClientName string    `json:"client_name" gorm:"not null"`
	Telegram   string    `json:"telegram" gorm:"not null"`
	Service    string    `json:"service" gorm:"not null"`
	Date       string    `json:"date" gorm:"not null;index"`
	Time       string    `json:"time" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CreateAppointmentRequest struct {
	ClientName string `json:"client_name"`
	Telegram   string `json:"telegram"`
	Service    string `json:"service"`
	Date       string `json:"date"`
	Time       string `json:"time"`
}

type Service struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Duration string `json:"duration" gorm:"not null"`
	Price    string `json:"price" gorm:"not null"`
	Category string `json:"category" gorm:"default:'general'"`
	SortOrder int   `json:"sort_order" gorm:"default:0"`
}

type CreateServiceRequest struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Price    string `json:"price"`
	Category string `json:"category"`
}

type AvailableDate struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Date string `json:"date" gorm:"not null;uniqueIndex"`
}
