package domain

import "time"

type Raffle struct {
	ID          int       `json:"id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	ValueTicket float64   `json:"value_ticket" bson:"value_ticket"`
	UserID      string    `json:"user_id" bson:"user_id"`
	DrawDate    time.Time `json:"draw_date" bson:"draw_date"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}
