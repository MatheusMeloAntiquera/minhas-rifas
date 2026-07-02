package auth

import "context"

// Authenticator verifica um token de sessão e retorna o identificador externo
// (subject) do usuário no provedor de autenticação.
//
// É a fronteira entre o app e o provedor: o middleware depende apenas desta
// interface, então trocar de provedor (Clerk, Auth0, JWT próprio, etc.) se
// resume a fornecer outra implementação — sem alterar middleware nem handlers.
type Authenticator interface {
	Verify(ctx context.Context, token string) (subject string, err error)
}
