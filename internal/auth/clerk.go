package auth

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	svix "github.com/svix/svix-webhooks/go"
)

// Este arquivo concentra tudo que é específico do Clerk (verificação de token e
// webhook). Para trocar de provedor de autenticação, basta substituir este
// adaptador por outro que implemente Authenticator — o middleware e os handlers
// permanecem inalterados.

// ClerkAuthenticator implementa Authenticator usando o SDK do Clerk.
type ClerkAuthenticator struct{}

// NewClerkAuthenticator configura a secret key do Clerk (usada globalmente pelo
// SDK para buscar o JWKS) e retorna o autenticador.
func NewClerkAuthenticator(secretKey string) *ClerkAuthenticator {
	clerk.SetKey(secretKey)
	return &ClerkAuthenticator{}
}

func (a *ClerkAuthenticator) Verify(ctx context.Context, token string) (string, error) {
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token})
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// RaffleCleaner e TicketCleaner são as dependências mínimas do webhook para a
// limpeza em cascata. Declaradas aqui (satisfação estrutural) para não acoplar
// o pacote auth aos pacotes de domínio raffle/ticket.
type RaffleCleaner interface {
	DeleteByUser(ctx context.Context, userID string) (int64, error)
}

type TicketCleaner interface {
	DeleteByUser(ctx context.Context, userID string) (int64, error)
}

// WebhookHandler recebe eventos do Clerk (via Svix). Como os usuários são
// controlados diretamente no Clerk, o único evento relevante é user.deleted,
// que dispara a limpeza em cascata das raffles/tickets do usuário.
type WebhookHandler struct {
	raffles       RaffleCleaner
	tickets       TicketCleaner
	signingSecret string
	logger        *slog.Logger
}

func NewWebhookHandler(raffles RaffleCleaner, tickets TicketCleaner, signingSecret string, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		raffles:       raffles,
		tickets:       tickets,
		signingSecret: signingSecret,
		logger:        logger,
	}
}

func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /webhooks/clerk", h.Handle)
}

// clerkEvent representa o envelope de um evento de webhook do Clerk.
type clerkEvent struct {
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("falha ao ler corpo do webhook", "error", err)
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	wh, err := svix.NewWebhook(h.signingSecret)
	if err != nil {
		h.logger.Error("falha ao inicializar verificador de webhook", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := wh.Verify(payload, r.Header); err != nil {
		h.logger.Warn("assinatura de webhook inválida", "error", err)
		writeError(w, http.StatusUnauthorized, "assinatura inválida")
		return
	}

	var event clerkEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Error("falha ao decodificar evento do webhook", "error", err)
		writeError(w, http.StatusBadRequest, "payload inválido")
		return
	}

	if event.Type == "user.deleted" {
		h.cleanupUser(r.Context(), event.Data.ID)
	} else {
		h.logger.Info("evento de webhook ignorado", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// cleanupUser remove as raffles e tickets pertencentes ao usuário deletado no
// Clerk. Falhas são apenas logadas (idempotência: o webhook não deve reprocessar
// indefinidamente por causa de limpeza parcial).
func (h *WebhookHandler) cleanupUser(ctx context.Context, userID string) {
	if userID == "" {
		return
	}

	if deleted, err := h.raffles.DeleteByUser(ctx, userID); err != nil {
		h.logger.Error("falha ao remover raffles do usuário deletado", "error", err, "user_id", userID)
	} else {
		h.logger.Info("raffles removidas do usuário deletado", "user_id", userID, "count", deleted)
	}

	if deleted, err := h.tickets.DeleteByUser(ctx, userID); err != nil {
		h.logger.Error("falha ao remover tickets do usuário deletado", "error", err, "user_id", userID)
	} else {
		h.logger.Info("tickets removidos do usuário deletado", "user_id", userID, "count", deleted)
	}
}
