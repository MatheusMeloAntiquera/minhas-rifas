package ticket

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/matheusantiquera/minhas-rifas/domain"
	"github.com/matheusantiquera/minhas-rifas/internal/raffle"
)

var ErrRaffleNotFound = errors.New("rifa não encontrada")

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (domain.Ticket, error)
	List(ctx context.Context, userID string, filters ListFilters) ([]domain.Ticket, error)
	DeleteByUser(ctx context.Context, userID string) (int64, error)
}

type service struct {
	validate         *validator.Validate
	repository       Repository
	raffleRepository raffle.Repository
	logger           *slog.Logger
}

func NewService(validate *validator.Validate, repository Repository, raffleRepository raffle.Repository, logger *slog.Logger) Service {
	return &service{
		validate:         validate,
		repository:       repository,
		raffleRepository: raffleRepository,
		logger:           logger,
	}
}

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (domain.Ticket, error) {
	if err := s.validate.Struct(input); err != nil {
		return domain.Ticket{}, err
	}

	_, err := s.raffleRepository.FindByID(ctx, input.RaffleID)
	if err != nil {
		s.logger.Error("rifa não encontrada para criação de ticket", "error", err, "raffle_id", input.RaffleID)
		return domain.Ticket{}, ErrRaffleNotFound
	}

	now := time.Now()
	ticket := domain.Ticket{
		UserID:    userID,
		RaffleID:  input.RaffleID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repository.Create(ctx, ticket)
}

func (s *service) List(ctx context.Context, userID string, filters ListFilters) ([]domain.Ticket, error) {
	tickets, err := s.repository.List(ctx, userID, filters)
	if err != nil {
		s.logger.Error("falha ao listar tickets", "error", err, "user_id", userID)
		return nil, err
	}

	return tickets, nil
}

func (s *service) DeleteByUser(ctx context.Context, userID string) (int64, error) {
	return s.repository.DeleteByUser(ctx, userID)
}
