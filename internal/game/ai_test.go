package game

import (
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

			score := g.evaluateHandStrength(player)

			if score < tc.minExpectedScore || score > tc.maxExpectedScore {
				t.Errorf("Expected score between %.2f and %.2f, but got %.2f", tc.minExpectedScore, tc.maxExpectedScore, score)
			}
		})
	}
}
