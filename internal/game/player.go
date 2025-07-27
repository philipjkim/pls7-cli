package game

import "pls7-cli/pkg/poker"

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
	Name       string
	Hand       []poker.Card
	Chips      int
	CurrentBet int
	Status     PlayerStatus
	IsCPU      bool
}
