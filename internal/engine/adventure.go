package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerAdventureMoves() {
	mr.register(model.MoveDefinition{
		ID:          "face_danger",
		Name:        "Face Danger",
		Category:    model.CategoryAdventure,
		Description: "When you attempt something risky or react to an imminent threat, envision your action and roll.",
		StrongHit:   "You are successful. Take +1 momentum.",
		WeakHit:     "You succeed, but face a troublesome cost. Choose one.",
		Miss:        "You fail, or your progress is undermined by a dramatic and costly turn of events. Pay the Price.",
		Stats:       []string{"edge", "heart", "iron", "shadow", "wits"},
	}, executeFaceDanger)

	mr.register(model.MoveDefinition{
		ID:          "secure_an_advantage",
		Name:        "Secure an Advantage",
		Category:    model.CategoryAdventure,
		Description: "When you assess a situation, make preparations, or attempt to gain leverage, envision your action and roll.",
		StrongHit:   "You gain advantage. Choose one: Take control (+2 momentum) or Prepare for action (+1 on next move).",
		WeakHit:     "Your advantage is short-lived. Take +1 momentum.",
		Miss:        "You fail or your assumptions betray you. Pay the Price.",
		Stats:       []string{"edge", "heart", "iron", "shadow", "wits"},
	}, executeSecureAnAdvantage)

	mr.register(model.MoveDefinition{
		ID:          "gather_information",
		Name:        "Gather Information",
		Category:    model.CategoryAdventure,
		Description: "When you search an area, ask questions, conduct an investigation, or follow a track, roll +wits.",
		StrongHit:   "You discover something helpful and specific. Take +2 momentum.",
		WeakHit:     "The information complicates your quest or introduces a new danger. Take +1 momentum.",
		Miss:        "Your investigation unearths a dire threat or reveals an unwelcome truth. Pay the Price.",
		Stats:       []string{"wits"},
	}, executeGatherInformation)

	mr.register(model.MoveDefinition{
		ID:          "heal",
		Name:        "Heal",
		Category:    model.CategoryAdventure,
		Description: "When you treat an injury or ailment, roll +iron or +wits.",
		StrongHit:   "Your care is effective. If you are wounded, clear the debility and take or give +2 health.",
		WeakHit:     "Your care is effective but you must give something up. Choose one.",
		Miss:        "Your aid is ineffective. Pay the Price.",
		Stats:       []string{"iron", "wits"},
	}, executeHeal)

	mr.register(model.MoveDefinition{
		ID:          "resupply",
		Name:        "Resupply",
		Category:    model.CategoryAdventure,
		Description: "When you hunt, forage, or scavenge, roll +wits.",
		StrongHit:   "You bolster your resources. Take +2 supply.",
		WeakHit:     "Take up to +2 supply, but suffer -1 momentum for each.",
		Miss:        "You find nothing helpful. Pay the Price.",
		Stats:       []string{"wits"},
	}, executeResupply)

	mr.register(model.MoveDefinition{
		ID:          "make_camp",
		Name:        "Make Camp",
		Category:    model.CategoryAdventure,
		Description: "When you rest and recover for several hours in the wild, roll +supply.",
		StrongHit:   "You and your allies may each choose two from the list.",
		WeakHit:     "You and your allies may each choose one from the list.",
		Miss:        "You take no comfort. Pay the Price.",
		Stats:       []string{"supply"},
	}, executeMakeCamp)

	mr.register(model.MoveDefinition{
		ID:          "undertake_a_journey",
		Name:        "Undertake a Journey",
		Category:    model.CategoryAdventure,
		Description: "When you travel across hazardous or unfamiliar lands, first set the rank of your journey. Then, for each segment, roll +wits.",
		StrongHit:   "You reach a waypoint. Mark progress on your journey.",
		WeakHit:     "You reach a waypoint and mark progress, but suffer -1 supply.",
		Miss:        "You are waylaid by a perilous event. Pay the Price.",
		Stats:       []string{"wits"},
		IsProgress:  false,
	}, executeUndertakeAJourney)

	mr.register(model.MoveDefinition{
		ID:          "reach_your_destination",
		Name:        "Reach Your Destination",
		Category:    model.CategoryAdventure,
		Description: "When your journey progress is complete, roll the challenge dice against your progress score.",
		StrongHit:   "You reach your destination and the situation favors you. Choose one.",
		WeakHit:     "You arrive but face an unforeseen hazard or complication.",
		Miss:        "You have gone hopelessly astray, your objective is lost, or you were misled. Pay the Price.",
		IsProgress:  true,
	}, executeReachYourDestination)
}

