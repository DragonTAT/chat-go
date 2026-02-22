package config

import (
	"log"
	"os"

	"ai-companion-cli-go/internal/models"
	"github.com/joho/godotenv"
)

// AppConfig holds system-level dependencies
type AppConfig struct {
	APIKey       string
	DBPath       string
	ModelProfile models.ModelProfile
}

// LoadConfig reads from .env and Env vars
func LoadConfig() *AppConfig {
	_ = godotenv.Load() // ignore error, might not have .env in prod

	apiKey := os.Getenv("OPENAI_API_KEY")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "companion.db"
	}

	modelConfig := models.ModelProfile{
		PrimaryModel:  "gpt-4o-mini", // Cost-effective default
		FallbackModel: "gpt-3.5-turbo",
		TimeoutMs:     30000,
	}

	// Overrides via env
	if m := os.Getenv("PRIMARY_MODEL"); m != "" {
		modelConfig.PrimaryModel = m
	}

	if apiKey == "" {
		log.Println("WARNING: OPENAI_API_KEY is not set. Chat features will error out until it is configured.")
	}

	return &AppConfig{
		APIKey:       apiKey,
		DBPath:       dbPath,
		ModelProfile: modelConfig,
	}
}
