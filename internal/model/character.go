package model

import "fmt"

// Stats represents the five core stats in Ironsworn.
type Stats struct {
	Edge   int `json:"edge"`
	Heart  int `json:"heart"`
	Iron   int `json:"iron"`
	Shadow int `json:"shadow"`
	Wits   int `json:"wits"`
}

// Validate checks that stats are in range (1-3) and sum to 9 (distributed as 3,2,2,1,1).
func (s Stats) Validate() error {
	vals := []int{s.Edge, s.Heart, s.Iron, s.Shadow, s.Wits}
	sum := 0
	for _, v := range vals {
		if v < 1 || v > 3 {
			return fmt.Errorf("each stat must be between 1 and 3, got %d", v)
		}
		sum += v
	}
	if sum != 9 {
		return fmt.Errorf("stats must sum to 9 (distributed as 3,2,2,1,1), got %d", sum)
	}
	return nil
}

// GetStat returns the stat value for the given stat name.
func (s Stats) GetStat(name string) (int, error) {
	switch name {
	case "edge":
		return s.Edge, nil
	case "heart":
		return s.Heart, nil
	case "iron":
		return s.Iron, nil
	case "shadow":
		return s.Shadow, nil
	case "wits":
		return s.Wits, nil
	default:
		return 0, fmt.Errorf("unknown stat: %s", name)
	}
}

// Debilities represents the eight possible debilities.
type Debilities struct {
	// Conditions
	Wounded    bool `json:"wounded"`
	Shaken     bool `json:"shaken"`
	Unprepared bool `json:"unprepared"`
	Encumbered bool `json:"encumbered"`
	// Banes
	Maimed    bool `json:"maimed"`
	Corrupted bool `json:"corrupted"`
	// Burdens
	Cursed    bool `json:"cursed"`
	Tormented bool `json:"tormented"`
}

// Count returns the number of active debilities.
func (d Debilities) Count() int {
	count := 0
	if d.Wounded {
		count++
	}
	if d.Shaken {
		count++
	}
	if d.Unprepared {
		count++
	}
	if d.Encumbered {
		count++
	}
	if d.Maimed {
		count++
	}
	if d.Corrupted {
		count++
	}
	if d.Cursed {
		count++
	}
	if d.Tormented {
		count++
	}
	return count
}

// Character represents a player character in Ironsworn.
type Character struct {
	ID     string `json:"id"`
	GameID string `json:"game_id"`
	Name   string `json:"name"`

	Stats Stats `json:"stats"`

	// Status tracks (0-5)
	Health int `json:"health"`
	Spirit int `json:"spirit"`
	Supply int `json:"supply"`

	// Momentum
	Momentum    int `json:"momentum"`
	MomentumMax int `json:"momentum_max"`
	MomentumReset int `json:"momentum_reset"`

	Debilities Debilities `json:"debilities"`

	// Experience
	ExperienceEarned int `json:"experience_earned"`
	ExperienceSpent  int `json:"experience_spent"`
}

// Clamp constrains a value between min and max.
func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// NewCharacter creates a new character with default values.
func NewCharacter(id, gameID, name string, stats Stats) (*Character, error) {
	if err := stats.Validate(); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("character name cannot be empty")
	}
	return &Character{
		ID:            id,
		GameID:        gameID,
		Name:          name,
		Stats:         stats,
		Health:        5,
		Spirit:        5,
		Supply:        5,
		Momentum:      2,
		MomentumMax:   10,
		MomentumReset: 2,
		Debilities:    Debilities{},
	}, nil
}

// UpdateMomentumLimits recalculates momentum max and reset based on debilities.
func (c *Character) UpdateMomentumLimits() {
	count := c.Debilities.Count()
	c.MomentumMax = 10 - count
	c.MomentumReset = 2 - count
	if c.MomentumReset < 0 {
		c.MomentumReset = 0
	}
	// Clamp current momentum to new max
	if c.Momentum > c.MomentumMax {
		c.Momentum = c.MomentumMax
	}
}

// Snapshot returns a CharacterSnapshot of the current state.
func (c *Character) Snapshot() CharacterSnapshot {
	return CharacterSnapshot{
		Health:   c.Health,
		Spirit:   c.Spirit,
		Supply:   c.Supply,
		Momentum: c.Momentum,
	}
}
