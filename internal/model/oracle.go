package model

// Likelihood represents the probability weighting for yes/no oracle questions.
type Likelihood string

const (
	LikelihoodAlmostCertain Likelihood = "almost_certain"
	LikelihoodLikely        Likelihood = "likely"
	LikelihoodFiftyFifty    Likelihood = "fifty_fifty"
	LikelihoodUnlikely      Likelihood = "unlikely"
	LikelihoodSmallChance   Likelihood = "small_chance"
)

// Threshold returns the d100 threshold at or below which the answer is "yes".
func (l Likelihood) Threshold() int {
	switch l {
	case LikelihoodAlmostCertain:
		return 90
	case LikelihoodLikely:
		return 75
	case LikelihoodFiftyFifty:
		return 50
	case LikelihoodUnlikely:
		return 25
	case LikelihoodSmallChance:
		return 10
	default:
		return 50
	}
}

// YesNoResult represents the result of an Ask the Oracle question.
type YesNoResult struct {
	Roll       int        `json:"roll"`
	Likelihood Likelihood `json:"likelihood"`
	Threshold  int        `json:"threshold"`
	Answer     bool       `json:"answer"` // true = yes
}

// OracleTableEntry is a single row in an oracle table.
type OracleTableEntry struct {
	Low    int    `json:"low"`
	High   int    `json:"high"`
	Result string `json:"result"`
}

// OracleTable represents a rollable oracle table.
type OracleTable struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Category string             `json:"category"`
	Entries  []OracleTableEntry `json:"entries"`
}