func executeFaceDanger(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("face danger: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "face_danger", Name: "Face Danger", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum"}
		result.Summary = fmt.Sprintf("Faced danger with %s: Strong Hit (+1 momentum)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Succeed with a cost (player chooses)"}
		result.Summary = fmt.Sprintf("Faced danger with %s: Weak Hit (succeed with cost)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Fail or progress undermined. Pay the Price."}
		result.Summary = fmt.Sprintf("Faced danger with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeSecureAnAdvantage(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("secure an advantage: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "secure_an_advantage", Name: "Secure an Advantage", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 momentum (or +1 on next move)"}
		result.Summary = fmt.Sprintf("Secured an advantage with %s: Strong Hit (+2 momentum)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum"}
		result.Summary = fmt.Sprintf("Secured an advantage with %s: Weak Hit (+1 momentum)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Fail. Pay the Price."}
		result.Summary = fmt.Sprintf("Secured an advantage with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeGatherInformation(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "wits"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("gather information: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "gather_information", Name: "Gather Information", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 momentum"}
		result.Summary = fmt.Sprintf("Gathered information with %s: Strong Hit (+2 momentum)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum"}
		result.Summary = fmt.Sprintf("Gathered information with %s: Weak Hit (+1 momentum)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Investigation reveals dire threat. Pay the Price."}
		result.Summary = fmt.Sprintf("Gathered information with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeHeal(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "wits"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("heal: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "heal", Name: "Heal", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Health = model.Clamp(ch.Health+2, 0, 5)
		result.MechanicalEffects = []string{"+2 health (or clear wounded and +2 health)"}
		result.Summary = fmt.Sprintf("Healed with %s: Strong Hit (+2 health)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Health = model.Clamp(ch.Health+2, 0, 5)
		result.MechanicalEffects = []string{"+2 health but at a cost (player chooses)"}
		result.Summary = fmt.Sprintf("Healed with %s: Weak Hit (+2 health with cost)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Aid is ineffective. Pay the Price."}
		result.Summary = fmt.Sprintf("Healed with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeResupply(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "wits"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("resupply: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "resupply", Name: "Resupply", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Supply = model.Clamp(ch.Supply+2, 0, 5)
		result.MechanicalEffects = []string{"+2 supply"}
		result.Summary = fmt.Sprintf("Resupplied with %s: Strong Hit (+2 supply)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		// Take up to +2 supply, but -1 momentum per supply taken; default: take +2 supply, -2 momentum
		ch.Supply = model.Clamp(ch.Supply+2, 0, 5)
		ch.Momentum = model.Clamp(ch.Momentum-2, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+2 supply, -2 momentum"}
		result.Summary = fmt.Sprintf("Resupplied with %s: Weak Hit (+2 supply, -2 momentum)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Find nothing helpful. Pay the Price."}
		result.Summary = fmt.Sprintf("Resupplied with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeMakeCamp(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	// Make Camp rolls +supply
	roll := RollAction(r, "supply", ch.Supply, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "make_camp", Name: "Make Camp", Category: model.CategoryAdventure},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		result.MechanicalEffects = []string{"Choose two: Recuperate (+1 health), Partake (+1 supply), Relax (+1 spirit), Focus (+1 momentum), Prepare (+1 on next move)"}
		result.Summary = "Made camp: Strong Hit (choose two benefits)"
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Choose one: Recuperate (+1 health), Partake (+1 supply), Relax (+1 spirit), Focus (+1 momentum), Prepare (+1 on next move)"}
		result.Summary = "Made camp: Weak Hit (choose one benefit)"
	default:
		result.MechanicalEffects = []string{"You take no comfort. Pay the Price."}
		result.Summary = "Made camp: Miss (Pay the Price)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeUndertakeAJourney(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "wits"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("undertake a journey: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "undertake_a_journey", Name: "Undertake a Journey", Category: model.CategoryAdventure},
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
		result.MechanicalEffects = []string{"Mark progress on journey"}
		result.Summary = fmt.Sprintf("Undertook a journey with %s: Strong Hit (mark progress)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					t.MarkProgress()
					break
				}
			}
		}
		ch.Supply = model.Clamp(ch.Supply-1, 0, 5)
		result.MechanicalEffects = []string{"Mark progress on journey, -1 supply"}
		result.Summary = fmt.Sprintf("Undertook a journey with %s: Weak Hit (mark progress, -1 supply)", req.Stat)
	default:
		result.MechanicalEffects = []string{"Waylaid by perilous event. Pay the Price."}
		result.Summary = fmt.Sprintf("Undertook a journey with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeReachYourDestination(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.TrackID == "" {
		return nil, fmt.Errorf("reach your destination requires a track_id")
	}

	var track *model.ProgressTrack
	for _, t := range tracks {
		if t.ID == req.TrackID {
			track = t
			break
		}
	}
	if track == nil {
		return nil, fmt.Errorf("journey track not found: %s", req.TrackID)
	}

	progRoll := RollProgress(r, track.Score())
	result := &model.MoveResult{
		Move:         model.MoveDefinition{ID: "reach_your_destination", Name: "Reach Your Destination", Category: model.CategoryAdventure},
		ProgressRoll: &progRoll,
		Outcome:      progRoll.Outcome,
	}

	switch {
	case progRoll.Outcome.IsStrongHit():
		track.Completed = true
		result.MechanicalEffects = []string{"Reach destination, situation favors you"}
		result.Summary = "Reached destination: Strong Hit (situation favors you)"
	case progRoll.Outcome == model.OutcomeWeakHit:
		track.Completed = true
		result.MechanicalEffects = []string{"Arrive but face unforeseen complication"}
		result.Summary = "Reached destination: Weak Hit (unforeseen complication)"
	default:
		result.MechanicalEffects = []string{"Gone astray or objective lost. Pay the Price."}
		result.Summary = "Reached destination: Miss (Pay the Price)"
	}

	return result, nil
}
