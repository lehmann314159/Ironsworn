package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerRelationshipMoves() {
	mr.register(model.MoveDefinition{
		ID:          "compel",
		Name:        "Compel",
		Category:    model.CategoryRelationship,
		Description: "When you attempt to persuade someone to do something, envision your approach. If charming/pacifying: +heart. If threatening/inciting: +iron. If lying/swindling: +shadow.",
		StrongHit:   "They'll do what you want or share what you know. Take +1 momentum.",
		WeakHit:     "They'll do it, but with a demand or complication. Take +1 momentum.",
		Miss:        "They refuse or make a demand that costs you. Pay the Price.",
		Stats:       []string{"heart", "iron", "shadow"},
	}, executeCompel)

	mr.register(model.MoveDefinition{
		ID:          "sojourn",
		Name:        "Sojourn",
		Category:    model.CategoryRelationship,
		Description: "When you spend time in a community seeking help, roll +heart.",
		StrongHit:   "You and your allies may each choose two from the list.",
		WeakHit:     "You and your allies may each choose one from the list.",
		Miss:        "You find no help here. Pay the Price.",
		Stats:       []string{"heart"},
	}, executeSojourn)

	mr.register(model.MoveDefinition{
		ID:          "draw_the_circle",
		Name:        "Draw the Circle",
		Category:    model.CategoryRelationship,
		Description: "When you challenge someone to a formal duel or contest, roll +heart.",
		StrongHit:   "Take +1 momentum. You may also choose up to three boasts.",
		WeakHit:     "Take +1 momentum.",
		Miss:        "You begin the duel at a disadvantage. Pay the Price.",
		Stats:       []string{"heart"},
	}, executeDrawTheCircle)

	mr.register(model.MoveDefinition{
		ID:          "forge_a_bond",
		Name:        "Forge a Bond",
		Category:    model.CategoryRelationship,
		Description: "When you spend significant time with a person or community, stand together to face hardships, or make sacrifices for their cause, mark progress on your bonds track.",
		StrongHit:   "You forge a bond. Mark progress on your bonds track.",
		WeakHit:     "They ask something more of you first. When done, mark progress.",
		Miss:        "They reveal a demand or complication that undermines your relationship.",
		Stats:       []string{"heart"},
	}, executeForgeABond)

	mr.register(model.MoveDefinition{
		ID:          "test_your_bond",
		Name:        "Test Your Bond",
		Category:    model.CategoryRelationship,
		Description: "When your bond is tested through conflict, betrayal, or circumstance, roll +heart.",
		StrongHit:   "This test only strengthens your bond. Choose one: take +1 spirit or +1 momentum.",
		WeakHit:     "Your bond is fragile. Choose: give something up or lose the bond.",
		Miss:        "The bond is broken. Pay the Price.",
		Stats:       []string{"heart"},
	}, executeTestYourBond)

	mr.register(model.MoveDefinition{
		ID:          "aid_your_ally",
		Name:        "Aid Your Ally",
		Category:    model.CategoryRelationship,
		Description: "When you Secure an Advantage in direct support of an ally, and score a hit, they take the benefits of the move.",
	}, executeAidYourAlly)

	mr.register(model.MoveDefinition{
		ID:          "write_your_epilogue",
		Name:        "Write Your Epilogue",
		Category:    model.CategoryRelationship,
		Description: "When you retire from your life as Ironsworn, roll the challenge dice against your bonds progress score.",
		StrongHit:   "Things are good. Envision your epilogue.",
		WeakHit:     "You settle but your life is not what you hoped.",
		Miss:        "Your days are bleak. Envision a dark epilogue.",
		IsProgress:  true,
	}, executeWriteYourEpilogue)
}

