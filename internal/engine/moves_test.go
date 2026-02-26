package engine

import (
	"testing"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func newTestCharacter() *model.Character {
	ch, _ := model.NewCharacter("ch1", "game1", "Kara", model.Stats{
		Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1,
	})
	return ch
}

func TestFaceDanger_StrongHit(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D6Values: []int{6}, D10Values: []int{1, 2}}
	ch := newTestCharacter()
	oldMomentum := ch.Momentum

	result, err := mr.Execute(r, ch, model.MoveRequest{MoveID: "face_danger", Stat: "edge"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Outcome.IsStrongHit() {
		t.Errorf("expected strong hit, got %s", result.Outcome)
	}
	if ch.Momentum != oldMomentum+1 {
		t.Errorf("expected momentum %d, got %d", oldMomentum+1, ch.Momentum)
	}
}

func TestFaceDanger_Miss(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D6Values: []int{1}, D10Values: []int{8, 9}}
	ch := newTestCharacter()

	result, err := mr.Execute(r, ch, model.MoveRequest{MoveID: "face_danger", Stat: "wits"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Outcome != model.OutcomeMiss {
		t.Errorf("expected miss, got %s", result.Outcome)
	}
}

func TestSwearAnIronVow_StrongHit(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D6Values: []int{5}, D10Values: []int{2, 3}}
	ch := newTestCharacter()
	oldMomentum := ch.Momentum

	result, err := mr.Execute(r, ch, model.MoveRequest{MoveID: "swear_an_iron_vow"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Outcome.IsStrongHit() {
		t.Errorf("expected strong hit, got %s", result.Outcome)
	}
	if ch.Momentum != oldMomentum+2 {
		t.Errorf("expected momentum %d, got %d", oldMomentum+2, ch.Momentum)
	}
}

func TestReachAMilestone(t *testing.T) {
	mr := NewMoveRegistry()
	ch := newTestCharacter()
	track := &model.ProgressTrack{
		ID: "vow1", GameID: "game1", Name: "Protect the Village",
		TrackType: model.TrackVow, Rank: model.RankDangerous,
	}

	result, err := mr.Execute(nil, ch, model.MoveRequest{
		MoveID: "reach_a_milestone", TrackID: "vow1",
	}, []*model.ProgressTrack{track})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Ticks != 8 { // Dangerous = 8 ticks per mark
		t.Errorf("expected 8 ticks, got %d", track.Ticks)
	}
	if track.Score() != 2 { // 8/4 = 2
		t.Errorf("expected score 2, got %d", track.Score())
	}
	_ = result
}

func TestFulfillYourVow_StrongHit(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D10Values: []int{2, 3}} // progress score 8 beats both
	ch := newTestCharacter()

	track := &model.ProgressTrack{
		ID: "vow1", GameID: "game1", Name: "Protect the Village",
		TrackType: model.TrackVow, Rank: model.RankDangerous, Ticks: 32, // score = 8
	}

	result, err := mr.Execute(r, ch, model.MoveRequest{
		MoveID: "fulfill_your_vow", TrackID: "vow1",
	}, []*model.ProgressTrack{track})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Outcome.IsStrongHit() {
		t.Errorf("expected strong hit, got %s", result.Outcome)
	}
	if !track.Completed {
		t.Error("expected track to be completed")
	}
	if ch.ExperienceEarned != 2 { // Dangerous = 2 XP
		t.Errorf("expected 2 XP, got %d", ch.ExperienceEarned)
	}
}

func TestEndureHarm(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D6Values: []int{4}, D10Values: []int{2, 3}}
	ch := newTestCharacter()
	ch.Health = 5

	result, err := mr.Execute(r, ch, model.MoveRequest{
		MoveID: "endure_harm", Amount: 2,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Health should drop by 2 (5 → 3), then strong hit restores 1 (→ 4)
	if ch.Health != 4 {
		t.Errorf("expected health 4, got %d", ch.Health)
	}
	_ = result
}

func TestEndureStress(t *testing.T) {
	mr := NewMoveRegistry()
	r := &FixedRoller{D6Values: []int{5}, D10Values: []int{1, 2}}
	ch := newTestCharacter()
	ch.Spirit = 5

	_, err := mr.Execute(r, ch, model.MoveRequest{
		MoveID: "endure_stress", Amount: 3,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Spirit 5 → 2 from stress, strong hit restores 1 → 3
	if ch.Spirit != 3 {
		t.Errorf("expected spirit 3, got %d", ch.Spirit)
	}
}

func TestVowLoop(t *testing.T) {
	// Test the complete vow loop: swear → milestones → fulfill
	mr := NewMoveRegistry()
	ch := newTestCharacter()

	// 1. Swear a troublesome vow
	r := &FixedRoller{D6Values: []int{5}, D10Values: []int{2, 3}}
	_, err := mr.Execute(r, ch, model.MoveRequest{MoveID: "swear_an_iron_vow"}, nil)
	if err != nil {
		t.Fatalf("swear: %v", err)
	}

	// 2. Create the track
	track := &model.ProgressTrack{
		ID: "vow1", GameID: "game1", Name: "Find the Lost Artifact",
		TrackType: model.TrackVow, Rank: model.RankTroublesome,
	}
	tracks := []*model.ProgressTrack{track}

	// 3. Reach milestones (troublesome = 12 ticks/mark, need 3 marks for score 9)
	for i := 0; i < 3; i++ {
		_, err := mr.Execute(nil, ch, model.MoveRequest{
			MoveID: "reach_a_milestone", TrackID: "vow1",
		}, tracks)
		if err != nil {
			t.Fatalf("milestone %d: %v", i+1, err)
		}
	}
	if track.Ticks != 36 { // 3 * 12 = 36
		t.Errorf("expected 36 ticks, got %d", track.Ticks)
	}
	if track.Score() != 9 {
		t.Errorf("expected score 9, got %d", track.Score())
	}

	// 4. Fulfill the vow (score 9 vs 3, 5 → strong hit)
	r2 := &FixedRoller{D10Values: []int{3, 5}}
	result, err := mr.Execute(r2, ch, model.MoveRequest{
		MoveID: "fulfill_your_vow", TrackID: "vow1",
	}, tracks)
	if err != nil {
		t.Fatalf("fulfill: %v", err)
	}
	if !result.Outcome.IsStrongHit() {
		t.Errorf("expected strong hit, got %s", result.Outcome)
	}
	if !track.Completed {
		t.Error("expected vow completed")
	}
	if ch.ExperienceEarned != 1 { // Troublesome = 1 XP
		t.Errorf("expected 1 XP, got %d", ch.ExperienceEarned)
	}
}

func TestListDefinitions(t *testing.T) {
	mr := NewMoveRegistry()
	defs := mr.ListDefinitions()
	if len(defs) == 0 {
		t.Error("expected at least one move definition")
	}

	// Check a known move exists
	found := false
	for _, d := range defs {
		if d.ID == "face_danger" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected face_danger in definitions")
	}
}
