package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-companion-cli-go/internal/llm"
	"ai-companion-cli-go/internal/models"
	"ai-companion-cli-go/internal/storage"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

// Orchestrator ties everything together (DB, LLM, Memory, Intimacy)
type Orchestrator struct {
	repo   *storage.Repository
	client *llm.Client
}

func NewOrchestrator(repo *storage.Repository, client *llm.Client) *Orchestrator {
	return &Orchestrator{
		repo:   repo,
		client: client,
	}
}

// GenerateReplyStream orchestrates fetching history, calculating intimacy, updating UI, and streaming LLM response
func (o *Orchestrator) GenerateReplyStream(
	ctx context.Context,
	userText string,
	profile *models.CharacterProfile,
	session *models.SessionState,
) (<-chan string, <-chan error) {
	// 1. Save User Message
	userMsg := &models.ChatMessage{
		SessionID:   session.SessionID,
		CharacterID: profile.CharacterID,
		Role:        openai.ChatMessageRoleUser,
		Content:     userText,
		Timestamp:   time.Now(),
	}
	_ = o.repo.AppendMessage(userMsg)

	// 2. Fetch Relationship State
	relState, _ := o.repo.GetRelationshipState(profile.CharacterID)
	intimacyLevel := 7 // Default CRUSH equivalent
	if relState != nil && relState.IntimacyLevel > 0 {
		intimacyLevel = relState.IntimacyLevel
	}

	// 3. Simple Interaction: Increment Intimacy (Placeholder for actual memory extraction rules)
	scoreBump := 0.5 // Base bump
	if len(userText) > 20 {
		scoreBump = 1.0 // Effort bump
	}
	UpdateIntimacy(o.repo, profile.CharacterID, scoreBump, session.TurnIndex)

	// 4. Fetch recent history (Last 10 messages for context)
	recentMsgs, _ := o.repo.GetRecentMessages(profile.CharacterID, 10)

	// 5. Build full Prompt
	systemPrompt := BuildSystemPrompt(profile, intimacyLevel)

	openAIMsgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
	}
	for _, m := range recentMsgs {
		role := openai.ChatMessageRoleUser
		if m.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		// Convert memory history to openAI format
		openAIMsgs = append(openAIMsgs, openai.ChatCompletionMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	// 6. Start Streaming
	tokenChan, apiErrChan := o.client.StreamChat(ctx, openAIMsgs, 0.7)

	// 7. Middlewear to save the final assistant answer stream to DB
	// So UI gets tokens, but we also save the complete answer when stream is done
	outTokenChan := make(chan string)
	outErrChan := make(chan error, 1)

	go func() {
		defer close(outTokenChan)
		defer close(outErrChan)

		var completeAnswer strings.Builder
		for err := range apiErrChan {
			// propagate error
			outErrChan <- err
			return
		}

		for chunk := range tokenChan {
			completeAnswer.WriteString(chunk)
			outTokenChan <- chunk
		}

		// Save assistant reply
		assistantMsg := &models.ChatMessage{
			SessionID:   session.SessionID,
			CharacterID: profile.CharacterID,
			Role:        openai.ChatMessageRoleAssistant,
			Content:     completeAnswer.String(),
			Timestamp:   time.Now(),
		}
		_ = o.repo.AppendMessage(assistantMsg)

		// Increment Turn
		session.TurnIndex++
		_ = o.repo.SaveSessionState(session)
	}()

	return outTokenChan, outErrChan
}

// EnsureSession creates a session if not exists
func (o *Orchestrator) EnsureSession(characterID string) *models.SessionState {
	// Find active session
	state, _ := o.repo.GetSessionState("sess_" + characterID)
	if state == nil {
		state = &models.SessionState{
			SessionID:   "sess_" + characterID,
			CharacterID: characterID,
			State:       "idle",
			TurnIndex:   0,
			StartedAt:   time.Now(),
		}
		_ = o.repo.SaveSessionState(state)
	}

	// Ensure relationship state exists as well
	rel, _ := o.repo.GetRelationshipState(characterID)
	if rel == nil {
		rel = &models.RelationshipState{
			CharacterID:     characterID,
			IntimacyLevel:   7,
			IntimacyScore:   50.0,
			LastUpdatedTurn: 0,
		}
		_ = o.repo.SaveRelationshipState(rel)
	}

	return state
}

// GenerateCharacterBackstory offline fallback generating profile
func GenerateCharacterBackstory(seed map[string]string) string {
	return fmt.Sprintf("你是一个名叫%s的%s，生活在%s，今年%s岁。你具有%s的性格特质，说话会带有“%s”的口头禅。",
		seed["name"], seed["occupation"], seed["city"], seed["age"], seed["personality_tags"], seed["catchphrase"])
}

// GenerateCharacterID helper
func GenerateCharacterID() string {
	return "chr_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12]
}
