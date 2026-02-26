package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerQuestMoves() {
	mr.register(model.MoveDefinition{
		ID:          "swear_an_iron_vow",
		Name:        "Swear an Iron Vow",
		Category:    model.CategoryQuest,
		Description: "When you swear upon iron to complete a quest, write your vow and give it a rank. Then roll +heart.",
		StrongHit:   "You are emboldened and it is clear what you must do next. Take +2 momentum.",
		WeakHit:     "You are determined but begin your quest with more questions than answers. Take +1 momentum.",
		Miss:        "You face a significant obstacle before you can begin your quest. Pay the Price.",
		Stats:       []string{"heart"},
	}, executeSwearAnIronVow)

	mr.register(model.MoveDefinition{
		ID:          "reach_a_milestone",
		Name:        "Reach a Milestone",
		Category:    model.CategoryQuest,
		Description: "When you make significant progress in your quest by overcoming a critical obstacle, completing a perilous journey, solving a complex mystery, defeating a powerful threat, gaining vital support, or acquiring a crucial item, mark progress on your vow.",
	}, executeReachAMilestone)

	mr.register(model.MoveDefinition{
		ID:          "fulfill_your_vow",
		Name:        "Fulfill Your Vow",
		Category:    model.CategoryQuest,
		Description: "When you achieve what you believe to be the fulfillment of your vow, roll the challenge dice against your progress score.",
		StrongHit:   "Your quest is complete. Mark experience: Troublesome=1, Dangerous=2, Formidable=3, Extreme=4, Epic=5.",
		WeakHit:     "There is more to be done or you realize the truth of your quest. Envision what you discover. Then mark experience as above.",
		Miss:        "Your quest is undone. Envision what happens and choose: recommit or give up (Forsake Your Vow).",
		IsProgress:  true,
	}, executeFulfillYourVow)

	mr.register(model.MoveDefinition{
		ID:          "forsake_your_vow",
		Name:        "Forsake Your Vow",
		Category:    model.CategoryQuest,
		Description: "When you renounce your quest, betray your promise, or the goal is lost to you, clear the vow. Envision the impact and Endure Stress. Suffer -spirit equal to rank: Troublesome=1, Dangerous=2, Formidable=3, Extreme=4, Epic=5.",
	}, executeForsakeYourVow)
}

func executeSwearAnIronVow(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("swear an iron vow: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "swear_an_iron_vow", Name: "Swear an Iron Vow", Category: model.CategoryQuest},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 momentum", "Clear what to do next"}
		result.Summary = "Swore an Iron Vow: Strong Hit (+2 momentum)"
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum", "More questions than answers"}
		result.Summary = "Swore an Iron Vow: Weak Hit (+1 momentum)"
	default:
		result.MechanicalEffects = []string{"Face significant obstacle. Pay the Price."}
		result.Summary = "Swore an Iron Vow: Miss (Pay the Price)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeReachAMilestone(_ Roller, _ *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.TrackID == "" {
		return nil, fmt.Errorf("reach a milestone requires a track_id")
	}

	var track *model.ProgressTrack
	for _, t := range tracks {
		if t.ID == req.TrackID {
			track = t
			break
		}
	}
	if track == nil {
		return nil, fmt.Errorf("vow track not found: %s", req.TrackID)
	}

	ticks, err := track.MarkProgress()
	if err != nil {
		return nil, err
	}

	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "reach_a_milestone", Name: "Reach a Milestone", Category: model.CategoryQuest},
		Outcome:           model.OutcomeStrongHit, // Milestones always succeed
		MechanicalEffects: []string{fmt.Sprintf("Marked %d ticks on %s (now %d/40, score %d)", ticks, track.Name, track.Ticks, track.Score())},
		Summary:           fmt.Sprintf("Reached a milestone on %s: +%d ticks (score: %d)", track.Name, ticks, track.Score()),
	}, nil
}

func executeFulfillYourVow(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.TrackID == "" {
		return nil, fmt.Errorf("fulfill your vow requires a track_id")
	}

	var track *model.ProgressTrack
	for _, t := range tracks {
		if t.ID == req.TrackID {
			track = t
			break
		}
	}
	if track == nil {
		return nil, fmt.Errorf("vow track not found: %s", req.TrackID)
	}

	progRoll := RollProgress(r, track.Score())
	result := &model.MoveResult{
		Move:         model.MoveDefinition{ID: "fulfill_your_vow", Name: "Fulfill Your Vow", Category: model.CategoryQuest},
		ProgressRoll: &progRoll,
		Outcome:      progRoll.Outcome,
	}

	xp := rankXP(track.Rank)

	switch {
	case progRoll.Outcome.IsStrongHit():
		track.Completed = true
		ch.ExperienceEarned += xp
		result.MechanicalEffects = []string{
			fmt.Sprintf("Vow complete! +%d experience", xp),
		}
		result.Summary = fmt.Sprintf("Fulfilled vow '%s': Strong Hit (+%d XP)", track.Name, xp)
	case progRoll.Outcome == model.OutcomeWeakHit:
		track.Completed = true
		ch.ExperienceEarned += xp
		result.MechanicalEffects = []string{
			fmt.Sprintf("+%d experience, but there is more to discover", xp),
		}
		result.Summary = fmt.Sprintf("Fulfilled vow '%s': Weak Hit (+%d XP, complication)", track.Name, xp)
	default:
		result.MechanicalEffects = []string{"Quest undone. Recommit or Forsake Your Vow."}
		result.Summary = fmt.Sprintf("Failed to fulfill vow '%s': Miss (quest undone)", track.Name)
	}

	return result, nil
}

func executeForsakeYourVow(_ Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.TrackID == "" {
		return nil, fmt.Errorf("forsake your vow requires a track_id")
	}

	var track *model.ProgressTrack
	for _, t := range tracks {
		if t.ID == req.TrackID {
			track = t
			break
		}
	}
	if track == nil {
		return nil, fmt.Errorf("vow track not found: %s", req.TrackID)
	}

	spiritLoss := rankXP(track.Rank)
	ch.Spirit = model.Clamp(ch.Spirit-spiritLoss, 0, 5)
	track.Completed = true

	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "forsake_your_vow", Name: "Forsake Your Vow", Category: model.CategoryQuest},
		Outcome:           model.OutcomeMiss,
		MechanicalEffects: []string{fmt.Sprintf("-%d spirit", spiritLoss), "Vow forsaken"},
		Summary:           fmt.Sprintf("Forsook vow '%s': -%d spirit", track.Name, spiritLoss),
	}, nil
}

func rankXP(rank model.ProgressRank) int {
	switch rank {
	case model.RankTroublesome:
		return 1
	case model.RankDangerous:
		return 2
	case model.RankFormidable:
		return 3
	case model.RankExtreme:
		return 4
	case model.RankEpic:
		return 5
	default:
		return 0
	}
}
