package orchestrator

import (
	"ai-companion-cli-go/internal/storage"
)

// UpdateIntimacy calculates and updates the relationship score
func UpdateIntimacy(repo *storage.Repository, characterID string, scoreBump float64, currentTurn int) {
	state, err := repo.GetRelationshipState(characterID)
	if err != nil || state == nil {
		return // If it doesn't exist, we skip (should be created in EnsureSession)
	}

	state.IntimacyScore += scoreBump
	
	// Level up logic (simplified)
	if state.IntimacyScore >= 100.0 && state.IntimacyLevel < 10 {
		state.IntimacyLevel += 1
		state.IntimacyScore = 0.0 // reset progress for the new level
	} else if state.IntimacyScore >= 100.0 && state.IntimacyLevel == 10 {
		state.IntimacyScore = 100.0 // Cap at max
	}

	// Level down logic
	if state.IntimacyScore < 0.0 && state.IntimacyLevel > 1 {
		state.IntimacyLevel -= 1
		state.IntimacyScore = 99.0
	} else if state.IntimacyScore < 0.0 && state.IntimacyLevel == 1 {
		state.IntimacyScore = 0.0 // Cap at absolute zero
	}

	state.LastUpdatedTurn = currentTurn
	_ = repo.SaveRelationshipState(state)
}
