package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/matheusantiquera/minhas-rifas/internal/authctx"
)

// Middleware autentica requisições verificando o token de sessão via um
// Authenticator (agnóstico de provedor) e injeta o id do usuário no contexto.
type Middleware struct {
	authenticator Authenticator
	logger        *slog.Logger
}

func NewMiddleware(authenticator Authenticator, logger *slog.Logger) *Middleware {
	return &Middleware{
		authenticator: authenticator,
		logger:        logger,
	}
}

// Authenticate verifica o header Authorization: Bearer <token>, valida o token
// com o provedor e injeta o id do usuário autenticado no contexto da requisição.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "token de autenticação ausente")
			return
		}

		subject, err := m.authenticator.Verify(r.Context(), token)
		if err != nil {
			m.logger.Warn("falha ao verificar token", "error", err)
			writeError(w, http.StatusUnauthorized, "token de autenticação inválido")
			return
		}

		ctx := authctx.WithUserID(r.Context(), subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}

	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}

	return token, true
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
