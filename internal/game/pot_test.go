package game

import (
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"testing"
)

// TestAwardPotToLastPlayer_SkipsEliminatedPlayers tests that the function correctly identifies
// the last non-folded player, skipping any players who were already eliminated.
func TestAwardPotToLastPlayer_SkipsEliminatedPlayers(t *testing.T) {
	// Scenario: 4 players. CPU 1 is eliminated. YOU and CPU 3 fold.
	// The winner must be CPU 2, not the eliminated CPU 1.
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	rules := &config.GameRules{
		HoleCards: config.HoleCardRules{
			Count: 3,
		},
		LowHand: config.LowHandRules{Enabled: false},
	}
	g := NewGame(playerNames, 10000, DifficultyMedium, rules, true, false)

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
	rules := &config.GameRules{
		HoleCards: config.HoleCardRules{
			Count: 3,
		},
		LowHand: config.LowHandRules{Enabled: false},
		HandRankings: config.HandRankingsRules{
			UseStandardRankings: true,
			CustomRankings:      []config.CustomHandRanking{},
		},
	}
	g := NewGame(playerNames, 0, DifficultyMedium, rules, true, false)

	// Setup player states
	g.Players[0].Chips = 0
	g.Players[0].TotalBetInHand = 2000
	g.Players[0].Status = PlayerStatusAllIn
	g.Players[0].Hand = poker.CardsFromStrings("As Ac Ad Ah Ks") // Four of a Kind (best hand)

	g.Players[1].Chips = 0
	g.Players[1].TotalBetInHand = 5000
	g.Players[1].Status = PlayerStatusAllIn
	g.Players[1].Hand = poker.CardsFromStrings("Qs Qc Qd Jh Js") // Full House (second best)

	g.Players[2].Chips = 0
	g.Players[2].TotalBetInHand = 10000
	g.Players[2].Status = PlayerStatusAllIn
	g.Players[2].Hand = poker.CardsFromStrings("Ts 9c 8d 7h 6s") // Straight (worst hand)

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

// TestDistributePot_FoldedPlayerBetNotLost tests that a folded player's contribution to the pot
// is not lost during distribution.
func TestDistributePot_FoldedPlayerBetNotLost(t *testing.T) {
	util.InitLogger(true)

	// Scenario: 3 players. Player C bets 1000 and folds. Player A and B go to showdown with 3000 each.
	// The total pot should be 7000. Player A has the winning hand.
	playerNames := []string{"Player A", "Player B", "Player C"}
	rules := &config.GameRules{
		HoleCards: config.HoleCardRules{
			Count: 5, // Does not matter for this test
		},
		LowHand: config.LowHandRules{Enabled: false},
	}
	g := NewGame(playerNames, 10000, DifficultyMedium, rules, true, false)

	// Setup player states
	g.Players[0].Chips = 7000
	g.Players[0].TotalBetInHand = 3000
	g.Players[0].Status = PlayerStatusPlaying                    // Showdown
	g.Players[0].Hand = poker.CardsFromStrings("As Ac Ad Ah Ks") // Four of a Kind (Winner)

	g.Players[1].Chips = 7000
	g.Players[1].TotalBetInHand = 3000
	g.Players[1].Status = PlayerStatusPlaying                    // Showdown
	g.Players[1].Hand = poker.CardsFromStrings("Qs Qc Qd Qh Js") // Four of a Kind (Loser)

	g.Players[2].Chips = 9000
	g.Players[2].TotalBetInHand = 1000
	g.Players[2].Status = PlayerStatusFolded // Folded

	// Total pot is the sum of all bets
	g.Pot = 3000 + 3000 + 1000

	// Action: Distribute the pot
	results := g.DistributePot()

	// --- Assertions ---
	// Expected distribution:
	// Player A should win the entire pot of 7000.
	// The current buggy implementation will only distribute 6000.

	if len(results) != 1 {
		t.Fatalf("Expected 1 distribution result, but got %d", len(results))
	}

	// Check chip distribution
	if g.Players[0].Chips != 14000 { // Initial 7000 + Pot 7000
		t.Errorf("Expected Player A to have 14000 chips, but got %d", g.Players[0].Chips)
	}
	if g.Players[1].Chips != 7000 {
		t.Errorf("Expected Player B to have 7000 chips, but got %d", g.Players[1].Chips)
	}
	if g.Pot != 0 {
		t.Errorf("Expected pot to be 0 after distribution, but got %d", g.Pot)
	}
}

// TestDistributePot_ComplexSidePotAndAllIn reproduces the specific bug found in the log file.
// This test covers a complex scenario with multiple all-ins, side pots, and a call.
func TestDistributePot_ComplexSidePotAndAllIn(t *testing.T) {
	util.InitLogger(true)

	// Scenario setup based on the bug log
	playerNames := []string{"YOU", "CPU 1", "CPU 4"}
	rules := &config.GameRules{
		HoleCards:    config.HoleCardRules{Count: 3},
		LowHand:      config.LowHandRules{Enabled: true, MaxRank: 7},
		HandRankings: config.HandRankingsRules{UseStandardRankings: true},
	}
	g := NewGame(playerNames, 0, DifficultyEasy, rules, true, false)

	// Player states based on the corrected scenario
	// YOU: Calls the final all-in
	g.Players[0].Chips = 1136500 - 254500 // Initial chips before the final bets
	g.Players[0].TotalBetInHand = 254500
	g.Players[0].Status = PlayerStatusPlaying
	g.Players[0].Hand = poker.CardsFromStrings("As 2s 3s") // Hand for Straight and Low

	// CPU 1: All-in with the highest bet
	g.Players[1].Chips = 0
	g.Players[1].TotalBetInHand = 254500
	g.Players[1].Status = PlayerStatusAllIn
	g.Players[1].Hand = poker.CardsFromStrings("6c Jc 9h") // A-High

	// CPU 4: All-in with a lower bet
	g.Players[2].Chips = 13000 - 205000 // Reflects state before final all-in
	g.Players[2].TotalBetInHand = 205000
	g.Players[2].Status = PlayerStatusAllIn
	g.Players[2].Hand = poker.CardsFromStrings("Ts 6s 4h") // Two Pair

	// Community cards
	g.CommunityCards = poker.CardsFromStrings("Ad Kd 5c 4d Th")

	// Total pot is the sum of all bets
	g.Pot = 254500 + 254500 + 205000

	// Action
	results := g.DistributePot()

	// --- Assertions ---
	// Expected distribution:
	// Main Pot (205,000 * 3 = 615,000): YOU wins high and low (scoops).
	// Side Pot 1 ((254,500 - 205,000) * 2 = 99,000): Between YOU and CPU 1. YOU wins high and low.
	// Total for YOU: 615,000 + 99,000 = 714,000

	// Find the result for "YOU"
	var youResult *DistributionResult
	for i := range results {
		if results[i].PlayerName == "YOU" {
			youResult = &results[i]
			break
		}
	}

	if youResult == nil {
		t.Fatalf("Expected results for player 'YOU', but found none")
	}

	if youResult.AmountWon != 714000 {
		t.Errorf("Expected YOU to win 714000, but got %d", youResult.AmountWon)
	}

	// Check final chip counts
	// Initial chips for YOU: 1136500 - 254500 = 882000
	// Final chips: 882000 + 714000 = 1596000
	if g.Players[0].Chips != 1596000 {
		t.Errorf("Expected YOU's final chips to be 1596000, but got %d", g.Players[0].Chips)
	}

	if g.Players[1].Chips != 0 {
		t.Errorf("Expected CPU 1's final chips to be 0, but got %d", g.Players[1].Chips)
	}
}
