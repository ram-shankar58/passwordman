package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover
	"github.com/joho/godotenv"

	"vault/internal/config"
	"vault/internal/db"
	"vault/internal/handlers"
	"vault/internal/middleware"
	"vault/internal/repository"
	"vault/internal/services"
)

func main() {
	_=godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	userRepo := repository.NewUserRepository(database)
	vaultRepo := repository.NewVaultRepository(database)

	cryptoSvc, err := services.NewCryptoService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("crypto error: %v", err)
	}

	// Initialize audit service with background worker goroutines
	auditSvc := services.NewAuditService(vaultRepo)
	workerPool := services.NewWorkerPool(cfg.WorkerPoolSize)

	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.TokenTTL)
	vaultSvc := services.NewVaultService(vaultRepo, cryptoSvc, auditSvc)

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	handler := handlers.NewHandler(authSvc, vaultSvc, workerPool)

	app.Get("/health", handlers.Health)

	api := app.Group("/api")
	api.Post("/auth/register", handler.Register)
	api.Post("/auth/login", handler.Login)

	vault := api.Group("/vault", middleware.JWT(cfg.JWTSecret))
	vault.Get("/entries", handler.ListEntries)
	vault.Post("/entries", handler.CreateEntry)
	vault.Get("/entries/:id", handler.GetEntry)
	vault.Put("/entries/:id", handler.UpdateEntry)
	vault.Delete("/entries/:id", handler.DeleteEntry)
	vault.Get("/search", handler.SearchEntries)

	// Graceful shutdown with context
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("shutting down...")

		// Shutdown audit service with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := auditSvc.Shutdown(shutdownCtx); err != nil {
			log.Printf("audit shutdown error: %v", err)
		}

		workerPool.Shutdown()

		if err := app.Shutdown(); err != nil {
			log.Printf("app shutdown error: %v", err)
		}
	}()

	log.Printf("server running on %s", cfg.Port)
	if err := app.Listen(cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
