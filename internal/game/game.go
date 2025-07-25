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
	PhaseHandOver // A new phase to signal the hand is finished
)

func (gp GamePhase) String() string {
	return []string{"Pre-Flop", "Flop", "Turn", "River", "Showdown", "Hand Over"}[gp]
}

// Game represents the state of a single hand of PLS7.
type Game struct {
	Players        []*Player
	Deck           *poker.Deck
	CommunityCards []poker.Card
	Pot            int
	DealerPos      int
	CurrentTurnPos int
	Phase          GamePhase
	BetToCall      int // Amount needed to call in the current round
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
