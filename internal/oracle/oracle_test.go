package oracle

import (
	"testing"

	"github.com/ironsworn/ironsworn/internal/engine"
	"github.com/ironsworn/ironsworn/internal/model"
)

func TestAskYesNo(t *testing.T) {
	tests := []struct {
		name       string
		roll       int
		likelihood model.Likelihood
		wantYes    bool
	}{
		{"fifty_fifty yes", 30, model.LikelihoodFiftyFifty, true},
		{"fifty_fifty no", 60, model.LikelihoodFiftyFifty, false},
		{"almost_certain yes", 85, model.LikelihoodAlmostCertain, true},
		{"almost_certain no", 95, model.LikelihoodAlmostCertain, false},
		{"small_chance yes", 5, model.LikelihoodSmallChance, true},
		{"small_chance no", 20, model.LikelihoodSmallChance, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &engine.FixedRoller{D100Values: []int{tt.roll}}
			result := AskYesNo(r, tt.likelihood)
			if result.Answer != tt.wantYes {
				t.Errorf("roll %d with %s: got %v, want %v",
					tt.roll, tt.likelihood, result.Answer, tt.wantYes)
			}
		})
	}
}

func TestRollTable(t *testing.T) {
	r := &engine.FixedRoller{D100Values: []int{42}}
	result, err := RollTable(r, "action")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Roll != 42 {
		t.Errorf("expected roll 42, got %d", result.Roll)
	}
	if result.TableName != "Action" {
		t.Errorf("expected table name 'Action', got %q", result.TableName)
	}
	if result.Result != "Kill" {
		t.Errorf("expected 'Kill' for roll 42, got %q", result.Result)
	}
}

func TestRollTable_NotFound(t *testing.T) {
	r := &engine.FixedRoller{D100Values: []int{50}}
	_, err := RollTable(r, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent table")
	}
}

func TestListTables(t *testing.T) {
	tables := ListTables()
	if len(tables) == 0 {
		t.Error("expected tables")
	}

	// Verify order matches tableOrder
	for i, id := range tableOrder {
		if tables[i].ID != id {
			t.Errorf("position %d: expected %s, got %s", i, id, tables[i].ID)
		}
	}
}
