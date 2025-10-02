package models

import "time"

type Spending struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	Note      string    `json:"note"`
	Date      time.Time `json:"date"`
	ImageURL  string    `json:"image_url"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
