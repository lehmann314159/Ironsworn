package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerSufferMoves() {
	mr.register(model.MoveDefinition{
		ID:          "endure_harm",
		Name:        "Endure Harm",
		Category:    model.CategorySuffer,
		Description: "When you face physical damage, suffer -health equal to your foe's rank or as pointed. If health is 0, suffer -momentum equal to remaining.",
		StrongHit:   "Choose one: Shake it off (+1 health) or Embrace the pain (+1 momentum).",
		WeakHit:     "You press on.",
		Miss:        "You must also suffer -1 momentum. If health is 0, you must mark wounded or roll on the Face Death table.",
		Stats:       []string{"health", "iron"},
	}, executeEndureHarm)

	mr.register(model.MoveDefinition{
		ID:          "endure_stress",
		Name:        "Endure Stress",
		Category:    model.CategorySuffer,
		Description: "When you face mental shock or despair, suffer -spirit equal to the stress. If spirit is 0, suffer -momentum equal to remaining.",
		StrongHit:   "Choose one: Shake it off (+1 spirit) or Embrace the darkness (+1 momentum).",
		WeakHit:     "You press on.",
		Miss:        "You must also suffer -1 momentum. If spirit is 0, you must mark shaken or roll on the Face Desolation table.",
		Stats:       []string{"spirit", "heart"},
	}, executeEndureStress)

	mr.register(model.MoveDefinition{
		ID:          "face_death",
		Name:        "Face Death",
		Category:    model.CategorySuffer,
		Description: "When you are brought to the brink of death, and your health is at 0, roll +heart.",
		StrongHit:   "Death rejects you. Take +1 health and +1 momentum.",
		WeakHit:     "You see a vision of a lost connection. Take +1 health.",
		Miss:        "You are dead.",
		Stats:       []string{"heart"},
	}, executeFaceDeath)

	mr.register(model.MoveDefinition{
		ID:          "face_desolation",
		Name:        "Face Desolation",
		Category:    model.CategorySuffer,
		Description: "When you are brought to the brink of desolation, and your spirit is at 0, roll +heart.",
		StrongHit:   "You resist and press on. Take +1 spirit.",
		WeakHit:     "Choose one: your spirit or sanity is shaken but you continue, or a connection bolsters you.",
		Miss:        "You succumb to despair or horror. You are lost.",
		Stats:       []string{"heart"},
	}, executeFaceDesolation)

	mr.register(model.MoveDefinition{
		ID:          "out_of_supply",
		Name:        "Out of Supply",
		Category:    model.CategorySuffer,
		Description: "When your supply is exhausted (supply is 0), mark unprepared. If already unprepared, suffer -health, -spirit, or -momentum.",
	}, executeOutOfSupply)

	mr.register(model.MoveDefinition{
		ID:          "face_a_setback",
		Name:        "Face a Setback",
		Category:    model.CategorySuffer,
		Description: "When your momentum is at its minimum (-6), and you must suffer -momentum, choose one: suffer a different track or lose progress on a vow.",
	}, executeFaceASetback)
}

