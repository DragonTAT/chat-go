package storage

import (
	"errors"

	"ai-companion-cli-go/internal/models"
	"gorm.io/gorm"
)

// Repository acts as the data access layer
type Repository struct {
	db *DB
}

// NewRepository initializes a data access repository
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// --- Characters ---

// CreateCharacter saves a new character profile
func (r *Repository) CreateCharacter(character *models.CharacterProfile) error {
	return r.db.Create(character).Error
}

// GetCharacter loads a character profile
func (r *Repository) GetCharacter(characterID string) (*models.CharacterProfile, error) {
	var profile models.CharacterProfile
	err := r.db.First(&profile, "character_id = ?", characterID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &profile, err
}

// ListCharacters loads all created character profiles
func (r *Repository) ListCharacters() ([]models.CharacterProfile, error) {
	var profiles []models.CharacterProfile
	err := r.db.Find(&profiles).Error
	return profiles, err
}

// --- Session State ---

// SaveSessionState upserts session state
func (r *Repository) SaveSessionState(state *models.SessionState) error {
	return r.db.Save(state).Error
}

// GetSessionState loads a session
func (r *Repository) GetSessionState(sessionID string) (*models.SessionState, error) {
	var state models.SessionState
	err := r.db.First(&state, "session_id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &state, err
}

// --- Messages ---

// AppendMessage appends to the conversation history
func (r *Repository) AppendMessage(msg *models.ChatMessage) error {
	return r.db.Create(msg).Error
}

// GetRecentMessages retrieves N recent messages for context
func (r *Repository) GetRecentMessages(characterID string, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.Where("character_id = ?", characterID).Order("timestamp desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// --- Relationship ---

// SaveRelationshipState upserts love metrics
func (r *Repository) SaveRelationshipState(state *models.RelationshipState) error {
	return r.db.Save(state).Error
}

// GetRelationshipState fetches relational state
func (r *Repository) GetRelationshipState(characterID string) (*models.RelationshipState, error) {
	var state models.RelationshipState
	err := r.db.First(&state, "character_id = ?", characterID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &state, err
}

// --- Memory ---

// AppendMemoryFact adds a new fact the AI learned
func (r *Repository) AppendMemoryFact(fact *models.MemoryFact) error {
	return r.db.Create(fact).Error
}

// ListMemoryFactsByCharacter gets all long-term memory facts
func (r *Repository) ListMemoryFactsByCharacter(characterID string) ([]models.MemoryFact, error) {
	var facts []models.MemoryFact
	err := r.db.Where("character_id = ?", characterID).Find(&facts).Error
	return facts, err
}
