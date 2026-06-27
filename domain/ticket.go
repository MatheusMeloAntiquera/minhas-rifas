package domain

import "time"

type Ticket struct {
	ID        int       `json:"id" bson:"_id"`
	UserID    int       `json:"user_id" bson:"user_id"`
	RaffleID  int       `json:"raffle_id" bson:"raffle_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
