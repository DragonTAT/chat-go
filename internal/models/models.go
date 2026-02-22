package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringSlice handles storing array of strings as JSON in SQLite
type StringSlice []string

func (s *StringSlice) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for StringSlice")
	}
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// MapJSON handles storing map[string]interface{} as JSON in SQLite
type MapJSON map[string]interface{}

func (m *MapJSON) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return errors.New("unsupported type for MapJSON")
	}
}

func (m MapJSON) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// CharacterProfile represents the AI companion's configuration and background
type CharacterProfile struct {
	CharacterID          string      `gorm:"primaryKey" json:"character_id"`
	Name                 string      `json:"name"`
	Age                  int         `json:"age"`
	Gender               string      `json:"gender"`
	RelationshipType     string      `json:"relationship_type"`
	SpeechStyle          string      `json:"speech_style"`
	Catchphrase          string      `json:"catchphrase"`
	PersonalityTags      StringSlice `gorm:"type:text" json:"personality_tags"`
	AnchorRef            string      `json:"anchor_ref"`
	ProfileJSON          MapJSON     `gorm:"type:text" json:"profile_json"`
	MBTI                 string      `json:"mbti"`
	ArtStyle             string      `json:"art_style"`
	FamilyBackground     string      `json:"family_background"`
	EducationDetail      string      `json:"education_detail"`
	DatingHistory        string      `json:"dating_history"`
	CharacterBackstory   string      `json:"character_backstory"`
	ReferenceImagePrompt string      `json:"reference_image_prompt"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID   string    `gorm:"index" json:"session_id"`
	CharacterID string    `gorm:"index" json:"character_id"`
	Role        string    `json:"role"` // system, user, assistant
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

// SessionState tracks the high-level conversation state
type SessionState struct {
	SessionID     string    `gorm:"primaryKey" json:"session_id"`
	CharacterID   string    `gorm:"index" json:"character_id"`
	State         string    `json:"state"` // idle, thinking, streaming, etc.
	TurnIndex     int       `json:"turn_index"`
	FallbackFrom  string    `json:"fallback_from"`
	LastErrorCode string    `json:"last_error_code"`
	StartedAt     time.Time `json:"started_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RelationshipState tracks the intimacy and relation metrics between user and AI
type RelationshipState struct {
	CharacterID           string    `gorm:"primaryKey" json:"character_id"`
	IntimacyLevel         int       `json:"intimacy_level"`         // 1-10
	IntimacyScore         float64   `json:"intimacy_score"`         // 0.0-100.0 within the level
	RelationshipNarrative string    `json:"relationship_narrative"` // LLM summary of the bond
	LastUpdatedTurn       int       `json:"last_updated_turn"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// MemoryFact is a distinct key-value piece of knowledge the AI remembers about the user
type MemoryFact struct {
	FactID          string    `gorm:"primaryKey" json:"fact_id"`
	CharacterID     string    `gorm:"index" json:"character_id"`
	FactType        string    `json:"fact_type"`
	FactKey         string    `json:"fact_key"`
	FactValue       string    `json:"fact_value"`
	Confidence      float64   `json:"confidence"`
	SourceMessageID string    `json:"source_message_id"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

// MemorySummary is a batched summary of conversation history
type MemorySummary struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CharacterID    string    `gorm:"index" json:"character_id"`
	Version        int       `json:"version"`
	BatchStartTurn int       `json:"batch_start_turn"`
	BatchEndTurn   int       `json:"batch_end_turn"`
	SummaryText    string    `json:"summary_text"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CharacterEmotionState tracks the transient emotion context
type CharacterEmotionState struct {
	CharacterID    string    `gorm:"primaryKey" json:"character_id"`
	CurrentEmotion string    `json:"current_emotion"`
	EmotionCause   string    `json:"emotion_cause"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ModelProfile defines the LLM settings (config, not DB)
type ModelProfile struct {
	PrimaryModel  string
	FallbackModel string
	TimeoutMs     int
}
