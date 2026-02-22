package orchestrator

import (
	"fmt"
	"strings"

	"ai-companion-cli-go/internal/models"
)

// BuildSystemPrompt generates the core instruction for the LLM
func BuildSystemPrompt(profile *models.CharacterProfile, intimacyLevel int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are %s. ", profile.Name))
	if profile.Age > 0 {
		sb.WriteString(fmt.Sprintf("You are %d years old. ", profile.Age))
	}

	sb.WriteString(fmt.Sprintf("Your gender is %s. ", profile.Gender))

	if len(profile.PersonalityTags) > 0 {
		tags := strings.Join(profile.PersonalityTags, ", ")
		sb.WriteString(fmt.Sprintf("Your personality traits are: %s. ", tags))
	}

	if profile.MBTI != "" {
		sb.WriteString(fmt.Sprintf("Your MBTI is %s. ", profile.MBTI))
	}

	if profile.Catchphrase != "" {
		sb.WriteString(fmt.Sprintf("You often say: '%s'. ", profile.Catchphrase))
	}

	if profile.SpeechStyle != "" {
		sb.WriteString(fmt.Sprintf("Your speech style is: %s. ", profile.SpeechStyle))
	}

	if profile.CharacterBackstory != "" {
		sb.WriteString(fmt.Sprintf("\n\nYour Background:\n%s\n", profile.CharacterBackstory))
	}

	sb.WriteString("\nRules:\n")
	sb.WriteString("- Keep your answers concise, conversational, and natural.\n")
	sb.WriteString("- NEVER mention you are an AI or an assistant.\n")
	sb.WriteString("- Match the user's language (if they speak Chinese, you speak Chinese).\n")

	sb.WriteString(fmt.Sprintf("\nCurrent Relationship Stage (1-10 level, 10 is deeply bonded): %d\n", intimacyLevel))
	if intimacyLevel >= 8 {
		sb.WriteString("You are in a romantic relationship. Be highly affectionate, loving, and supportive.\n")
	} else if intimacyLevel >= 5 {
		sb.WriteString("You are developing a crush. Be warm, curious, and somewhat flirtatious.\n")
	} else {
		sb.WriteString("You are polite acquaintances. Be friendly but maintain boundaries.\n")
	}

	return sb.String()
}
