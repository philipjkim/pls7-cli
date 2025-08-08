package game

import (
	"fmt"
	"math/rand"
	"pls7-cli/internal/config"
	"pls7-cli/pkg/poker"
	"time"
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
	DevMode       bool // Flag to indicate if the game is in development mode
	ShowsOuts     bool // Flag to indicate if outs should be shown (if DevMode is true, this is always true)
	Rules         *config.GameRules
	Rand          *rand.Rand // Centralized random number generator
}

// CPUThinkTime returns the delay for CPU actions based on the development mode.
func (g *Game) CPUThinkTime() time.Duration {
	if g.DevMode {
		return 0 // No delay in dev mode
	}
	return 500 * time.Millisecond // Default delay for CPU thinking
}

// NewGame initializes a new game with players and difficulty.
func NewGame(
	playerNames []string,
	initialChips int,
	difficulty Difficulty,
	rules *config.GameRules,
	isDev bool,
	showsOuts bool,
) *Game {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a single random source for the game
	players := make([]*Player, len(playerNames))
	for i, name := range playerNames {
		isCPU := (name != "YOU")
		players[i] = &Player{
			Name:     name,
			Chips:    initialChips,
			IsCPU:    isCPU,
			Position: i,
		}
		// Assign a random profile to CPU players
		if isCPU {
			assignRandomProfile(players[i], r)
		}
	}

	g := &Game{
		Players:    players,
		DealerPos:  -1,
		Difficulty: difficulty,
		DevMode:    isDev,
		ShowsOuts:  showsOuts,
		Rules:      rules,
		Rand:       r,
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

// CanShowOuts checks if the outs can be shown for a player based on the game state.
func (g *Game) CanShowOuts(p *Player) bool {
	humanPlayerInPlay := p.Name == "YOU" && p.Status != PlayerStatusFolded
	availablePhase := g.Phase == PhaseFlop || g.Phase == PhaseTurn
	optionEnabled := g.DevMode || g.ShowsOuts
	return humanPlayerInPlay && optionEnabled && availablePhase
}

// minRaiseAmount calculates the minimum amount for a valid raise.
func (g *Game) minRaiseAmount() int {
	minRaiseIncrease := g.LastRaiseAmount
	if minRaiseIncrease == 0 {
		minRaiseIncrease = g.BetToCall
	}
	if g.BetToCall == 0 {
		minRaiseIncrease = BigBlindAmt
	}
	return g.BetToCall + minRaiseIncrease
}
