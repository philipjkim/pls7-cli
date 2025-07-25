package game

import "pls7-cli/pkg/poker"

// GamePhase defines the current phase of the game.
//
//goland:noinspection GoNameStartsWithPackageName
type GamePhase int

const (
	PhasePreFlop GamePhase = iota
	PhaseFlop
	PhaseTurn
	PhaseRiver
	PhaseShowdown
)

// Game represents the state of a single hand of PLS7.
type Game struct {
	Players        []*Player
	Deck           *poker.Deck
	CommunityCards []poker.Card
	Pot            int       // Total chips in the pot
	DealerPos      int       // Index of the dealer in the Players slice
	CurrentTurnPos int       // Index of the player whose turn it is
	Phase          GamePhase // Current game phase
}

// NewGame initializes a new game with players.
func NewGame(playerNames []string, initialChips int) *Game {
	players := make([]*Player, len(playerNames))
	for i, name := range playerNames {
		isCPU := name != "YOU"
		players[i] = &Player{
			Name:  name,
			Chips: initialChips,
			IsCPU: isCPU,
		}
	}

	return &Game{
		Players:   players,
		DealerPos: -1, // Will be set to 0 on the first hand
	}
}
