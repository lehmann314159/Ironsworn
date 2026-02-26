package model

// Outcome represents the result of a move roll.
type Outcome string

const (
	OutcomeStrongHit        Outcome = "strong_hit"
	OutcomeWeakHit          Outcome = "weak_hit"
	OutcomeMiss             Outcome = "miss"
	OutcomeStrongHitMatch   Outcome = "strong_hit_match"   // Both challenge dice match on strong hit
	OutcomeMissMatch        Outcome = "miss_match"          // Both challenge dice match on miss
)

// IsHit returns true if the outcome is any kind of hit.
func (o Outcome) IsHit() bool {
	return o == OutcomeStrongHit || o == OutcomeStrongHitMatch || o == OutcomeWeakHit
}

// IsStrongHit returns true if the outcome is a strong hit (with or without match).
func (o Outcome) IsStrongHit() bool {
	return o == OutcomeStrongHit || o == OutcomeStrongHitMatch
}

// IsMatch returns true if the challenge dice matched.
func (o Outcome) IsMatch() bool {
	return o == OutcomeStrongHitMatch || o == OutcomeMissMatch
}

// ActionRoll represents a standard action roll: 1d6 + stat + adds vs 2d10.
type ActionRoll struct {
	ActionDie    int     `json:"action_die"`    // 1d6 result
	ChallengeDie1 int   `json:"challenge_die1"` // First d10
	ChallengeDie2 int   `json:"challenge_die2"` // Second d10
	Stat         string  `json:"stat"`          // Which stat was used
	StatValue    int     `json:"stat_value"`    // The stat's value
	Adds         int     `json:"adds"`          // Additional bonuses
	ActionScore  int     `json:"action_score"`  // action_die + stat + adds (capped at 10)
	Outcome      Outcome `json:"outcome"`
}

// ProgressRollResult represents a progress roll: progress score vs 2d10.
type ProgressRollResult struct {
	ProgressScore int     `json:"progress_score"` // 0-10
	ChallengeDie1 int     `json:"challenge_die1"`
	ChallengeDie2 int     `json:"challenge_die2"`
	Outcome       Outcome `json:"outcome"`
}

// OracleRollResult represents a d100 oracle roll.
type OracleRollResult struct {
	Roll       int    `json:"roll"`        // 1-100
	TableID    string `json:"table_id,omitempty"`
	TableName  string `json:"table_name,omitempty"`
	Result     string `json:"result"`
}

// DetermineOutcome calculates the outcome given an action score and two challenge dice.
func DetermineOutcome(actionScore, challenge1, challenge2 int) Outcome {
	match := challenge1 == challenge2
	beats1 := actionScore > challenge1
	beats2 := actionScore > challenge2

	if beats1 && beats2 {
		if match {
			return OutcomeStrongHitMatch
		}
		return OutcomeStrongHit
	}
	if beats1 || beats2 {
		return OutcomeWeakHit
	}
	if match {
		return OutcomeMissMatch
	}
	return OutcomeMiss
}
