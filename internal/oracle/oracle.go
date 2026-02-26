package oracle

import (
	"fmt"

	"github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

// AskYesNo rolls on the oracle for a yes/no question with the given likelihood.
func AskYesNo(r engine.Roller, likelihood model.Likelihood) model.YesNoResult {
	roll := r.D100()
	threshold := likelihood.Threshold()
	return model.YesNoResult{
		Roll:       roll,
		Likelihood: likelihood,
		Threshold:  threshold,
		Answer:     roll <= threshold,
	}
}

// RollTable rolls on a named oracle table and returns the result.
func RollTable(r engine.Roller, tableID string) (*model.OracleRollResult, error) {
	table, ok := tables[tableID]
	if !ok {
		return nil, fmt.Errorf("oracle table not found: %s", tableID)
	}

	roll := r.D100()
	result := lookupResult(table, roll)

	return &model.OracleRollResult{
		Roll:      roll,
		TableID:   table.ID,
		TableName: table.Name,
		Result:    result,
	}, nil
}

// ListTables returns summary info for all available oracle tables.
func ListTables() []model.OracleTable {
	result := make([]model.OracleTable, 0, len(tables))
	for _, t := range tableOrder {
		result = append(result, tables[t])
	}
	return result
}

func lookupResult(table model.OracleTable, roll int) string {
	for _, entry := range table.Entries {
		if roll >= entry.Low && roll <= entry.High {
			return entry.Result
		}
	}
	return "Unknown result"
}
