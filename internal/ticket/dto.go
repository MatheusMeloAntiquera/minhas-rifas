package ticket

type CreateInput struct {
	RaffleID int `json:"raffle_id" validate:"required"`
}

type ListFilters struct {
	RaffleID int `json:"raffle_id"`
}
