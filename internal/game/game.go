package game

import "pls7-cli/pkg/poker"

// Game represents the state of "a single hand of PLS7".
type Game struct {
	Players        []*Player
	CommunityCards []poker.Card
}
