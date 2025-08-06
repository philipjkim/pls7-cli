package poker

import (
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"testing"
)

func TestNLHHighHands(t *testing.T) {
	util.InitLogger(true)

	testCases := []struct {
		name         string
		cardString   string
		expectedRank HandRank
	}{
		{name: "Royal Flush", cardString: "As Ks Qs Js Ts 2c 3d", expectedRank: RoyalFlush},
		{name: "Straight Flush (A-5)", cardString: "As 2s 3s 4s 5s Kc Qd", expectedRank: StraightFlush},
		{name: "Four of a Kind with a Pair", cardString: "As Ac Ah Ad Ks Kc Qs", expectedRank: FourOfAKind},
		{name: "Full House", cardString: "As Ac Ah Ks Kc 2d 3c", expectedRank: FullHouse},
		{name: "Flush", cardString: "As Ks Qs Js 2s 3c 4d", expectedRank: Flush},
		{name: "Straight", cardString: "As Kc Qd Jh Ts 2c 3d", expectedRank: Straight},
		{name: "Three of a Kind", cardString: "As Ac Ah Ks Qc Jd 8c", expectedRank: ThreeOfAKind},
		{name: "Two Pair", cardString: "As Ac Ks Kc Qs Jd 7c", expectedRank: TwoPair},
		{name: "One Pair", cardString: "As Ac Ks Qc Jd 6c 2h", expectedRank: OnePair},
		{name: "High Card", cardString: "As Ks Qs Jc 9d 2h 3c", expectedRank: HighCard},

		// Ranking & Tie-Breakers
		{name: "Flush vs Straight", cardString: "As Ks Qs Js 2s Tc 5d", expectedRank: Flush},

		// Hand Composition
		{name: "Board Play (Straight)", cardString: "2c 3d Ah Ks Qd Jc Tc", expectedRank: Straight},
		{name: "1 Card Play (Royal Flush)", cardString: "As 2c 3d Ks Qs Js Ts", expectedRank: RoyalFlush},
		{name: "2 Card Play (Full House)", cardString: "As Ks 2d Ac Ah Kd 3h", expectedRank: FullHouse},
		{name: "3 Card Play (Full House)", cardString: "As Ac Ah Ks Kc 2d 3c", expectedRank: FullHouse},
	}

	gameRules := &config.GameRules{
		LowHand: config.LowHandRules{
			Enabled: false, // Low hands are not enabled for these tests
			MaxRank: 0,     // No low hand rules apply
		},
		HandRankings: config.HandRankingsRules{
			UseStandardRankings: true,
			CustomRankings:      []config.CustomHandRanking{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := CardsFromStrings(tc.cardString)
			highHand, _ := EvaluateHand(pool[:2], pool[2:], gameRules)

			if highHand == nil {
				t.Fatalf("Expected rank %v, but got nil", tc.expectedRank)
			}
			if highHand.Rank != tc.expectedRank {
				t.Errorf("Expected rank %v, but got %v", tc.expectedRank, highHand.Rank)
			}
		})
	}
}

func TestNLHFindBestStraight(t *testing.T) {
	testCases := []struct {
		name             string
		cardString       string
		expectStraight   bool
		expectedTopRank  Rank   // The highest card of the expected straight
		expectedCardsStr string // For visual confirmation
	}{
		{
			name:             "Standard Straight",
			cardString:       "9s 8d 7c 6h 5s 2s 3d",
			expectStraight:   true,
			expectedTopRank:  Nine,
			expectedCardsStr: "[9s][8d][7c][6h][5s]",
		},
		{
			name:             "Ace High Straight (Mountain)",
			cardString:       "As Kd Qc Jh Ts 2s 3d",
			expectStraight:   true,
			expectedTopRank:  Ace,
			expectedCardsStr: "[As][Kd][Qc][Jh][Ts]",
		},
		{
			name:             "Ace Low Straight (Wheel)",
			cardString:       "As 2d 3c 4h 5s Ks Qd",
			expectStraight:   true,
			expectedTopRank:  Five, // In a wheel, Five is the high card for ranking purposes
			expectedCardsStr: "[5s][4h][3c][2d][As]",
		},
		{
			name:           "No Straight",
			cardString:     "As Ks Qs Js 9s 2c 3d",
			expectStraight: false,
		},
		{
			name:             "Straight with Pairs",
			cardString:       "As Ac 5d 4c 3h 2s Ks",
			expectStraight:   true,
			expectedTopRank:  Five,
			expectedCardsStr: "[5d][4c][3h][2s][As]",
		},
		{
			name:             "Longer than 5 card straight",
			cardString:       "9s 8d 7c 6h 5s 4d 3c",
			expectStraight:   true,
			expectedTopRank:  Nine, // Should find the highest possible straight
			expectedCardsStr: "[9s][8d][7c][6h][5s]",
		},
		{
			name:           "Broken Straight",
			cardString:     "As Ks Qs Jc 9d 8h 7c",
			expectStraight: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := CardsFromStrings(tc.cardString)
			analysis := newHandAnalysis(pool)

			straightCards, ok := findBestStraight(analysis)

			if ok != tc.expectStraight {
				t.Fatalf("Expected straight existence to be %v, but got %v", tc.expectStraight, ok)
			}

			if tc.expectStraight {
				if len(straightCards) != 5 {
					t.Fatalf("Expected 5 cards for a straight, but got %d", len(straightCards))
				}
				// The first card in the returned slice should be the highest rank
				if straightCards[0].Rank != tc.expectedTopRank {
					t.Errorf("Expected straight to be topped by %v, but got %v. Hand: %v", tc.expectedTopRank, straightCards[0].Rank, straightCards)
				}
			}
		})
	}
}
