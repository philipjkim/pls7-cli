package engine

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// PlayerStatus defines the current state of a player within a single hand of poker.
type PlayerStatus int

// PlayerStatus constants represent the possible states a player can be in.
const (
	PlayerStatusPlaying    PlayerStatus = iota // PlayerStatusPlaying indicates the player is still actively participating in the hand.
	PlayerStatusFolded                         // PlayerStatusFolded indicates the player has folded and is no longer in the hand.
	PlayerStatusAllIn                          // PlayerStatusAllIn indicates the player has bet all their remaining chips.
	PlayerStatusEliminated                     // PlayerStatusEliminated indicates the player has run out of chips and is out of the game entirely.
)

// String returns the human-readable representation of a PlayerStatus.
// It implements the fmt.Stringer interface.
func (s PlayerStatus) String() string {
	switch s {
	case PlayerStatusPlaying:
		return "Playing"
	case PlayerStatusFolded:
		return "Folded"
	case PlayerStatusAllIn:
		return "All In"
	case PlayerStatusEliminated:
		return "Eliminated"
	default:
		return "Unknown"
	}
}

// AIProfile defines the behavioral characteristics and decision-making parameters
// for a CPU-controlled player. It allows for creating different "personalities"
// for AI opponents, from tight and passive to loose and aggressive.
type AIProfile struct {
	// Name is the identifier for the profile, e.g., "Tight-Aggressive".
	Name string
	// PlayHandThreshold is the minimum hand strength score required for the AI to
	// consider playing a hand pre-flop. A higher value means the AI is "tighter"
	// and plays fewer hands.
	PlayHandThreshold float64
	// RaiseHandThreshold is the minimum hand strength score required for the AI
	// to open with a raise pre-flop.
	RaiseHandThreshold float64
	// BluffingFrequency is the probability (0.0 to 1.0) that the AI will attempt
	// a bluff with a weak hand.
	BluffingFrequency float64
	// AggressionFactor is the probability (0.0 to 1.0) that the AI will choose
	// to bet or raise instead of check or call when it has a reasonably strong hand.
	AggressionFactor float64
	// MinRaiseMultiplier is the minimum multiplier for a raise amount, e.g., 2.0x the bet.
	MinRaiseMultiplier float64
	// MaxRaiseMultiplier is the maximum multiplier for a raise amount.
	MaxRaiseMultiplier float64
}

// Player represents a single participant in the poker game. It holds all state
// information relevant to the player, such as their cards, chip count, and status.
type Player struct {
	// Name is the unique identifier for the player.
	Name string
	// Hand holds the player's private hole cards.
	Hand []poker.Card
	// Chips is the player's current stack size.
	Chips int
	// CurrentBet is the amount of chips the player has committed to the pot in the
	// current betting round only. It is reset at the end of each round.
	CurrentBet int
	// TotalBetInHand is the cumulative amount of chips the player has put into the
	// pot throughout the entire current hand (across all betting rounds).
	TotalBetInHand int
	// Status indicates the player's current state in the hand (e.g., Playing, Folded).
	Status PlayerStatus
	// IsCPU is true if the player is controlled by the AI.
	IsCPU bool
	// LastActionDesc is a human-readable string describing the player's last action.
	LastActionDesc string
	// Profile contains the AI behavior parameters if the player is a CPU. It is nil for human players.
	Profile *AIProfile
	// Position is the player's seat at the table, represented by an index in the Game.Players slice.
	Position int
}

// String provides a formatted string representation of the Player's state,
// useful for debugging and logging.
func (p *Player) String() string {
	return fmt.Sprintf(
		"Player{Name: %s, Chips: %d, Status: %v, CurrentBet: %d, IsCPU: %t}",
		p.Name, p.Chips, p.Status, p.CurrentBet, p.IsCPU,
	)
}
