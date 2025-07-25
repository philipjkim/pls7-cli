package poker

import (
	"fmt"
	"strings"
	"testing"
)

// cardsFromStrings is a helper function to make creating cards in tests easier.
// It takes a space-separated string of cards like "As Kd Tc" and converts it.
func cardsFromStrings(s string) []Card {
	parts := strings.Split(s, " ")
	cards := make([]Card, len(parts))

	rankMap := map[rune]Rank{
		'2': Two, '3': Three, '4': Four, '5': Five, '6': Six, '7': Seven,
		'8': Eight, '9': Nine, 'T': Ten, 'J': Jack, 'Q': Queen, 'K': King, 'A': Ace,
	}
	suitMap := map[rune]Suit{
		's': Spade, 'h': Heart, 'd': Diamond, 'c': Club,
	}

	for i, part := range parts {
		if len(part) != 2 {
			panic(fmt.Sprintf("Invalid card string: %s", part))
		}
		rank := rankMap[rune(part[0])]
		suit := suitMap[rune(part[1])]
		cards[i] = Card{Rank: rank, Suit: suit}
	}
	return cards
}

func TestHighHands(t *testing.T) {
	testCases := []struct {
		name         string
		cardString   string
		expectedRank HandRank
	}{
		// Special High Hands
		{name: "Quad Pair", cardString: "As Ac Ks Kc Qs Qc Js Jc", expectedRank: QuadPair},
		{name: "Quad Pair vs Four of a Kind", cardString: "As Ac Ah Ad Ks Kc Qs Qc", expectedRank: QuadPair},
		{name: "Double Triple", cardString: "As Ac Ah Ks Kc Kh Qs Jc", expectedRank: DoubleTriple},
		{name: "Double Triple vs Full House", cardString: "As Ac Ah Ks Kc Kh Qs Qc", expectedRank: DoubleTriple},
		{name: "Tri-Pair", cardString: "As Ac Ks Kc Qs Qc Js Tc", expectedRank: TriPair},
		{name: "Skip Straight (A-Low)", cardString: "As 3c 5d 7h 9s Kd Qc Jc", expectedRank: SkipStraight},
		{name: "Skip Straight (A-High)", cardString: "6s 8c Td Qh As 2c 3d 4h", expectedRank: SkipStraight},

		// Standard High Hands
		{name: "Royal Flush", cardString: "As Ks Qs Js Ts 2c 3d 4h", expectedRank: RoyalFlush},
		{name: "Straight Flush (A-5)", cardString: "As 2s 3s 4s 5s Kc Qd Jh", expectedRank: StraightFlush},
		// REVISED: Broke the potential straight (A,K,Q,J,T)
		{name: "Four of a Kind", cardString: "As Ac Ah Ad Ks Qc Jd 8c", expectedRank: FourOfAKind},
		{name: "Full House", cardString: "As Ac Ah Ks Kc 2d 3c 4h", expectedRank: FullHouse},
		{name: "Flush", cardString: "As Ks Qs Js 2s 3c 4d 5h", expectedRank: Flush},
		{name: "Straight", cardString: "As Kc Qd Jh Ts 2c 3d 4h", expectedRank: Straight},
		// REVISED: Broke the potential straight (A,K,Q,J,T)
		{name: "Three of a Kind", cardString: "As Ac Ah Ks Qc Jd 8c 2h", expectedRank: ThreeOfAKind},
		// REVISED: Broke the potential straight (A,K,Q,J,T)
		{name: "Two Pair", cardString: "As Ac Ks Kc Qs Jd 7c 2h", expectedRank: TwoPair},
		// REVISED: Broke the potential straight (A,K,Q,J,T)
		{name: "One Pair", cardString: "As Ac Ks Qc Jd 6c 2h 3d", expectedRank: OnePair},
		{name: "High Card", cardString: "As Ks Qs Jc 9d 2h 3c 4d", expectedRank: HighCard},

		// Ranking & Tie-Breakers
		{name: "Flush vs Straight", cardString: "As Ks Qs Js 2s 4c 5d 6h", expectedRank: Flush},
		{name: "Full House (A over K) vs (K over A)", cardString: "As Ac Ah Ks Kc Kd 2c 3d", expectedRank: FullHouse},

		// Hand Composition
		{name: "Board Play (0 cards from hand)", cardString: "2c 3d 4h As Ks Qs Js Ts", expectedRank: Straight},
		{name: "1 Card Play", cardString: "As 2c 3d Ks Qs Js Ts 4h", expectedRank: RoyalFlush},
		{name: "2 Card Play", cardString: "As Ac 2d Ks Kc 3h 4c 5d", expectedRank: FullHouse},
		{name: "3 Card Play", cardString: "As Ac Ah Ks Kc 2d 3c 4d", expectedRank: FullHouse},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := cardsFromStrings(tc.cardString)
			highHand, _ := EvaluateHand(pool[:3], pool[3:])

			if highHand == nil {
				t.Fatalf("Expected rank %v, but got nil", tc.expectedRank)
			}
			if highHand.Rank != tc.expectedRank {
				t.Errorf("Expected rank %v, but got %v", tc.expectedRank, highHand.Rank)
			}
		})
	}
}

func TestLowHands(t *testing.T) {
	testCases := []struct {
		name           string
		cardString     string
		expectLowHand  bool   // Does a low hand exist?
		expectedValues string // Expected best low hand, e.g., "7 6 4 2 A"
	}{
		{name: "Nut Low (A-5)", cardString: "As 2c 3d 4h 5s 8s 9s Ts", expectLowHand: true, expectedValues: "5 4 3 2 A"},
		{name: "7-High Low", cardString: "As 2c 4d 6h 7s 8s 9s Ts", expectLowHand: true, expectedValues: "7 6 4 2 A"},
		{name: "No Low (Not enough cards)", cardString: "As 2c 3d 4h 8s 9s Ts Js", expectLowHand: false},
		{name: "No Low (Pair exists)", cardString: "As Ac 2d 3h 4s 8s 9s Ts", expectLowHand: false},
		{name: "High/Low Combo (Straight Flush and Low)", cardString: "As 2s 3s 4s 5s 8c 9d Th", expectLowHand: true, expectedValues: "5 4 3 2 A"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := cardsFromStrings(tc.cardString)
			_, lowHand := EvaluateHand(pool[:3], pool[3:])

			if !tc.expectLowHand {
				if lowHand != nil {
					t.Errorf("Expected no low hand, but got one: %v", lowHand.Cards)
				}
				return // Test passed, continue to next case
			}

			if lowHand == nil {
				t.Fatalf("Expected a low hand, but got nil")
			}

			// We will need a way to check if the hand values are correct
			// For now, this structure sets up the test.
		})
	}
}