func executeCompel(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("compel: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "compel", Name: "Compel", Category: model.CategoryRelationship},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum", "They do what you want"}
		result.Summary = fmt.Sprintf("Compelled with %s: Strong Hit (+1 momentum)", req.Stat)
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum", "They agree but with a demand"}
		result.Summary = fmt.Sprintf("Compelled with %s: Weak Hit (+1 momentum, with demand)", req.Stat)
	default:
		result.MechanicalEffects = []string{"They refuse. Pay the Price."}
		result.Summary = fmt.Sprintf("Compelled with %s: Miss (Pay the Price)", req.Stat)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeSojourn(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("sojourn: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "sojourn", Name: "Sojourn", Category: model.CategoryRelationship},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		result.MechanicalEffects = []string{"Choose two: Mend (+2 health), Hearten (+2 spirit), Equip (+2 supply), Plan (+2 momentum)"}
		result.Summary = "Sojourned: Strong Hit (choose two benefits)"
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Choose one: Mend (+2 health), Hearten (+2 spirit), Equip (+2 supply), Plan (+2 momentum)"}
		result.Summary = "Sojourned: Weak Hit (choose one benefit)"
	default:
		result.MechanicalEffects = []string{"No help here. Pay the Price."}
		result.Summary = "Sojourned: Miss (Pay the Price)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeDrawTheCircle(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("draw the circle: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "draw_the_circle", Name: "Draw the Circle", Category: model.CategoryRelationship},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum", "May make up to three boasts"}
		result.Summary = "Drew the Circle: Strong Hit (+1 momentum, boasts)"
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 momentum"}
		result.Summary = "Drew the Circle: Weak Hit (+1 momentum)"
	default:
		result.MechanicalEffects = []string{"Begin at a disadvantage. Pay the Price."}
		result.Summary = "Drew the Circle: Miss (Pay the Price)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeForgeABond(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("forge a bond: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "forge_a_bond", Name: "Forge a Bond", Category: model.CategoryRelationship},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		// Mark bonds progress
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					t.MarkProgress()
					break
				}
			}
		}
		result.MechanicalEffects = []string{"Bond forged. Mark progress on bonds track."}
		result.Summary = "Forged a Bond: Strong Hit (mark bonds progress)"
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"They ask something more first. Then mark bonds progress."}
		result.Summary = "Forged a Bond: Weak Hit (demand first)"
	default:
		result.MechanicalEffects = []string{"Relationship undermined. Pay the Price."}
		result.Summary = "Forged a Bond: Miss (Pay the Price)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeTestYourBond(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	if req.Stat == "" {
		req.Stat = "heart"
	}
	statVal, err := ch.Stats.GetStat(req.Stat)
	if err != nil {
		return nil, fmt.Errorf("test your bond: %w", err)
	}
	roll := RollAction(r, req.Stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "test_your_bond", Name: "Test Your Bond", Category: model.CategoryRelationship},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"Bond strengthened. +1 momentum (or +1 spirit)"}
		result.Summary = "Tested Bond: Strong Hit (bond strengthened)"
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Bond fragile. Choose: give something up or lose the bond."}
		result.Summary = "Tested Bond: Weak Hit (bond fragile)"
	default:
		result.MechanicalEffects = []string{"Bond broken. Pay the Price."}
		result.Summary = "Tested Bond: Miss (bond broken)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeAidYourAlly(_ Roller, _ *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "aid_your_ally", Name: "Aid Your Ally", Category: model.CategoryRelationship},
		Outcome:           model.OutcomeStrongHit,
		MechanicalEffects: []string{"Use Secure an Advantage in support of ally; they take the benefits"},
		Summary:           "Aid Your Ally: use Secure an Advantage for ally's benefit",
	}, nil
}

func executeWriteYourEpilogue(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	// Find bonds track
	var bondsTrack *model.ProgressTrack
	for _, t := range tracks {
		if t.TrackType == model.TrackBonds {
			bondsTrack = t
			break
		}
	}
	if bondsTrack == nil {
		if req.TrackID != "" {
			for _, t := range tracks {
				if t.ID == req.TrackID {
					bondsTrack = t
					break
				}
			}
		}
	}
	if bondsTrack == nil {
		return nil, fmt.Errorf("bonds track not found")
	}

	progRoll := RollProgress(r, bondsTrack.Score())
	result := &model.MoveResult{
		Move:         model.MoveDefinition{ID: "write_your_epilogue", Name: "Write Your Epilogue", Category: model.CategoryRelationship},
		ProgressRoll: &progRoll,
		Outcome:      progRoll.Outcome,
	}

	switch {
	case progRoll.Outcome.IsStrongHit():
		result.MechanicalEffects = []string{"Things are good. Envision a fulfilling epilogue."}
		result.Summary = "Wrote Epilogue: Strong Hit (good ending)"
	case progRoll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"You settle but life is not what you hoped."}
		result.Summary = "Wrote Epilogue: Weak Hit (bittersweet ending)"
	default:
		result.MechanicalEffects = []string{"Your days are bleak."}
		result.Summary = "Wrote Epilogue: Miss (dark ending)"
	}

	return result, nil
}
