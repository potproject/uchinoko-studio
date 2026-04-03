package data

type MemoryScope string

const (
	MemoryScopeCharacter    MemoryScope = "character"
	MemoryScopeRelationship MemoryScope = "relationship"
)

type MemoryItem struct {
	ID           string      `json:"id"`
	CharacterID  string      `json:"characterId"`
	OwnerID      string      `json:"ownerId"`
	Scope        MemoryScope `json:"scope"`
	Kind         string      `json:"kind"`
	Content      string      `json:"content"`
	KeywordsText string      `json:"keywordsText"`
	Pinned       bool        `json:"pinned"`
	Confidence   float64     `json:"confidence"`
	Salience     float64     `json:"salience"`
	Source       string      `json:"source"`
	UpdatedAt    string      `json:"updatedAt"`
}

type MemoryItemList struct {
	Items []MemoryItem `json:"items"`
}

type SessionSummary struct {
	OwnerID     string `json:"ownerId"`
	CharacterID string `json:"characterId"`
	SessionID   string `json:"sessionId"`
	Summary     string `json:"summary"`
	UpdatedAt   string `json:"updatedAt"`
}
