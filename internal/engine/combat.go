package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerCombatMoves() {
	mr.register(model.MoveDefinition{
		ID:          "enter_the_fray",
		Name:        "Enter the Fray",
		Category:    model.CategoryCombat,
		Description: "When you enter into combat, set the rank of each foe. Then determine who is in control. If acting from ambush/stealth: +shadow. If acting within a community: +heart. Otherwise: +wits.",
		StrongHit:   "Take +2 momentum. You have initiative.",
		WeakHit:     "Choose: take +2 momentum (foe has initiative) or you have initiative.",
		Miss:        "Combat begins with you at a disadvantage. Pay the Price. Foe has initiative.",
		Stats:       []string{"heart", "shadow", "wits"},
	}, executeEnterTheFray)

	mr.register(model.MoveDefinition{
		ID:          "strike",
		Name:        "Strike",
		Category:    model.CategoryCombat,
		Description: "When you have initiative and attack in close quarters, roll +iron. When you attack at range, roll +edge.",
		StrongHit:   "Inflict +1 harm. You retain initiative.",
		WeakHit:     "Inflict harm. You lose initiative.",
		Miss:        "Your attack fails. Pay the Price. You lose initiative.",
		Stats:       []string{"iron", "edge"},
	}, executeStrike)

	mr.register(model.MoveDefinition{
		ID:          "clash",
		Name:        "Clash",
		Category:    model.CategoryCombat,
		Description: "When your foe has initiative and you fight with them, roll +iron (close) or +edge (range).",
		StrongHit:   "Inflict harm. You gain initiative.",
		WeakHit:     "Inflict harm, but your foe also inflicts harm. You lose initiative.",
		Miss:        "You are outmatched. Pay the Price. You lose initiative.",
		Stats:       []string{"iron", "edge"},
	}, executeClash)

	mr.register(model.MoveDefinition{
		ID:          "turn_the_tide",
		Name:        "Turn the Tide",
		Category:    model.CategoryCombat,
		Description: "Once per fight, when you risk it all, you may steal initiative. Take +1 momentum. On your next move, add +1. If you fail that move, Pay the Price (dire consequence).",
	}, executeTurnTheTide)

	mr.register(model.MoveDefinition{
		ID:          "end_the_fight",
		Name:        "End the Fight",
		Category:    model.CategoryCombat,
		Description: "When you make a decisive move to end the fight, roll the challenge dice against your progress score.",
		StrongHit:   "This foe is no longer in the fight. Choose appropriate narrative.",
		WeakHit:     "This foe is no longer in the fight, but you must choose one cost.",
		Miss:        "You have lost this fight. Pay the Price.",
		IsProgress:  true,
	}, executeEndTheFight)

	mr.register(model.MoveDefinition{
		ID:          "battle",
		Name:        "Battle",
		Category:    model.CategoryCombat,
		Description: "When you fight a battle and it happens in a blur, envision your approach and roll. For a pointed attack: +iron. For speed/range: +edge. For deception: +shadow. For strategy: +wits.",
		StrongHit:   "You achieve your objective unconditionally. Take +2 momentum.",
		WeakHit:     "You achieve your objective, but not without cost. Pay the Price.",
		Miss:        "You are defeated or the objective is lost. Pay the Price.",
		Stats:       []string{"edge", "iron", "shadow", "wits"},
	}, executeBattle)
}

