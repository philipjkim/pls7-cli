package game

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"testing"
)

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
	testCases := []struct {
		name           string
		seed           int64 // Each test case gets its own seed for deterministic results.
		difficulty     Difficulty
		betToCall      int
		playerBet      int
		handStrength   float64 // For Hard AI
		phase          GamePhase
		expectedAction ActionType
	}{
		// Seed 12345's first float is approx 0.515 (> 0.25 and > 0.20)
		{
			name:           "Easy AI - Folds on high random number",
			seed:           12345,
			difficulty:     DifficultyEasy,
			betToCall:      1000,
			playerBet:      0,
			expectedAction: ActionFold,
		},
		{
			name:           "Hard AI - No Bluff on high random number",
			seed:           12345,
			difficulty:     DifficultyHard,
			betToCall:      0,
			playerBet:      0,
			handStrength:   float64(poker.HighCard),
			phase:          PhaseFlop,
			expectedAction: ActionCheck,
		},
		// Seed 2's first float is approx 0.167 (< 0.25 and < 0.20)
		{
			name:           "Easy AI - Calls on low random number",
			seed:           2,
			difficulty:     DifficultyEasy,
			betToCall:      1000,
			playerBet:      0,
			expectedAction: ActionCall,
		},
		{
			name:           "Hard AI - Bluffs on low random number",
			seed:           2,
			difficulty:     DifficultyHard,
			betToCall:      0,
			playerBet:      0,
			handStrength:   float64(poker.HighCard),
			phase:          PhaseFlop,
			expectedAction: ActionBet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new, deterministic random generator for EACH test case.
			r := rand.New(rand.NewSource(tc.seed))

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
				// Mock the handEvaluator function field for this specific test run.
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
