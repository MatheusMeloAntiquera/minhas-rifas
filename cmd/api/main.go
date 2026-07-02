package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/matheusantiquera/minhas-rifas/config"
	"github.com/matheusantiquera/minhas-rifas/internal/auth"
	"github.com/matheusantiquera/minhas-rifas/internal/raffle"
	"github.com/matheusantiquera/minhas-rifas/internal/ticket"
	"github.com/matheusantiquera/minhas-rifas/pkg/logger"
	"github.com/matheusantiquera/minhas-rifas/pkg/mongodb"
	pkgvalidator "github.com/matheusantiquera/minhas-rifas/pkg/validator"
)

func main() {
	log := logger.New()

	cfg, err := config.New()
	if err != nil {
		log.Error("falha ao carregar configuração", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	mongoClient, err := mongodb.NewConnection(ctx, cfg.MongoURI)
	if err != nil {
		log.Error("falha ao conectar ao MongoDB", "error", err)
		os.Exit(1)
	}

	db := mongodb.GetDatabase(mongoClient, cfg.MongoDatabaseName)
	validate := pkgvalidator.New()

	raffleRepository := raffle.NewRepository(db)
	ticketRepository := ticket.NewRepository(db)

	raffleService := raffle.NewService(validate, raffleRepository, ticketRepository, log)
	raffleHandler := raffle.NewHandler(raffleService, log)

	ticketService := ticket.NewService(validate, ticketRepository, raffleRepository, log)
	ticketHandler := ticket.NewHandler(ticketService, log)

	// Adaptador do provedor de autenticação (Clerk). Trocar de provedor se
	// resume a substituir esta linha por outra implementação de auth.Authenticator.
	authenticator := auth.NewClerkAuthenticator(cfg.ClerkSecretKey)
	authMiddleware := auth.NewMiddleware(authenticator, log)

	// O webhook do Clerk cuida da limpeza em cascata quando um usuário é
	// deletado (user.deleted), removendo suas raffles/tickets.
	clerkWebhookHandler := auth.NewWebhookHandler(raffleService, ticketService, cfg.ClerkWebhookSigningSecret, log)

	// Rotas protegidas: exigem token de sessão válido do Clerk.
	mux := http.NewServeMux()
	raffleHandler.RegisterRoutes(mux)
	ticketHandler.RegisterRoutes(mux)

	// Roteador raiz: o webhook do Clerk é autenticado por assinatura (Svix) e
	// fica fora do middleware; todo o resto passa pela autenticação.
	root := http.NewServeMux()
	clerkWebhookHandler.RegisterRoutes(root)
	root.Handle("/", authMiddleware.Authenticate(mux))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: root,
	}

	go func() {
		log.Info("servidor iniciado", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("falha ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("desligando servidor...")

	if err := server.Shutdown(ctx); err != nil {
		log.Error("falha ao desligar servidor", "error", err)
		os.Exit(1)
	}

	log.Info("servidor desligado com sucesso")
}
