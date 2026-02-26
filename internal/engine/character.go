package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn/internal/model"
)

// CreateCharacter validates and creates a new character.
func CreateCharacter(id, gameID, name string, stats model.Stats) (*model.Character, error) {
	return model.NewCharacter(id, gameID, name, stats)
}

// ValidateStatDistribution checks that the 5 stats are distributed as 3,2,2,1,1.
func ValidateStatDistribution(stats model.Stats) error {
	vals := []int{stats.Edge, stats.Heart, stats.Iron, stats.Shadow, stats.Wits}
	counts := map[int]int{}
	for _, v := range vals {
		counts[v]++
	}
	// Must have: one 3, two 2s, two 1s
	if counts[3] != 1 || counts[2] != 2 || counts[1] != 2 {
		return fmt.Errorf("stats must be distributed as 3,2,2,1,1; got edge=%d heart=%d iron=%d shadow=%d wits=%d",
			stats.Edge, stats.Heart, stats.Iron, stats.Shadow, stats.Wits)
	}
	return nil
}

// SpendExperience spends XP to buy upgrades.
func SpendExperience(ch *model.Character, amount int) error {
	available := ch.ExperienceEarned - ch.ExperienceSpent
	if amount > available {
		return fmt.Errorf("not enough experience: have %d, need %d", available, amount)
	}
	ch.ExperienceSpent += amount
	return nil
}
