package engine

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// PlayerStatus defines the current state of a player in a hand.
type PlayerStatus int

const (
	PlayerStatusPlaying    PlayerStatus = iota // Still in the hand
	PlayerStatusFolded                         // Folded the hand
	PlayerStatusAllIn                          // All-in
	PlayerStatusEliminated                     // Out of chips and out of the game
)

// String makes PlayerStatus implement the Stringer interface.
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

// AIProfile defines the behavioral characteristics of a CPU player.
type AIProfile struct {
	Name               string
	PlayHandThreshold  float64 // Minimum hand strength to play (0-100). Higher is tighter.
	RaiseHandThreshold float64 // Minimum hand strength to raise.
	BluffingFrequency  float64 // Chance to bluff (0.0 to 1.0).
	AggressionFactor   float64 // How likely to bet/raise vs. check/call.
	MinRaiseMultiplier float64 // Minimum multiplier for a raise (e.g., 2x the call amount).
	MaxRaiseMultiplier float64 // Maximum multiplier for a raise.
}

// Player represents a single player in the game.
type Player struct {
	Name           string
	Hand           []poker.Card
	Chips          int
	CurrentBet     int
	TotalBetInHand int // New field to track total chips contributed to the pot in the current hand
	Status         PlayerStatus
	IsCPU          bool
	LastActionDesc string // Describes the last action taken in the round
	Profile        *AIProfile
	Position       int
}

// String makes PlayerStatus implement the Stringer interface.
func (p *Player) String() string {
	return fmt.Sprintf(
		"Player{Name: %s, Chips: %d, Status: %v, CurrentBet: %d, IsCPU: %t}",
		p.Name, p.Chips, p.Status, p.CurrentBet, p.IsCPU,
	)
}
