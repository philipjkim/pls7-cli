package engine

import (
	"pls7-cli/internal/config"
	"reflect"
	"testing"
)

// TestHand_EliminatedPlayersAreSkipped tests that players with zero chips are properly excluded from a new hand.
func TestHand_EliminatedPlayersAreSkipped(t *testing.T) {
	playerNames := []string{"YOU", "CPU1", "CPU2", "CPU3"}
	initialChips := 100000
	rules, err := config.LoadGameRulesFromFile("../../rules/pls7.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame(playerNames, initialChips, 500, 1000, DifficultyMedium, rules, true, false, 0)

	// Manually eliminate two players
	g.Players[1].Chips = 0
	g.Players[1].Status = PlayerStatusEliminated
	g.Players[3].Chips = 0
	g.Players[3].Status = PlayerStatusEliminated

	// Start a new hand
	g.StartNewHand()

	// --- Assertion 1: Eliminated players should not have been dealt cards. ---
	// This test will FAIL before the fix.
	if len(g.Players[1].Hand) != 0 {
		t.Errorf("Expected eliminated player CPU 1 to have 0 cards, but got %d", len(g.Players[1].Hand))
	}
	if len(g.Players[3].Hand) != 0 {
		t.Errorf("Expected eliminated player CPU 3 to have 0 cards, but got %d", len(g.Players[3].Hand))
	}

	// --- Assertion 2: Active players should have been dealt cards. ---
	if len(g.Players[0].Hand) != 3 {
		t.Errorf("Expected active player YOU to have 3 cards, but got %d", len(g.Players[0].Hand))
	}
	if len(g.Players[2].Hand) != 3 {
		t.Errorf("Expected active player CPU 2 to have 3 cards, but got %d", len(g.Players[2].Hand))
	}
}

func TestNewGame_AssignsCorrectCalculator(t *testing.T) {
	testCases := []struct {
		name               string
		ruleStr            string
		expectedCalculator interface{}
	}{
		{
			name:               "Pot Limit Game",
			ruleStr:            "pls7",
			expectedCalculator: &PotLimitCalculator{},
		},
		{
			name:               "No Limit Game",
			ruleStr:            "nlh",
			expectedCalculator: &NoLimitCalculator{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := config.LoadGameRulesFromFile("../../rules/" + tc.ruleStr + ".yml")
			if err != nil {
				t.Fatalf("Failed to load game rules: %v", err)
			}
			g := NewGame([]string{"YOU", "CPU1"}, 1000, 500, 1000, DifficultyEasy, rules, false, false, 0)

			if g.BettingCalculator == nil {
				t.Fatal("g.BettingCalculator is nil")
			}

			actualType := reflect.TypeOf(g.BettingCalculator)
			expectedType := reflect.TypeOf(tc.expectedCalculator)

			if actualType != expectedType {
				t.Errorf("Expected calculator of type %v, but got %v", expectedType, actualType)
			}
		})
	}
}
