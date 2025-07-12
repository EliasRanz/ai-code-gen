package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/auth"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/database"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/llm"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger := observability.NewLogger("info", "console")

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.NewDBConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", err)
		os.Exit(1)
	}

	// Initialize repositories with proper database connection
	userRepo, err := database.NewPostgreSQLUserRepository(db)
	if err != nil {
		logger.Error("Failed to initialize user repository", err)
		os.Exit(1)
	}

	// Initialize infrastructure services
	tokenProvider := auth.NewJWTTokenProvider(cfg.Auth.JWTSecret, "ai-ui-generator")
	passwordHasher := auth.NewBCryptPasswordHasher()
	llmService := llm.NewOpenAIService(cfg.AI.LLM.APIKey)

	// Create minimal use cases for demonstration
	// Note: This demonstrates that our components can be instantiated and integrated
	logger.Info("Initialized services", map[string]interface{}{
		"userRepo":       userRepo != nil,
		"tokenProvider":  tokenProvider != nil,
		"passwordHasher": passwordHasher != nil,
		"llmService":     llmService != nil,
	})

	logger.Info("Starting server", map[string]interface{}{"port": cfg.Server.Port})

	// Create a simple HTTP server for now
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "ai-ui-generator"}`))
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Server listening", map[string]interface{}{"addr": addr})

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", map[string]interface{}{"addr": srv.Addr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
}

// initUserRepository initializes the user repository with database connection
func initUserRepository(db *gorm.DB) (user.Repository, error) {
	// Initialize the PostgreSQL user repository
	repo, err := database.NewPostgreSQLUserRepository(db)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
