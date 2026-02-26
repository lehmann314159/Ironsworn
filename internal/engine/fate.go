package engine

import (
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (mr *MoveRegistry) registerFateMoves() {
	mr.register(model.MoveDefinition{
		ID:          "pay_the_price",
		Name:        "Pay the Price",
		Category:    model.CategoryFate,
		Description: "When you suffer the outcome of a move, choose: make the most obvious negative outcome happen, Ask the Oracle for inspiration (using the Pay the Price table), or roll on the table.",
	}, executePayThePrice)

	mr.register(model.MoveDefinition{
		ID:          "ask_the_oracle",
		Name:        "Ask the Oracle",
		Category:    model.CategoryFate,
		Description: "When you seek to resolve questions, discover details, reveal locations, determine actions of NPCs, or trigger encounters, you may ask a yes/no question or roll on an oracle table.",
	}, executeAskTheOracle)
}

func executePayThePrice(r Roller, _ *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	roll := r.D100()
	var result string
	switch {
	case roll <= 2:
		result = "Roll again and apply the result but make it worse."
	case roll <= 5:
		result = "A person or community you trusted loses faith in you, or acts against you."
	case roll <= 9:
		result = "A person or community you care about is exposed to danger."
	case roll <= 16:
		result = "You are separated from something or someone."
	case roll <= 23:
		result = "Your action has an unintended effect."
	case roll <= 32:
		result = "Something of value is lost or destroyed."
	case roll <= 41:
		result = "The current situation worsens."
	case roll <= 50:
		result = "A new danger or foe is revealed."
	case roll <= 59:
		result = "It causes a delay or forces a new approach."
	case roll <= 68:
		result = "It is harmful."
	case roll <= 76:
		result = "It is stressful."
	case roll <= 85:
		result = "A surprising development complicates your quest."
	case roll <= 90:
		result = "It wastes resources."
	case roll <= 94:
		result = "It forces you to act against your best intentions."
	case roll <= 98:
		result = "A friend, companion, or ally is put in harm's way."
	default:
		result = "Roll twice more. Both results occur. If they are the same, make it worse."
	}

	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "pay_the_price", Name: "Pay the Price", Category: model.CategoryFate},
		Outcome:           model.OutcomeMiss,
		MechanicalEffects: []string{result},
		Summary:           "Pay the Price: " + result,
	}, nil
}

func executeAskTheOracle(_ Roller, _ *model.Character, _ model.MoveRequest, _ []*model.ProgressTrack) (*model.MoveResult, error) {
	// The actual oracle roll is handled by the oracle package through the handler.
	// This move definition exists for the move registry catalog.
	return &model.MoveResult{
		Move:              model.MoveDefinition{ID: "ask_the_oracle", Name: "Ask the Oracle", Category: model.CategoryFate},
		MechanicalEffects: []string{"Use the oracle endpoints to ask yes/no questions or roll on tables."},
		Summary:           "Ask the Oracle: use oracle endpoints for resolution",
	}, nil
}
