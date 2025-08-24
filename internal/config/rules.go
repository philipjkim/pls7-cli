package config

import (
	"fmt"
	os "os"

	"gopkg.in/yaml.v3"
)

// HoleCardRules defines the rules for the player's private cards.
type HoleCardRules struct {
	// Count is the number of hole cards dealt to each player.
	// e.g., 2 for NLH, 3 for PLS/PLS7, 4 for PLO/PLO8
	Count int `yaml:"count"`

	// UseConstraint specifies how many hole cards must be used to form a 5-card hand.
	// Possible values:
	//   - "any": Players can use any number of their hole cards (0, 1, 2, etc.). (e.g., NLH, PLS, PLS7)
	//   - "exact": Players must use a specific number of hole cards. (e.g., PLO, PLO8)
	//   - "max": Players can use up to a specific number of hole cards.
	UseConstraint string `yaml:"use_constraint"`

	// UseCount is the number associated with UseConstraint.
	// It's 0 for "any", and the specific number for "exact" or "max".
	// e.g., 2 for PLO's "exact" constraint.
	UseCount int `yaml:"use_count"`
}

// HandRankingsRules lists the poker hands that are valid in this game, ordered from strongest to weakest.
type HandRankingsRules struct {
	// UseStandardRankings specifies whether to use standard poker hand rankings.
	// If true, the game uses the universally recognized poker hand hierarchy.
	// If false, custom_rankings will be used to define additional or modified rankings.
	UseStandardRankings bool `yaml:"use_standard_rankings"`

	// CustomRankings defines a list of custom hand rankings and their insertion points
	// within the standard hierarchy (if UseStandardRankings is true) or as a complete
	// custom hierarchy (if UseStandardRankings is false, though this is not fully supported yet).
	CustomRankings []CustomHandRanking `yaml:"custom_rankings"`
}

// CustomHandRanking defines a custom poker hand rank and where it should be inserted.
type CustomHandRanking struct {
	// Name is the name of the custom hand ranking (e.g., "skip_straight_flush").
	Name string `yaml:"name"`

	// InsertAfterRank specifies the name of the standard or custom hand ranking
	// after which this custom ranking should be inserted in the hierarchy.
	// (e.g., "royal_flush" means this custom rank is just below royal_flush).
	InsertAfterRank string `yaml:"insert_after_rank"`
}

// LowHandRules defines the rules for the low hand portion of the game (for Hi-Lo variants).
type LowHandRules struct {
	// Enabled specifies if the game includes a low hand.
	Enabled bool `yaml:"enabled"`

	// MaxRank is the highest rank a card can be to qualify for a low hand.
	// e.g., 7 for PLS7 (7-or-better), 8 for PLO8 (8-or-better).
	MaxRank int `yaml:"max_rank"`
}

// GameRules defines all the rules for a specific poker game variant.
// It is loaded from a YAML file to allow for flexible game creation.
type GameRules struct {
	// Name is the full, human-readable name of the poker game.
	// e.g., "Pot-Limit Sampyeong 7-or-Better", "No-Limit Texas Hold'em"
	Name string `yaml:"name"`

	// Abbreviation is the short-form name for the game.
	// e.g., "PLS7", "NLH", "PLO"
	Abbreviation string `yaml:"abbreviation"`

	// BettingLimit specifies the betting structure.
	// Possible values: "pot_limit", "no_limit", "fixed_limit"
	BettingLimit string `yaml:"betting_limit"`

	HoleCards    HoleCardRules     `yaml:"hole_cards"`
	HandRankings HandRankingsRules `yaml:"hand_rankings"`
	LowHand      LowHandRules      `yaml:"low_hand"`
}

// LoadGameRulesFromFile reads a YAML file from the given path and returns a GameRules struct.
func LoadGameRulesFromFile(filePath string) (*GameRules, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rules GameRules
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

// LoadGameRulesFromOptions loads game rules from a YAML string by option value.
// - Available ruleStr: "pls", "pls7"
func LoadGameRulesFromOptions(ruleStr string) (*GameRules, error) {
	filePath := fmt.Sprintf("rules/%s.yml", ruleStr)
	return LoadGameRulesFromFile(filePath)
}
