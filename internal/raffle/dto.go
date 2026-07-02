package raffle

import (
	"time"

	"github.com/matheusantiquera/minhas-rifas/domain"
)

type CreateInput struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
	ValueTicket float64   `json:"value_ticket" validate:"required,gt=0"`
	DrawDate    time.Time `json:"draw_date" validate:"required"`
}

type GetResponse struct {
	domain.Raffle
	TicketsSold int64 `json:"tickets_sold"`
}
