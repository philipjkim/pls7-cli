package engine

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

// Game represents the central state of a poker game. It manages all aspects of a single hand,
// including players, the deck, community cards, and the betting rounds. It is the primary
// orchestrator for the game logic.
type Game struct {
	// Players is a slice of pointers to Player objects participating in the game.
	// The order of players in this slice determines their seating position at the table.
	Players []*Player
	// Deck is the deck of cards used for the game. It is shuffled at the beginning of each hand.
	Deck *poker.Deck
	// CommunityCards are the shared cards on the board, used by all players to make their hands.
	CommunityCards []poker.Card
	// Pot is the total amount of chips that has been bet in the current hand.
	Pot int
	// DealerPos is the index of the player who is the dealer in the current hand.
	// The dealer button moves to the next active player after each hand.
	DealerPos int
	// CurrentTurnPos is the index of the player whose turn it is to act.
	CurrentTurnPos int
	// Phase represents the current stage of the hand (e.g., Pre-Flop, Flop, Turn, River).
	Phase GamePhase
	// BetToCall is the amount that a player must bet to stay in the hand.
	// It is the highest bet made so far in the current betting round.
	BetToCall int
	// LastRaiseAmount stores the size of the last raise made in the current betting round.
	// This is used to calculate the minimum valid raise amount for the next player.
	LastRaiseAmount int
	// HandCount tracks the number of hands that have been played in the current game session.
	HandCount int
	// SmallBlind is the current small blind amount.
	SmallBlind int
	// BigBlind is the current big blind amount.
	BigBlind int
	// Difficulty stores the selected AI difficulty level for CPU players.
	Difficulty Difficulty
	// handEvaluator is a function field that allows for mocking hand evaluation logic in tests.
	// In a normal game, it points to the `evaluateHandStrength` function.
	handEvaluator func(g *Game, player *Player) float64
	// DevMode is a boolean flag that enables or disables development-specific features,
	// such as detailed logging and predictable card dealing.
	DevMode bool
	// ShowsOuts is a boolean flag that, when enabled, shows the player their potential "outs"
	// (cards that could improve their hand) during the game. This is primarily a debugging
	// and learning tool.
	ShowsOuts bool
	// Rules is a pointer to a GameRules struct, which defines the specific variant of poker being played
	// (e.g., No-Limit Hold'em, Pot-Limit Sampyeong). The rules are loaded from a YAML file.
	Rules *poker.GameRules
	// Rand is a centralized random number generator for the game, ensuring that all random events
	// (like shuffling and CPU decisions) are derived from a single source.
	Rand *rand.Rand
	// BlindUpInterval is the number of hands after which the small and big blinds double.
	// A value of 0 disables this feature.
	BlindUpInterval int
	// BettingCalculator is an interface that provides the logic for calculating betting limits
	// based on the game's rules (e.g., Pot-Limit, No-Limit).
	BettingCalculator BettingLimitCalculator
	// Aggressor is a pointer to the player who made the last aggressive action (bet or raise)
	// in the current betting round. It is used to determine when the action is closed.
	Aggressor *Player
	// ActionCloserPos is the position of the player who closes the betting action if there are no raises.
	// For pre-flop, this is the Big Blind. For post-flop, it's the player to the left of the dealer.
	ActionCloserPos int
	// ActionsTakenThisRound counts the number of actions (fold, check, call, bet, raise) taken
	// in the current betting round. This is used to determine when the round is over.
	ActionsTakenThisRound int
	// TotalInitialChips is the sum of all players' initial chip counts. This is used for sanity checks
	// to ensure that chips are not being created or destroyed during the game.
	TotalInitialChips int
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
	smallBlind int,
	bigBlind int,
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
		SmallBlind:        smallBlind,
		BigBlind:          bigBlind,
		Difficulty:        difficulty,
		DevMode:           isDev,
		ShowsOuts:         showsOuts,
		Rules:             rules,
		Rand:              r,
		BlindUpInterval:   blindUpInterval,
		BettingCalculator: calculator,
		TotalInitialChips: initialChips * len(playerNames),
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
		minRaiseIncrease = g.BigBlind
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
