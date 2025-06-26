package main

import (
	"fmt"
	"log"

	"github.com/ai-code-gen/ai-ui-generator/internal/config"
	"github.com/ai-code-gen/ai-ui-generator/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Print database configuration for debugging
	fmt.Printf("Database config:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  User: %s\n", cfg.Database.User)
	fmt.Printf("  Password: %s\n", cfg.Database.Password)
	fmt.Printf("  DBName: %s\n", cfg.Database.DBName)
	fmt.Printf("  SSLMode: %s\n", cfg.Database.SSLMode)
	fmt.Printf("  DSN: %s\n", cfg.Database.DSN())

	// Try to connect
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer database.Close(db)

	fmt.Println("✅ Database connection successful!")
}
