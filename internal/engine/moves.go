package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn/internal/model"
)

// MoveFunc executes a move, mutating character state and returning a result.
type MoveFunc func(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error)

// MoveRegistry maps move IDs to their definitions and execution functions.
type MoveRegistry struct {
	definitions map[string]model.MoveDefinition
	executors   map[string]MoveFunc
}

// NewMoveRegistry creates and populates the move registry with all defined moves.
func NewMoveRegistry() *MoveRegistry {
	mr := &MoveRegistry{
		definitions: make(map[string]model.MoveDefinition),
		executors:   make(map[string]MoveFunc),
	}
	mr.registerAdventureMoves()
	mr.registerQuestMoves()
	mr.registerSufferMoves()
	mr.registerCombatMoves()
	mr.registerRelationshipMoves()
	mr.registerFateMoves()
	return mr
}

func (mr *MoveRegistry) register(def model.MoveDefinition, fn MoveFunc) {
	mr.definitions[def.ID] = def
	mr.executors[def.ID] = fn
}

// GetDefinition returns the move definition for the given ID.
func (mr *MoveRegistry) GetDefinition(id string) (model.MoveDefinition, bool) {
	def, ok := mr.definitions[id]
	return def, ok
}

// ListDefinitions returns all registered move definitions.
func (mr *MoveRegistry) ListDefinitions() []model.MoveDefinition {
	defs := make([]model.MoveDefinition, 0, len(mr.definitions))
	for _, d := range mr.definitions {
		defs = append(defs, d)
	}
	return defs
}

// Execute runs a move by ID.
func (mr *MoveRegistry) Execute(r Roller, ch *model.Character, req model.MoveRequest, tracks []*model.ProgressTrack) (*model.MoveResult, error) {
	fn, ok := mr.executors[req.MoveID]
	if !ok {
		return nil, fmt.Errorf("unknown move: %s", req.MoveID)
	}
	return fn(r, ch, req, tracks)
}
