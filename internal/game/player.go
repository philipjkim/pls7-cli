package game

import "pls7-cli/pkg/poker"

// Player represents a single player in the game.
type Player struct {
	Name string
	Hand []poker.Card
}
