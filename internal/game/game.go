package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

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
	Players         []*Player
	Deck            *poker.Deck
	CommunityCards  []poker.Card
	Pot             int
	DealerPos       int
	CurrentTurnPos  int
	Phase           GamePhase
	BetToCall       int
	LastRaiseAmount int
	HandCount       int
	Difficulty      Difficulty // To store the selected AI difficulty
	// handEvaluator is a function field to allow mocking in tests.
	handEvaluator func(g *Game, player *Player) float64
}

// NewGame initializes a new game with players and difficulty.
func NewGame(playerNames []string, initialChips int, difficulty Difficulty) *Game {
	players := make([]*Player, len(playerNames))
	for i, name := range playerNames {
		isCPU := (name != "YOU")
		players[i] = &Player{
			Name:  name,
			Chips: initialChips,
			IsCPU: isCPU,
		}
	}

	g := &Game{
		Players:    players,
		DealerPos:  -1,
		Difficulty: difficulty,
	}
	// Set the default hand evaluator.
	g.handEvaluator = evaluateHandStrength
	return g
}

// String returns a string representation of the game state.
func (g *Game) String() string {
	dealerName := "N/A"
	if g.DealerPos >= 0 && g.DealerPos < len(g.Players) {
		dealerName = g.Players[g.DealerPos].Name
	}

	turnPlayerName := "N/A"
	if g.CurrentTurnPos >= 0 && g.CurrentTurnPos < len(g.Players) {
		turnPlayerName = g.Players[g.CurrentTurnPos].Name
	}

	return fmt.Sprintf(
		"[Game State]:\n"+
			"- Hand #%d, Phase: %s, Difficulty: %v\n"+
			"- Pot: %d, BetToCall: %d, LastRaise: %d\n"+
			"- Dealer: %s, Turn: %s\n"+
			"- Community: %v\n"+
			"- Players: %+v\n",
		g.HandCount, g.Phase, g.Difficulty, g.Pot, g.BetToCall, g.LastRaiseAmount,
		dealerName, turnPlayerName, g.CommunityCards, g.Players,
	)
}
