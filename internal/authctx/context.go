// Package authctx guarda o identificador do usuário autenticado (o id do
// provedor externo, ex.: Clerk) no context.Context da requisição. Fica em um
// pacote neutro para que tanto o middleware de autenticação quanto os handlers
// possam usá-lo sem criar ciclos de importação.
package authctx

import "context"

type contextKey string

const userIDContextKey contextKey = "authenticatedUserID"

// WithUserID retorna um contexto derivado contendo o id do usuário autenticado.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext recupera o id do usuário autenticado injetado pelo middleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}
