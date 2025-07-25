package game

import "pls7-cli/pkg/poker"

// PlayerStatus defines the current state of a player in a hand.
type PlayerStatus int

const (
	PlayerStatusPlaying PlayerStatus = iota // Still in the hand
	PlayerStatusFolded                      // Folded the hand
	PlayerStatusAllIn                       // All-in
)

// Player represents a single player in the game.
type Player struct {
	Name       string
	Hand       []poker.Card
	Chips      int          // Total chips the player has
	CurrentBet int          // Chips bet in the current round
	Status     PlayerStatus // Current status (Playing, Folded, AllIn)
	IsCPU      bool         // Flag to distinguish CPU from human player
}
