package engine

import (
	"testing"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func TestRollAction_StrongHit(t *testing.T) {
	r := &FixedRoller{D6Values: []int{6}, D10Values: []int{1, 2}}
	roll := RollAction(r, "iron", 2, 0) // 6 + 2 = 8 vs 1, 2 → strong hit
	if roll.Outcome != model.OutcomeStrongHit {
		t.Errorf("expected strong_hit, got %s", roll.Outcome)
	}
	if roll.ActionScore != 8 {
		t.Errorf("expected action score 8, got %d", roll.ActionScore)
	}
}

func TestRollAction_WeakHit(t *testing.T) {
	r := &FixedRoller{D6Values: []int{3}, D10Values: []int{2, 8}}
	roll := RollAction(r, "edge", 2, 0) // 3 + 2 = 5 vs 2, 8 → weak hit (beats 2 but not 8)
	if roll.Outcome != model.OutcomeWeakHit {
		t.Errorf("expected weak_hit, got %s", roll.Outcome)
	}
}

func TestRollAction_Miss(t *testing.T) {
	r := &FixedRoller{D6Values: []int{1}, D10Values: []int{8, 9}}
	roll := RollAction(r, "wits", 1, 0) // 1 + 1 = 2 vs 8, 9 → miss
	if roll.Outcome != model.OutcomeMiss {
		t.Errorf("expected miss, got %s", roll.Outcome)
	}
}

func TestRollAction_StrongHitMatch(t *testing.T) {
	r := &FixedRoller{D6Values: []int{6}, D10Values: []int{3, 3}}
	roll := RollAction(r, "iron", 3, 0) // 6 + 3 = 9 vs 3, 3 → strong hit with match
	if roll.Outcome != model.OutcomeStrongHitMatch {
		t.Errorf("expected strong_hit_match, got %s", roll.Outcome)
	}
}

func TestRollAction_MissMatch(t *testing.T) {
	r := &FixedRoller{D6Values: []int{1}, D10Values: []int{5, 5}}
	roll := RollAction(r, "wits", 1, 0) // 1 + 1 = 2 vs 5, 5 → miss with match
	if roll.Outcome != model.OutcomeMissMatch {
		t.Errorf("expected miss_match, got %s", roll.Outcome)
	}
}

func TestRollAction_ActionScoreCappedAt10(t *testing.T) {
	r := &FixedRoller{D6Values: []int{6}, D10Values: []int{1, 1}}
	roll := RollAction(r, "iron", 3, 5) // 6 + 3 + 5 = 14, capped at 10
	if roll.ActionScore != 10 {
		t.Errorf("expected action score capped at 10, got %d", roll.ActionScore)
	}
}

func TestRollProgress(t *testing.T) {
	r := &FixedRoller{D10Values: []int{3, 7}}
	result := RollProgress(r, 8) // 8 vs 3, 7 → strong hit
	if result.Outcome != model.OutcomeStrongHit {
		t.Errorf("expected strong_hit, got %s", result.Outcome)
	}
}

func TestCheckBurnMomentum(t *testing.T) {
	roll := model.ActionRoll{
		ActionDie: 1, ChallengeDie1: 4, ChallengeDie2: 6,
		StatValue: 1, ActionScore: 2,
		Outcome: model.OutcomeMiss,
	}

	// Momentum of 7 beats both 4 and 6
	canBurn, newOutcome := CheckBurnMomentum(7, roll)
	if !canBurn {
		t.Error("expected can burn momentum")
	}
	if !newOutcome.IsStrongHit() {
		t.Errorf("expected strong hit after burn, got %s", newOutcome)
	}

	// Momentum of 0 cannot burn
	canBurn, _ = CheckBurnMomentum(0, roll)
	if canBurn {
		t.Error("should not be able to burn with 0 momentum")
	}

	// Already strong hit - no burn needed
	strongRoll := model.ActionRoll{
		ActionDie: 6, ChallengeDie1: 1, ChallengeDie2: 2,
		ActionScore: 8, Outcome: model.OutcomeStrongHit,
	}
	canBurn, _ = CheckBurnMomentum(9, strongRoll)
	if canBurn {
		t.Error("should not burn when already strong hit")
	}
}
