package ticket

type CreateInput struct {
	UserID   int `json:"user_id" validate:"required"`
	RaffleID int `json:"raffle_id" validate:"required"`
}

type ListFilters struct {
	RaffleID int `json:"raffle_id"`
}
