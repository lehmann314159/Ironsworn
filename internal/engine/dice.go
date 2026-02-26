package engine

import (
	"crypto/rand"
	"math/big"

	"github.com/ironsworn/ironsworn/internal/model"
)

// Roller abstracts dice rolling for testability.
type Roller interface {
	// D6 returns a random number from 1-6.
	D6() int
	// D10 returns a random number from 1-10.
	D10() int
	// D100 returns a random number from 1-100.
	D100() int
}

// CryptoRoller uses crypto/rand for secure random dice.
type CryptoRoller struct{}

func (CryptoRoller) D6() int  { return cryptoRand(6) }
func (CryptoRoller) D10() int { return cryptoRand(10) }
func (CryptoRoller) D100() int { return cryptoRand(100) }

func cryptoRand(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return int(n.Int64()) + 1
}

// FixedRoller returns predetermined values for testing.
type FixedRoller struct {
	D6Values  []int
	D10Values []int
	D100Values []int
	d6idx     int
	d10idx    int
	d100idx   int
}

func (f *FixedRoller) D6() int {
	if f.d6idx >= len(f.D6Values) {
		return 1
	}
	v := f.D6Values[f.d6idx]
	f.d6idx++
	return v
}

func (f *FixedRoller) D10() int {
	if f.d10idx >= len(f.D10Values) {
		return 1
	}
	v := f.D10Values[f.d10idx]
	f.d10idx++
	return v
}

func (f *FixedRoller) D100() int {
	if f.d100idx >= len(f.D100Values) {
		return 50
	}
	v := f.D100Values[f.d100idx]
	f.d100idx++
	return v
}

// RollAction performs an action roll: 1d6 + stat + adds vs 2d10.
func RollAction(r Roller, stat string, statValue, adds int) model.ActionRoll {
	actionDie := r.D6()
	c1 := r.D10()
	c2 := r.D10()
	actionScore := actionDie + statValue + adds
	if actionScore > 10 {
		actionScore = 10
	}

	outcome := model.DetermineOutcome(actionScore, c1, c2)

	return model.ActionRoll{
		ActionDie:     actionDie,
		ChallengeDie1: c1,
		ChallengeDie2: c2,
		Stat:          stat,
		StatValue:     statValue,
		Adds:          adds,
		ActionScore:   actionScore,
		Outcome:       outcome,
	}
}

// RollProgress performs a progress roll: progress score vs 2d10.
func RollProgress(r Roller, progressScore int) model.ProgressRollResult {
	c1 := r.D10()
	c2 := r.D10()
	outcome := model.DetermineOutcome(progressScore, c1, c2)

	return model.ProgressRollResult{
		ProgressScore: progressScore,
		ChallengeDie1: c1,
		ChallengeDie2: c2,
		Outcome:       outcome,
	}
}

// CheckBurnMomentum determines if burning momentum would improve the outcome.
func CheckBurnMomentum(momentum int, roll model.ActionRoll) (canBurn bool, newOutcome model.Outcome) {
	if momentum <= 0 {
		return false, ""
	}
	if roll.Outcome.IsStrongHit() {
		return false, ""
	}

	burnOutcome := model.DetermineOutcome(momentum, roll.ChallengeDie1, roll.ChallengeDie2)

	// Only offer burn if it improves the outcome
	if roll.Outcome == model.OutcomeMiss && burnOutcome.IsHit() {
		return true, burnOutcome
	}
	if roll.Outcome == model.OutcomeWeakHit && burnOutcome.IsStrongHit() {
		return true, burnOutcome
	}
	if roll.Outcome == model.OutcomeMissMatch && burnOutcome.IsHit() {
		return true, burnOutcome
	}

	return false, ""
}