func executeEndureHarm(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	amount := req.Amount
	if amount <= 0 {
		amount = 1
	}

	// Suffer the harm
	oldHealth := ch.Health
	ch.Health = model.Clamp(ch.Health-amount, 0, 5)
	actualLoss := oldHealth - ch.Health

	// If health would go below 0, overflow to momentum
	overflow := amount - actualLoss
	if overflow > 0 && oldHealth == 0 {
		ch.Momentum = model.Clamp(ch.Momentum-overflow, -6, ch.MomentumMax)
	}

	// Roll to endure
	stat := req.Stat
	if stat == "" {
		stat = "iron"
	}
	var statVal int
	if stat == "health" {
		statVal = ch.Health
	} else {
		var err error
		statVal, err = ch.Stats.GetStat(stat)
		if err != nil {
			return nil, fmt.Errorf("endure harm: %w", err)
		}
	}

	roll := RollAction(r, stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "endure_harm", Name: "Endure Harm", Category: model.CategorySuffer},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		// Player chooses: +1 health or +1 momentum
		ch.Health = model.Clamp(ch.Health+1, 0, 5)
		result.MechanicalEffects = []string{
			fmt.Sprintf("-%d health", amount),
			"Choose: +1 health (applied) or +1 momentum",
		}
		result.Summary = fmt.Sprintf("Endured %d harm: Strong Hit (shake it off)", amount)
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{fmt.Sprintf("-%d health", amount), "Press on"}
		result.Summary = fmt.Sprintf("Endured %d harm: Weak Hit (press on)", amount)
	default:
		ch.Momentum = model.Clamp(ch.Momentum-1, -6, ch.MomentumMax)
		effects := []string{fmt.Sprintf("-%d health", amount), "-1 momentum"}
		if ch.Health == 0 {
			effects = append(effects, "Health at 0: mark Wounded or Face Death")
		}
		result.MechanicalEffects = effects
		result.Summary = fmt.Sprintf("Endured %d harm: Miss (-1 momentum)", amount)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeEndureStress(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	amount := req.Amount
	if amount <= 0 {
		amount = 1
	}

	oldSpirit := ch.Spirit
	ch.Spirit = model.Clamp(ch.Spirit-amount, 0, 5)
	actualLoss := oldSpirit - ch.Spirit

	overflow := amount - actualLoss
	if overflow > 0 && oldSpirit == 0 {
		ch.Momentum = model.Clamp(ch.Momentum-overflow, -6, ch.MomentumMax)
	}

	stat := req.Stat
	if stat == "" {
		stat = "heart"
	}
	var statVal int
	if stat == "spirit" {
		statVal = ch.Spirit
	} else {
		var err error
		statVal, err = ch.Stats.GetStat(stat)
		if err != nil {
			return nil, fmt.Errorf("endure stress: %w", err)
		}
	}

	roll := RollAction(r, stat, statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "endure_stress", Name: "Endure Stress", Category: model.CategorySuffer},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Spirit = model.Clamp(ch.Spirit+1, 0, 5)
		result.MechanicalEffects = []string{
			fmt.Sprintf("-%d spirit", amount),
			"Choose: +1 spirit (applied) or +1 momentum",
		}
		result.Summary = fmt.Sprintf("Endured %d stress: Strong Hit (shake it off)", amount)
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{fmt.Sprintf("-%d spirit", amount), "Press on"}
		result.Summary = fmt.Sprintf("Endured %d stress: Weak Hit (press on)", amount)
	default:
		ch.Momentum = model.Clamp(ch.Momentum-1, -6, ch.MomentumMax)
		effects := []string{fmt.Sprintf("-%d spirit", amount), "-1 momentum"}
		if ch.Spirit == 0 {
			effects = append(effects, "Spirit at 0: mark Shaken or Face Desolation")
		}
		result.MechanicalEffects = effects
		result.Summary = fmt.Sprintf("Endured %d stress: Miss (-1 momentum)", amount)
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeFaceDeath(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	statVal, err := ch.Stats.GetStat("heart")
	if err != nil {
		return nil, err
	}
	roll := RollAction(r, "heart", statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "face_death", Name: "Face Death", Category: model.CategorySuffer},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Health = model.Clamp(ch.Health+1, 0, 5)
		ch.Momentum = model.Clamp(ch.Momentum+1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"+1 health", "+1 momentum", "Death rejects you"}
		result.Summary = "Faced Death: Strong Hit (death rejects you)"
	case roll.Outcome == model.OutcomeWeakHit:
		ch.Health = model.Clamp(ch.Health+1, 0, 5)
		result.MechanicalEffects = []string{"+1 health", "Vision of a lost connection"}
		result.Summary = "Faced Death: Weak Hit (vision of loss)"
	default:
		result.MechanicalEffects = []string{"You are dead."}
		result.Summary = "Faced Death: Miss (you are dead)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeFaceDesolation(r Roller, ch *model.Character, req model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	statVal, err := ch.Stats.GetStat("heart")
	if err != nil {
		return nil, err
	}
	roll := RollAction(r, "heart", statVal, req.Adds)
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "face_desolation", Name: "Face Desolation", Category: model.CategorySuffer},
		Roll:    &roll,
		Outcome: roll.Outcome,
	}

	switch {
	case roll.Outcome.IsStrongHit():
		ch.Spirit = model.Clamp(ch.Spirit+1, 0, 5)
		result.MechanicalEffects = []string{"+1 spirit", "Resist and press on"}
		result.Summary = "Faced Desolation: Strong Hit (resist)"
	case roll.Outcome == model.OutcomeWeakHit:
		result.MechanicalEffects = []string{"Shaken but continue"}
		result.Summary = "Faced Desolation: Weak Hit (shaken)"
	default:
		result.MechanicalEffects = []string{"You succumb to despair. You are lost."}
		result.Summary = "Faced Desolation: Miss (you are lost)"
	}

	canBurn, burnOutcome := CheckBurnMomentum(ch.Momentum, roll)
	result.CanBurnMomentum = canBurn
	result.BurnOutcome = burnOutcome
	return result, nil
}

func executeOutOfSupply(_ Roller, ch *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	result := &model.MoveResult{
		Move:    model.MoveDefinition{ID: "out_of_supply", Name: "Out of Supply", Category: model.CategorySuffer},
		Outcome: model.OutcomeMiss,
	}

	if !ch.Debilities.Unprepared {
		ch.Debilities.Unprepared = true
		ch.UpdateMomentumLimits()
		result.MechanicalEffects = []string{"Marked unprepared"}
		result.Summary = "Out of Supply: marked unprepared"
	} else {
		// Already unprepared, suffer -health, -spirit, or -momentum (default: -momentum)
		ch.Momentum = model.Clamp(ch.Momentum-1, -6, ch.MomentumMax)
		result.MechanicalEffects = []string{"Already unprepared. -1 momentum (or choose -health/-spirit)"}
		result.Summary = "Out of Supply: already unprepared (-1 momentum)"
	}

	return result, nil
}

func executeFaceASetback(_ Roller, ch *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "face_a_setback", Name: "Face a Setback", Category: model.CategorySuffer},
		Outcome:           model.OutcomeMiss,
		MechanicalEffects: []string{"Momentum at minimum. Choose: suffer different track or lose progress on a vow."},
		Summary:           "Face a Setback: momentum at minimum, must choose consequence",
	}, nil
}
