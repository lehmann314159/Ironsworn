package model

// MoveCategory groups related moves.
type MoveCategory string

const (
	CategoryAdventure    MoveCategory = "adventure"
	CategoryRelationship MoveCategory = "relationship"
	CategoryCombat       MoveCategory = "combat"
	CategorySuffer       MoveCategory = "suffer"
	CategoryQuest        MoveCategory = "quest"
	CategoryFate         MoveCategory = "fate"
)

// MoveDefinition describes an Ironsworn move.
type MoveDefinition struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Category    MoveCategory `json:"category"`
	Description string       `json:"description"`
	StrongHit   string       `json:"strong_hit"`
	WeakHit     string       `json:"weak_hit"`
	Miss        string       `json:"miss"`
	Stats       []string     `json:"stats,omitempty"` // Which stats can be used
	IsProgress  bool         `json:"is_progress"`     // Uses progress roll instead of action roll
}

// MoveRequest is the input for executing a move.
type MoveRequest struct {
	MoveID   string `json:"move_id"`
	Stat     string `json:"stat,omitempty"`     // Required for action rolls
	Adds     int    `json:"adds,omitempty"`     // Bonus adds
	TrackID  string `json:"track_id,omitempty"` // For progress moves
	Amount   int    `json:"amount,omitempty"`   // For suffer moves (harm/stress amount)
	Narrative string `json:"narrative,omitempty"` // Player fiction context
}

// MoveResult is the output of executing a move.
type MoveResult struct {
	Move             MoveDefinition     `json:"move"`
	Roll             *ActionRoll        `json:"roll,omitempty"`
	ProgressRoll     *ProgressRollResult `json:"progress_roll,omitempty"`
	Outcome          Outcome            `json:"outcome"`
	MechanicalEffects []string          `json:"mechanical_effects"`
	Summary          string             `json:"summary"`
	CanBurnMomentum  bool               `json:"can_burn_momentum"`
	BurnOutcome      Outcome            `json:"burn_outcome,omitempty"` // What outcome would be if momentum burned
}
