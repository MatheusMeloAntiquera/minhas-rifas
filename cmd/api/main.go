package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/matheusantiquera/minhas-rifas/config"
	"github.com/matheusantiquera/minhas-rifas/internal/raffle"
	"github.com/matheusantiquera/minhas-rifas/internal/ticket"
	"github.com/matheusantiquera/minhas-rifas/internal/user"
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

	userRepository := user.NewRepository(db)
	userService := user.NewService(validate, userRepository, log)
	userHandler := user.NewHandler(userService, log)

	raffleRepository := raffle.NewRepository(db)
	raffleService := raffle.NewService(validate, raffleRepository, userRepository, log)
	raffleHandler := raffle.NewHandler(raffleService, log)

	ticketRepository := ticket.NewRepository(db)
	ticketService := ticket.NewService(validate, ticketRepository, userRepository, raffleRepository, log)
	ticketHandler := ticket.NewHandler(ticketService, log)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	raffleHandler.RegisterRoutes(mux)
	ticketHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
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
