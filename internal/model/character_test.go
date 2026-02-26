package model

import "testing"

func TestStatsValidate(t *testing.T) {
	valid := Stats{Edge: 3, Heart: 2, Iron: 2, Shadow: 1, Wits: 1}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid stats rejected: %v", err)
	}

	// Sum != 9
	invalid := Stats{Edge: 3, Heart: 3, Iron: 2, Shadow: 1, Wits: 1}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for sum != 9")
	}

	// Value out of range
	outOfRange := Stats{Edge: 4, Heart: 2, Iron: 1, Shadow: 1, Wits: 1}
	if err := outOfRange.Validate(); err == nil {
		t.Error("expected error for out-of-range stat")
	}
}

func TestGetStat(t *testing.T) {
	s := Stats{Edge: 3, Heart: 2, Iron: 2, Shadow: 1, Wits: 1}

	val, err := s.GetStat("edge")
	if err != nil || val != 3 {
		t.Errorf("edge: got %d, %v", val, err)
	}

	_, err = s.GetStat("nonexistent")
	if err == nil {
		t.Error("expected error for unknown stat")
	}
}

func TestDebilitiesCount(t *testing.T) {
	d := Debilities{}
	if d.Count() != 0 {
		t.Errorf("expected 0, got %d", d.Count())
	}

	d.Wounded = true
	d.Maimed = true
	d.Cursed = true
	if d.Count() != 3 {
		t.Errorf("expected 3, got %d", d.Count())
	}
}

func TestNewCharacter(t *testing.T) {
	ch, err := NewCharacter("id", "gid", "Kara", Stats{Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Health != 5 || ch.Spirit != 5 || ch.Supply != 5 {
		t.Error("expected default tracks at 5")
	}
	if ch.Momentum != 2 {
		t.Errorf("expected momentum 2, got %d", ch.Momentum)
	}
}

func TestUpdateMomentumLimits(t *testing.T) {
	ch, _ := NewCharacter("id", "gid", "Kara", Stats{Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1})
	ch.Momentum = 10

	ch.Debilities.Wounded = true
	ch.Debilities.Shaken = true
	ch.UpdateMomentumLimits()

	if ch.MomentumMax != 8 { // 10 - 2
		t.Errorf("expected max 8, got %d", ch.MomentumMax)
	}
	if ch.MomentumReset != 0 { // 2 - 2
		t.Errorf("expected reset 0, got %d", ch.MomentumReset)
	}
	if ch.Momentum != 8 { // Clamped to new max
		t.Errorf("expected momentum clamped to 8, got %d", ch.Momentum)
	}
}

func TestClamp(t *testing.T) {
	if Clamp(15, 0, 10) != 10 {
		t.Error("expected clamped to max")
	}
	if Clamp(-5, 0, 10) != 0 {
		t.Error("expected clamped to min")
	}
	if Clamp(5, 0, 10) != 5 {
		t.Error("expected unchanged")
	}
}
