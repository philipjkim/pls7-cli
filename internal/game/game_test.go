package game

import (
	"testing"
)

// TestHand_EliminatedPlayersAreSkipped tests that players with zero chips are properly excluded from a new hand.
func TestHand_EliminatedPlayersAreSkipped(t *testing.T) {
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	initialChips := 100000
	rules := &GameRules{
		HoleCards: HoleCardRules{
			Count: 3,
		},
	}
	g := NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false)

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
