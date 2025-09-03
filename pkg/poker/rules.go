package poker

// HoleCardRules defines the rules governing the use of a player's private cards
// (hole cards) when forming a 5-card poker hand.
type HoleCardRules struct {
	// Count is the number of hole cards dealt to each player at the start of a hand.
	// For example, this is 2 for Texas Hold'em, 3 for PLS7, or 4 for Omaha.
	Count int `yaml:"count"`

	// UseConstraint specifies the rule for how many hole cards a player must or can use.
	// Valid options are:
	//  - "any": The player can use any number of their hole cards (from 0 to Count).
	//           This is typical for games like No-Limit Hold'em.
	//  - "exact": The player must use a specific number of hole cards, defined by UseCount.
	//             This is the rule in Omaha, where players must use exactly 2.
	//  - "max": The player can use up to a specific number of hole cards.
	UseConstraint string `yaml:"use_constraint"`

	// UseCount specifies the number of hole cards to be used when UseConstraint is
	// "exact" or "max". It is ignored if UseConstraint is "any".
	UseCount int `yaml:"use_count"`
}

// HandRankingsRules defines the hierarchy of poker hands for a game. It allows for
// both standard and custom hand rankings.
type HandRankingsRules struct {
	// UseStandardRankings, if true, enables the conventional poker hand hierarchy
	// (e.g., Royal Flush > Straight Flush > ... > High Card). If custom rankings
	// are also provided, they are inserted into this standard order.
	UseStandardRankings bool `yaml:"use_standard_rankings"`

	// CustomRankings is a list of non-standard hands to be added to the game's
	// ranking system. Each custom rank is defined by a name and its position
	// relative to another hand.
	CustomRankings []CustomHandRanking `yaml:"custom_rankings"`
}

// CustomHandRanking defines a non-standard poker hand and its position within the
// overall hand hierarchy.
type CustomHandRanking struct {
	// Name is the identifier for the custom hand, e.g., "skip_straight_flush".
	// This name must correspond to a hand evaluation logic in the engine.
	Name string `yaml:"name"`

	// InsertAfterRank specifies the hand immediately above this custom hand in the
	// hierarchy. For example, to make "skip_straight_flush" the second-best hand,
	// InsertAfterRank would be "royal_flush".
	InsertAfterRank string `yaml:"insert_after_rank"`
}

// LowHandRules defines the criteria for qualifying for the "low" half of the pot
// in a High-Low split game variant.
type LowHandRules struct {
	// Enabled, if true, signifies that the game is a High-Low split variant where
	// a low hand can win a portion of the pot.
	Enabled bool `yaml:"enabled"`

	// MaxRank specifies the maximum rank a card can have to be included in a low hand.
	// For example, in an "8-or-better" game, MaxRank would be 8. A qualifying low
	// hand consists of five unique cards with ranks at or below this value.
	MaxRank int `yaml:"max_rank"`
}

// GameRules is the top-level container for all the rules that define a specific
// poker game variant. This struct is typically populated by loading a YAML configuration
// file, allowing for flexible and dynamic game creation without changing the engine's code.
type GameRules struct {
	// Name is the full, human-readable name of the poker game variant,
	// e.g., "Pot-Limit Sampyeong 7-or-Better".
	Name string `yaml:"name"`

	// Abbreviation is the common short-form name for the game, e.g., "PLS7", "NLH".
	Abbreviation string `yaml:"abbreviation"`

	// BettingLimit defines the betting structure for the game.
	// Common values are "pot_limit", "no_limit", and "fixed_limit".
	BettingLimit string `yaml:"betting_limit"`

	// HoleCards defines the rules for the player's private cards.
	HoleCards HoleCardRules `yaml:"hole_cards"`
	// HandRankings defines the hierarchy of valid poker hands.
	HandRankings HandRankingsRules `yaml:"hand_rankings"`
	// LowHand defines the rules for the low hand in High-Low split games.
	LowHand LowHandRules `yaml:"low_hand"`
}
