package game

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
}

// String makes PlayerStatus implement the Stringer interface.
func (p *Player) String() string {
	return fmt.Sprintf(
		"Player{Name: %s, Chips: %d, Status: %v, CurrentBet: %d, IsCPU: %t}",
		p.Name, p.Chips, p.Status, p.CurrentBet, p.IsCPU,
	)
}
