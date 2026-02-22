package main

import (
	"fmt"
	"log"
	"os"

	"ai-companion-cli-go/internal/config"
	"ai-companion-cli-go/internal/llm"
	"ai-companion-cli-go/internal/models"
	"ai-companion-cli-go/internal/orchestrator"
	"ai-companion-cli-go/internal/storage"
	"ai-companion-cli-go/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fmt.Println("Starting AI Companion CLI (Go Edition)...")

	// 1. Load config
	appCfg := config.LoadConfig()

	// 2. Initialize Database layer
	db := storage.NewDB(appCfg.DBPath)
	err := db.Initialize()
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	repo := storage.NewRepository(db)

	// 3. Initialize OpenAI Client
	client := llm.NewClient(appCfg.APIKey, appCfg.ModelProfile)

	// 4. Initialize Orchestrator
	orch := orchestrator.NewOrchestrator(repo, client)

	// --- CLI MOCKING BEHAVIOR FOR FIRST RUN ---
	// If no characters exist in DB, create a default one for the user so TUI handles nicely
	chars, _ := repo.ListCharacters()
	var currentUserProfile *models.CharacterProfile

	if len(chars) == 0 {
		currentUserProfile = &models.CharacterProfile{
			CharacterID:      orchestrator.GenerateCharacterID(),
			Name:             "苏晚晴",
			Age:              24,
			Gender:           "女性",
			RelationshipType: "恋人",
			MBTI:             "INFJ",
			PersonalityTags:  []string{"温柔体贴", "知性"},
			Catchphrase:      "我在呢。",
			SpeechStyle:      "温柔自然，像恋人日常聊天",
		}
		repo.CreateCharacter(currentUserProfile)
	} else {
		currentUserProfile = &chars[0]
	}

	// Session management
	session := orch.EnsureSession(currentUserProfile.CharacterID)

	// 5. Build and Run TUI
	// The original POC Python uses Textual or Rich. Here we use Bubbletea model
	model := ui.InitialModel(repo, client, orch, currentUserProfile, session)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
