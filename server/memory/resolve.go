package memory

import (
	"fmt"
	"strings"

	"github.com/potproject/uchinoko-studio/data"
)

func BuildSystemPrompt(character data.CharacterConfig, ownerID string, sessionID string, requestText string, recentMessages []data.ChatCompletionMessage) (string, error) {
	cfg := memoryDefaults(character.Memory)
	if !cfg.Enabled {
		return character.Chat.SystemPrompt, nil
	}

	summary := data.SessionSummary{}
	var err error
	if cfg.EnableSessionSummary {
		summary, err = GetSessionSummary(ownerID, character.General.ID, sessionID)
		if err != nil {
			return "", err
		}
	}

	candidates, err := SearchMemoryCandidates(ownerID, character.General.ID, buildSearchQuery(requestText, recentMessages), cfg)
	if err != nil {
		return "", err
	}

	return composeSystemPrompt(character.Chat.SystemPrompt, candidates, summary, cfg), nil
}

func buildSearchQuery(requestText string, recentMessages []data.ChatCompletionMessage) string {
	parts := []string{requestText}
	for i := len(recentMessages) - 1; i >= 0 && len(parts) < 4; i-- {
		content := strings.TrimSpace(recentMessages[i].Content)
		if content == "" {
			continue
		}
		parts = append(parts, content)
	}
	return strings.Join(parts, "\n")
}

func composeSystemPrompt(base string, candidates []SearchCandidate, summary data.SessionSummary, cfg data.CharacterConfigMemory) string {
	var characterMemories []string
	var relationshipMemories []string
	for _, candidate := range candidates {
		line := fmt.Sprintf("- [%s] %s", candidate.Item.Kind, candidate.Item.Content)
		if candidate.Item.Scope == data.MemoryScopeCharacter {
			characterMemories = append(characterMemories, line)
		} else {
			relationshipMemories = append(relationshipMemories, line)
		}
	}

	sections := []string{strings.TrimSpace(base)}
	sections = append(sections, strings.TrimSpace(`
# Memory Policy
- 以下の memory は会話履歴とは別の補助文脈です。
- 新しい発話が memory と矛盾する場合は、現在の発話を優先してください。
- memory が古い、曖昧、または不確かな場合は断定せず自然に確認してください。
- character memory は人格・世界観の優先文脈として扱ってください。
`))
	if len(characterMemories) > 0 {
		sections = append(sections, "# Character Memory\n"+strings.Join(characterMemories, "\n"))
	}
	if cfg.EnableRelationshipMemory && len(relationshipMemories) > 0 {
		sections = append(sections, "# Relationship Memory\n"+strings.Join(relationshipMemories, "\n"))
	}
	if cfg.EnableSessionSummary && strings.TrimSpace(summary.Summary) != "" {
		sections = append(sections, "# Session Summary\n"+summary.Summary)
	}
	return strings.TrimSpace(strings.Join(sections, "\n\n"))
}
