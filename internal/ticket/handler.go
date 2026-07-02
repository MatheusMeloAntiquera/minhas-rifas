package ticket

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/matheusantiquera/minhas-rifas/internal/authctx"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /tickets", h.Create)
	mux.HandleFunc("GET /tickets", h.List)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := authctx.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("falha ao decodificar corpo da requisição", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisição inválido"})
		return
	}

	ticket, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrRaffleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao criar ticket", "error", err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := authctx.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	var filters ListFilters
	if raffleID := r.URL.Query().Get("raffle_id"); raffleID != "" {
		parsed, err := strconv.Atoi(raffleID)
		if err != nil {
			h.logger.Error("raffle_id inválido na query string", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "raffle_id inválido"})
			return
		}
		filters.RaffleID = parsed
	}

	tickets, err := h.service.List(r.Context(), userID, filters)
	if err != nil {
		h.logger.Error("falha ao listar tickets", "error", err, "user_id", userID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
		return
	}

	writeJSON(w, http.StatusOK, tickets)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
