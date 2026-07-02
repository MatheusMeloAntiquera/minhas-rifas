package raffle

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
	mux.HandleFunc("POST /raffles", h.Create)
	mux.HandleFunc("GET /raffles", h.ListByUser)
	mux.HandleFunc("GET /raffles/{raffle_id}", h.Get)
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

	raffle, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		h.logger.Error("falha ao criar rifa", "error", err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, raffle)
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := authctx.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}

	raffles, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("falha ao listar rifas do usuário", "error", err, "user_id", userID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
		return
	}

	writeJSON(w, http.StatusOK, raffles)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	raffleID, err := strconv.Atoi(r.PathValue("raffle_id"))
	if err != nil {
		h.logger.Error("raffle_id inválido na requisição", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "raffle_id inválido"})
		return
	}

	detail, err := h.service.Get(r.Context(), raffleID)
	if err != nil {
		if errors.Is(err, ErrRaffleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao buscar rifa", "error", err, "raffle_id", raffleID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
