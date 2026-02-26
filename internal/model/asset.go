package model

// AssetType categorizes assets.
type AssetType string

const (
	AssetTypeCompanion AssetType = "companion"
	AssetTypePath      AssetType = "path"
	AssetTypeCombat    AssetType = "combat_talent"
	AssetTypeRitual    AssetType = "ritual"
)

// AssetAbility represents one of the three abilities on an asset.
type AssetAbility struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// AssetDefinition describes an available asset from the rulebook.
type AssetDefinition struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      AssetType      `json:"type"`
	Abilities []AssetAbility `json:"abilities"`
	// Companion-specific
	CompanionHealth int `json:"companion_health,omitempty"`
}

// CharacterAsset represents an asset owned by a character.
type CharacterAsset struct {
	ID           string `json:"id"`
	CharacterID  string `json:"character_id"`
	AssetID      string `json:"asset_id"`
	Name         string `json:"name"`
	// Which abilities are unlocked (index 0 is always unlocked on acquire)
	Ability1     bool   `json:"ability1"`
	Ability2     bool   `json:"ability2"`
	Ability3     bool   `json:"ability3"`
	// Companion health (if applicable)
	Health       int    `json:"health,omitempty"`
}