func executeEnterTheFray(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "wits"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("enter the fray: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "enter_the_fray", Name: "Enter the Fray", Category: model.CategoryCombat},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 momentum", "You have initiative"}
		result.Summary = fmt.Sprintf("Entered the fray with %s: Strong Hit (+2 momentum, initiative)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Choose: +2 momentum (foe has initiative) or you have initiative"}
		result.Summary = fmt.Sprintf("Entered the fray with %s: Weak Hit (choose momentum or initiative)", req.Stat)
	default:
		result.MechanicalEffects = []string{"At a disadvantage. Pay the Price. Foe has initiative."}
		result.Summary = fmt.Sprintf("Entered the fray with %s: Miss (foe has initiative)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeStrike(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "iron"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("strike: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "strike", Name: "Strike", Category: model.CategoryCombat},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	// Mark progress on combat track if provided
	harm := 1
	switch {
	case roll.Outcome.IsStrongHit():
		harm = 2 // +1 harm on strong hit
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					for i := 0; i < harm; i++ {
						t.MarkProgress()
					}
					break
				}
			}
		}
		result.MechanicalEffects = []string{fmt.Sprintf("Inflict %d harm (mark progress x%d)", harm, harm), "Retain initiative"}
		result.Summary = fmt.Sprintf("Strike with %s: Strong Hit (inflict %d harm, retain initiative)", req.Stat, harm)
	case roll.Outcome == model.OutcomeWeakHit:
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					t.MarkProgress()
					break
				}
			}
		}
		result.MechanicalEffects = []string{fmt.Sprintf("Inflict %d harm (mark progress)", harm), "Lose initiative"}
		result.Summary = fmt.Sprintf("Strike with %s: Weak Hit (inflict %d harm, lose initiative)", req.Stat, harm)
	default:
		result.MechanicalEffects = []string{"Attack fails. Pay the Price.", "Lose initiative"}
		result.Summary = fmt.Sprintf("Strike with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeClash(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "iron"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("clash: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "clash", Name: "Clash", Category: model.CategoryCombat},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					t.MarkProgress()
					break
				}
			}
		}
		result.MechanicalEffects = []string{"Inflict harm (mark progress)", "Gain initiative"}
		result.Summary = fmt.Sprintf("Clashed with %s: Strong Hit (inflict harm, gain initiative)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					t.MarkProgress()
					break
				}
			}
		}
		result.MechanicalEffects = []string{"Inflict harm (mark progress)", "Foe also inflicts harm", "Lose initiative"}
		result.Summary = fmt.Sprintf("Clashed with %s: Weak Hit (trade harm, lose initiative)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Outmatched. Pay the Price.", "Lose initiative"}
		result.Summary = fmt.Sprintf("Clashed with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeTurnTheTide(_ Roller, ch *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "turn_the_tide", Name: "Turn the Tide", Category: model.CategoryCombat},
		Outcome:           model.OutcomeStrongHit,
		MechanicalEffects: []string{"+1 momentum", "Steal initiative", "+1 on next move", "If next move fails: dire consequence"},
		Summary:           "Turned the Tide: +1 momentum, steal initiative (+1 on next move)",
	}, nil
}

func executeEndTheFight(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.TrackID == "" {
		return nil, fmt.Errorf("end the fight requires a track_id")
	}

	var track *model.ProgressTrack
	for _, t := range tracks {
		if t.ID == req.TrackID {
			track = t
			break
		}
	}
	if track == nil {
		return nil, fmt.Errorf("combat track not found: %s", req.TrackID)
	}

	progRoll := RollProgress(r, track.Score())
	result := &model.MoveResult{
		Move:         model.MoveDefinition{ID: "end_the_fight", Name: "End the Fight", Category: model.CategoryCombat},
		ProgressRoll: &progRoll,
		Outcome:      progRoll.Outcome,
	}

	switch {
	case progRoll.Outcome.IsStrongHit():
		track.Completed = true
		result.MechanicalEffects = []string{"Foe defeated"}
		result.Summary = "Ended the fight: Strong Hit (foe defeated)"
	case progRoll.Outcome == model.OutcomeWeakHit:
		track.Completed = true
		result.MechanicalEffects = []string{"Foe defeated but at a cost (choose one)"}
		result.Summary = "Ended the fight: Weak Hit (foe defeated with cost)"
	default:
		result.MechanicalEffects = []string{"You have lost this fight. Pay the Price."}
		result.Summary = "Ended the fight: Miss (you lose)"
	}

	return result, nil
}

func executeBattle(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "iron"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("battle: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "battle", Name: "Battle", Category: model.CategoryCombat},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 momentum", "Objective achieved unconditionally"}
		result.Summary = fmt.Sprintf("Battle with %s: Strong Hit (+2 momentum, objective achieved)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Objective achieved but at a cost. Pay the Price."}
		result.Summary = fmt.Sprintf("Battle with %s: Weak Hit (objective at a cost)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Defeated or objective lost. Pay the Price."}
		result.Summary = fmt.Sprintf("Battle with %s: Miss (defeated)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}
