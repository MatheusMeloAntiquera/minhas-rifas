package ticket

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/matheusantiquera/minhas-rifas/domain"
	"github.com/matheusantiquera/minhas-rifas/internal/raffle"
	"github.com/matheusantiquera/minhas-rifas/internal/user"
)

var (
	ErrUserNotFound   = errors.New("usuário não encontrado")
	ErrRaffleNotFound = errors.New("rifa não encontrada")
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (domain.Ticket, error)
	List(ctx context.Context, userID int, filters ListFilters) ([]domain.Ticket, error)
}

type service struct {
	validate         *validator.Validate
	repository       Repository
	userRepository   user.Repository
	raffleRepository raffle.Repository
	logger           *slog.Logger
}

func NewService(validate *validator.Validate, repository Repository, userRepository user.Repository, raffleRepository raffle.Repository, logger *slog.Logger) Service {
	return &service{
		validate:         validate,
		repository:       repository,
		userRepository:   userRepository,
		raffleRepository: raffleRepository,
		logger:           logger,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (domain.Ticket, error) {
	if err := s.validate.Struct(input); err != nil {
		return domain.Ticket{}, err
	}

	_, err := s.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		s.logger.Error("usuário não encontrado para criação de ticket", "error", err, "user_id", input.UserID)
		return domain.Ticket{}, ErrUserNotFound
	}

	_, err = s.raffleRepository.FindByID(ctx, input.RaffleID)
	if err != nil {
		s.logger.Error("rifa não encontrada para criação de ticket", "error", err, "raffle_id", input.RaffleID)
		return domain.Ticket{}, ErrRaffleNotFound
	}

	now := time.Now()
	ticket := domain.Ticket{
		UserID:    input.UserID,
		RaffleID:  input.RaffleID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repository.Create(ctx, ticket)
}

func (s *service) List(ctx context.Context, userID int, filters ListFilters) ([]domain.Ticket, error) {
	_, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("usuário não encontrado para listagem de tickets", "error", err, "user_id", userID)
		return nil, ErrUserNotFound
	}

	tickets, err := s.repository.List(ctx, userID, filters)
	if err != nil {
		s.logger.Error("falha ao listar tickets", "error", err, "user_id", userID)
		return nil, err
	}

	return tickets, nil
}
