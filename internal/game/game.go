package game

import (
	"fmt"
	"math/rand"
	"os"
	"pls7-cli/pkg/poker"
	"time"

	"github.com/sirupsen/logrus"
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
	handEvaluator     func(g *Game, player *Player) float64
	DevMode           bool // Flag to indicate if the game is in development mode
	ShowsOuts         bool // Flag to indicate if outs should be shown (if DevMode is true, this is always true)
	Rules             *poker.GameRules
	Rand              *rand.Rand // Centralized random number generator
	BlindUpInterval   int
	BettingCalculator BettingLimitCalculator
	Aggressor         *Player
	ActionCloserPos   int
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
	rules *poker.GameRules,
	isDev bool,
	showsOuts bool,
	blindUpInterval int,
) *Game {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a single random source for the game
	players := make([]*Player, len(playerNames))
	cpuProfilesToAssign, err := cpuProfiles(difficulty, len(playerNames)-1)
	if err != nil {
		logrus.Errorf("Failed to get CPU profiles: %v", err)
		os.Exit(1)
	}

	// playerNames: 1 human + n CPUs
	// cpuProfilesToAssign: n CPU profiles based on difficulty
	if len(playerNames)-1 != len(cpuProfilesToAssign) {
		logrus.Errorf(
			"Mismatch in number of CPU profiles and players. %d != %d - 1",
			len(cpuProfilesToAssign), len(playerNames),
		)
		os.Exit(1)
	}

	for i, name := range playerNames {
		isCPU := (name != "YOU")
		players[i] = &Player{
			Name:     name,
			Chips:    initialChips,
			IsCPU:    isCPU,
			Position: i,
		}

		if isCPU {
			if profile, ok := aiProfiles[cpuProfilesToAssign[i-1]]; ok {
				players[i].Profile = &profile
			} else {
				logrus.Errorf("Unknown AI profile: %s", cpuProfilesToAssign[i-1])
				os.Exit(1)
			}
		}
	}

	var calculator BettingLimitCalculator
	switch rules.BettingLimit {
	case "pot_limit":
		calculator = &PotLimitCalculator{}
	case "no_limit":
		calculator = &NoLimitCalculator{}
	default:
		logrus.Fatalf("Unknown betting limit type: %s", rules.BettingLimit)
	}

	g := &Game{
		Players:           players,
		DealerPos:         -1,
		Difficulty:        difficulty,
		DevMode:           isDev,
		ShowsOuts:         showsOuts,
		Rules:             rules,
		Rand:              r,
		BlindUpInterval:   blindUpInterval,
		BettingCalculator: calculator,
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

// CalculateBettingLimits delegates the calculation to the game's betting calculator.
func (g *Game) CalculateBettingLimits() (minRaiseTotal int, maxRaiseTotal int) {
	return g.BettingCalculator.CalculateBettingLimits(g)
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

// cpuProfiles returns a list of CPU profiles based on the game difficulty.
// numCPUs should be between 1 and 5, inclusive.
func cpuProfiles(difficulty Difficulty, numCPUs int) ([]string, error) {
	if numCPUs < 1 || numCPUs > 5 {
		return []string{}, fmt.Errorf("numCPUs must be between 1 and 5, got %d", numCPUs)
	}

	switch difficulty {
	case DifficultyEasy:
		return []string{
			"Loose-Passive", "Loose-Passive",
			"Loose-Passive", "Loose-Passive", "Loose-Passive",
		}[:numCPUs], nil
	case DifficultyMedium:
		return []string{
			"Loose-Passive", "Loose-Passive",
			"Tight-Passive", "Tight-Passive", "Tight-Passive",
		}[:numCPUs], nil
	case DifficultyHard:
		return []string{
			"Tight-Passive",
			"Loose-Aggressive", "Loose-Aggressive",
			"Tight-Aggressive", "Tight-Aggressive",
		}[:numCPUs], nil
	default:
		return []string{}, fmt.Errorf("unknown difficulty: %v", difficulty)
	}
}
