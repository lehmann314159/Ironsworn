package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn/internal/model"
)

// BurnMomentum burns momentum to replace the action score with the momentum value.
// Returns the new outcome, or an error if burn is not beneficial.
func BurnMomentum(ch *model.Character, roll *model.ActionRoll) (model.Outcome, error) {
	if ch.Momentum <= 0 {
		return "", fmt.Errorf("momentum must be positive to burn (current: %d)", ch.Momentum)
	}

	newOutcome := model.DetermineOutcome(ch.Momentum, roll.ChallengeDie1, roll.ChallengeDie2)

	// Verify it's actually an improvement
	improved := false
	if roll.Outcome == model.OutcomeMiss || roll.Outcome == model.OutcomeMissMatch {
		if newOutcome.IsHit() {
			improved = true
		}
	}
	if roll.Outcome == model.OutcomeWeakHit {
		if newOutcome.IsStrongHit() {
			improved = true
		}
	}

	if !improved {
		return "", fmt.Errorf("burning momentum would not improve the outcome")
	}

	// Reset momentum
	ch.Momentum = ch.MomentumReset

	return newOutcome, nil
}

// ResetMomentum resets momentum to the character's reset value.
func ResetMomentum(ch *model.Character) {
	ch.Momentum = ch.MomentumReset
}

// NegativeMomentumCancel checks if negative momentum cancels the action die.
// In Ironsworn, if momentum is negative and its absolute value equals the action die,
// the action die is effectively cancelled (counts as 0).
func NegativeMomentumCancel(momentum int, actionDie int) bool {
	return momentum < 0 && (-momentum) == actionDie
}

// AdjustedActionScore returns the action score accounting for negative momentum cancellation.
func AdjustedActionScore(momentum, actionDie, statValue, adds int) int {
	effective := actionDie
	if NegativeMomentumCancel(momentum, actionDie) {
		effective = 0
	}
	score := effective + statValue + adds
	if score > 10 {
		score = 10
	}
	return score
}
