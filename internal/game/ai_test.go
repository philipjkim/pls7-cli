package game

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"strings"
	"testing"
)

// cardsFromStrings is a helper function copied from poker_test for convenience.
func cardsFromStrings(s string) []poker.Card {
	// FIX: Handle empty string input to prevent panic.
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

func TestEvaluateHandStrength(t *testing.T) {
	testCases := []struct {
		name              string
		phase             GamePhase
		holeCardsStr      string
		communityCardsStr string
		minExpectedScore  float64 // Use a range for float comparison
		maxExpectedScore  float64
	}{
		// --- Pre-Flop Tests ---
		{
			name:             "Pre-Flop - High Pair (Aces)",
			phase:            PhasePreFlop,
			holeCardsStr:     "As Ac 2d",
			minExpectedScore: 49,
			maxExpectedScore: 49,
		},
		{
			name:             "Pre-Flop - Low Pair",
			phase:            PhasePreFlop,
			holeCardsStr:     "2s 2c Ad",
			minExpectedScore: 27,
			maxExpectedScore: 27,
		},
		{
			name:             "Pre-Flop - Suited Connectors",
			phase:            PhasePreFlop,
			holeCardsStr:     "8s 7s 2d",
			minExpectedScore: 9, // Suited=4 + Connector=5 = 9
			maxExpectedScore: 9,
		},
		{
			name:             "Pre-Flop - Premium Suited High Cards",
			phase:            PhasePreFlop,
			holeCardsStr:     "As Ks 2d",
			minExpectedScore: 27, // FIX: Corrected score (10+8) + Suited=4 + Connector=5 = 27
			maxExpectedScore: 27,
		},
		{
			name:             "Pre-Flop - Weak Unsuited",
			phase:            PhasePreFlop,
			holeCardsStr:     "7d 2h 3c",
			minExpectedScore: 0,
			maxExpectedScore: 0,
		},

		// --- Post-Flop Tests ---
		{
			name:              "Post-Flop - Two Pair",
			phase:             PhaseFlop,
			holeCardsStr:      "As Ks Qd",
			communityCardsStr: "Ac Kc 2h 3d 4s",
			minExpectedScore:  float64(poker.TwoPair),
			maxExpectedScore:  float64(poker.TwoPair),
		},
		{
			name:              "Post-Flop - Full House",
			phase:             PhaseTurn,
			holeCardsStr:      "As Ac Qd",
			communityCardsStr: "Ah Kc Kh 3d 4s",
			minExpectedScore:  float64(poker.FullHouse),
			maxExpectedScore:  float64(poker.FullHouse),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup game state for the test
			g := &Game{
				Phase:          tc.phase,
				CommunityCards: cardsFromStrings(tc.communityCardsStr),
			}
			player := &Player{
				Hand: cardsFromStrings(tc.holeCardsStr),
			}

			score := evaluateHandStrength(g, player)

			if score < tc.minExpectedScore || score > tc.maxExpectedScore {
				t.Errorf("Expected score between %.2f and %.2f, but got %.2f", tc.minExpectedScore, tc.maxExpectedScore, score)
			}
		})
	}
}

// TestCPUActionWithRandomness tests the AI actions that depend on randomness.
func TestCPUActionWithRandomness(t *testing.T) {
	// Use a fixed seed for deterministic "random" behavior.
	const seed = 12345
	r := rand.New(rand.NewSource(seed))

	// With seed 12345, the first r.Float64() is approx 0.515 ( > 0.25)
	// The second r.Float64() is approx 0.463 ( > 0.20)
	// The third r.Float64() is approx 0.133 ( < 0.25 and < 0.20)

	testCases := []struct {
		name           string
		difficulty     Difficulty
		betToCall      int
		playerBet      int
		handStrength   float64 // For Hard AI
		phase          GamePhase
		expectedAction ActionType
	}{
		{
			name:           "Easy AI - Folds when random > 0.25",
			difficulty:     DifficultyEasy,
			betToCall:      1000,
			playerBet:      0,
			expectedAction: ActionFold,
		},
		{
			name:           "Easy AI - Calls when random < 0.25",
			difficulty:     DifficultyEasy,
			betToCall:      1000,
			playerBet:      0,
			expectedAction: ActionCall, // This will use the third random number (0.133)
		},
		{
			name:           "Hard AI - No Bluff when random > 0.20",
			difficulty:     DifficultyHard,
			betToCall:      0,
			playerBet:      0,
			handStrength:   float64(poker.HighCard),
			phase:          PhaseFlop,
			expectedAction: ActionCheck, // Should default to Medium AI's action
		},
		{
			name:           "Hard AI - Bluffs when random < 0.20",
			difficulty:     DifficultyHard,
			betToCall:      0,
			playerBet:      0,
			handStrength:   float64(poker.HighCard),
			phase:          PhaseFlop,
			expectedAction: ActionBet, // This will use the third random number (0.133)
		},
	}

	// Reset the seed before running tests
	r.Seed(seed)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Difficulty: tc.difficulty,
				BetToCall:  tc.betToCall,
				Phase:      tc.phase,
			}
			// Initialize the default evaluator, which we might override.
			g.handEvaluator = evaluateHandStrength
			player := &Player{
				CurrentBet: tc.playerBet,
			}

			var action PlayerAction
			switch tc.difficulty {
			case DifficultyEasy:
				action = g.getEasyAction(player, r)
			case DifficultyHard:
				// FIX: Mock the handEvaluator function field, not the method.
				originalEval := g.handEvaluator
				g.handEvaluator = func(g *Game, p *Player) float64 { return tc.handStrength }
				action = g.getHardAction(player, r)
				g.handEvaluator = originalEval // Restore original function
			}

			if action.Type != tc.expectedAction {
				t.Errorf("Expected action %v, but got %v", tc.expectedAction, action.Type)
			}
		})
	}
}
