package game

import (
	"testing"
)

// TestAwardPotToLastPlayer_SkipsEliminatedPlayers tests that the function correctly identifies
// the last non-folded player, skipping any players who were already eliminated.
func TestAwardPotToLastPlayer_SkipsEliminatedPlayers(t *testing.T) {
	// Scenario: 4 players. CPU 1 is eliminated. YOU and CPU 3 fold.
	// The winner must be CPU 2, not the eliminated CPU 1.
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	g := NewGame(playerNames, 10000, DifficultyMedium)

	// Setup the game state
	g.Pot = 1500
	g.Players[0].Status = PlayerStatusFolded     // YOU folds.
	g.Players[1].Status = PlayerStatusEliminated // CPU 1 is eliminated.
	g.Players[2].Status = PlayerStatusPlaying    // CPU 2 is the last one playing.
	g.Players[3].Status = PlayerStatusFolded     // CPU 3 folds.

	// Action: Award the pot.
	results := g.AwardPotToLastPlayer()

	// --- Assertion ---
	// This test will FAIL before the fix because the loop will find CPU 1 first
	// (since their status is not PlayerStatusFolded) and incorrectly award them the pot.
	if len(results) != 1 {
		t.Fatalf("Expected 1 winner, but got %d", len(results))
	}
	if results[0].PlayerName != "CPU 2" {
		t.Errorf("Expected winner to be 'CPU 2', but got '%s'", results[0].PlayerName)
	}
	if g.Players[2].Chips != 11500 { // 10000 initial + 1500 pot
		t.Errorf("Expected CPU 2's chips to be 11500, but got %d", g.Players[2].Chips)
	}
}
