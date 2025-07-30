package game

import (
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"strings"
	"testing"
)

// TestAwardPotToLastPlayer_SkipsEliminatedPlayers tests that the function correctly identifies
// the last non-folded player, skipping any players who were already eliminated.
func TestAwardPotToLastPlayer_SkipsEliminatedPlayers(t *testing.T) {
	// Scenario: 4 players. CPU 1 is eliminated. YOU and CPU 3 fold.
	// The winner must be CPU 2, not the eliminated CPU 1.
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	g := NewGame(playerNames, 10000, DifficultyMedium, true, false)

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

// TestDistributePot_SidePots tests the pot distribution logic with multiple all-in players,
// which should create side pots.
func TestDistributePot_SidePots(t *testing.T) {
	util.InitLogger(true)

	// Scenario: 3 players go all-in with different stack sizes.
	// ShortStack (2000) has the best hand.
	// MidStack (5000) has the second best hand.
	// BigStack (10000) has the worst hand.
	// No low hands qualify.
	playerNames := []string{"ShortStack", "MidStack", "BigStack"}
	g := NewGame(playerNames, 0, DifficultyMedium, true, false)

	// Setup player states
	g.Players[0].Chips = 0
	g.Players[0].TotalBetInHand = 2000
	g.Players[0].Status = PlayerStatusAllIn
	g.Players[0].Hand = cardsFromStrings("As Ac Ad Ah Ks") // Four of a Kind (best hand)

	g.Players[1].Chips = 0
	g.Players[1].TotalBetInHand = 5000
	g.Players[1].Status = PlayerStatusAllIn
	g.Players[1].Hand = cardsFromStrings("Qs Qc Qd Jh Js") // Full House (second best)

	g.Players[2].Chips = 0
	g.Players[2].TotalBetInHand = 10000
	g.Players[2].Status = PlayerStatusAllIn
	g.Players[2].Hand = cardsFromStrings("Ts 9c 8d 7h 6s") // Straight (worst hand)

	// Total pot is the sum of all bets
	g.Pot = 2000 + 5000 + 10000

	// Action: Distribute the pot
	results := g.DistributePot()

	// --- Assertions ---
	// Expected distribution:
	// Main Pot (2000 * 3 = 6000) goes to ShortStack.
	// Side Pot 1 ((5000-2000) * 2 = 6000) goes to MidStack.
	// Side Pot 2 ((10000-5000) * 1 = 5000) is returned to BigStack.

	if len(results) != 3 {
		t.Fatalf("Expected 3 distribution results, but got %d", len(results))
	}

	// Check chip distribution
	if g.Players[0].Chips != 6000 {
		t.Errorf("Expected ShortStack to win 6000, but got %d", g.Players[0].Chips)
	}
	if g.Players[1].Chips != 6000 {
		t.Errorf("Expected MidStack to win 6000, but got %d", g.Players[1].Chips)
	}
	if g.Players[2].Chips != 5000 {
		t.Errorf("Expected BigStack to get back 5000, but got %d", g.Players[2].Chips)
	}
}
