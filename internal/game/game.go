package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
	"strings"
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
	LowlessMode   bool // Flag to indicate if the game is in lowless mode
	ShowsOuts     bool // Flag to indicate if outs should be shown (if DevMode is true, this is always true)
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
	isDev bool,
	isLowless bool,
	showsOuts bool,
) *Game {
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
		Players:     players,
		DealerPos:   -1,
		Difficulty:  difficulty,
		DevMode:     isDev,
		LowlessMode: isLowless,
		ShowsOuts:   showsOuts,
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

// cardsFromStrings is a helper function to make creating cards in tests easier.
func cardsFromStrings(s string) []poker.Card {
	if s == "" {
		return []poker.Card{}
	}
	parts := strings.Split(s, " ")
	cards := make([]poker.Card, len(parts))
	rankMap := map[rune]poker.Rank{
		'2': poker.Two, '3': poker.Three, '4': poker.Four, '5': poker.Five, '6': poker.Six, '7': poker.Seven,
		'8': poker.Eight, '9': poker.Nine, 'T': poker.Ten, 'J': poker.Jack, 'Q': poker.Queen, 'K': poker.King, 'A': poker.Ace,
	}
	suitMap := map[rune]poker.Suit{
		's': poker.Spade, 'h': poker.Heart, 'd': poker.Diamond, 'c': poker.Club,
	}
	for i, part := range parts {
		rank := rankMap[rune(part[0])]
		suit := suitMap[rune(part[1])]
		cards[i] = poker.Card{Rank: rank, Suit: suit}
	}
	return cards
}
