package poker

import (
	"pls7-cli/internal/util"
	"testing"
)

func TestPLS7HighHands(t *testing.T) {
	util.InitLogger(true)

	testCases := []struct {
		name         string
		cardString   string
		expectedRank HandRank
	}{
		// Special High Hands
		{name: "Skip Straight (A-3-5-7-9)", cardString: "As 9c 7s 6d 5s 5c 3s", expectedRank: SkipStraight},
		{name: "Skip Straight (2-4-6-8-T)", cardString: "2s 4c 6d 8h Ts 3c 5d", expectedRank: SkipStraight},
		{name: "Skip Straight (3-5-7-9-J)", cardString: "3s 5c 7d 9h Js 2c 4d", expectedRank: SkipStraight},
		{name: "Skip Straight (4-6-8-T-Q)", cardString: "4s 6c 8d Th Qs 2c 3d", expectedRank: SkipStraight},
		{name: "Skip Straight (5-7-9-J-K)", cardString: "5s 7c 9d Jh Ks 2c 3d", expectedRank: SkipStraight},
		{name: "Skip Straight (6-8-T-Q-A)", cardString: "6s 8c Td Qh As 2c 3d", expectedRank: SkipStraight},
		{name: "Not Skip Straight", cardString: "Qs Tc 8d 2h As Ac 8d", expectedRank: TwoPair},

		// Standard High Hands
		{name: "Royal Flush", cardString: "As Ks Qs Js Ts 2c 3d 4h", expectedRank: RoyalFlush},
		{name: "Straight Flush (A-5)", cardString: "As 2s 3s 4s 5s Kc Qd Jh", expectedRank: StraightFlush},
		{name: "Four of a Kind with a Pair", cardString: "As Ac Ah Ad Ks Kc Qs Jc", expectedRank: FourOfAKind},
		{name: "Full House", cardString: "As Ac Ah Ks Kc 2d 3c 4h", expectedRank: FullHouse},
		{name: "Flush", cardString: "As Ks Qs Js 2s 3c 4d 5h", expectedRank: Flush},
		{name: "Straight", cardString: "As Kc Qd Jh Ts 2c 3d 4h", expectedRank: Straight},
		{name: "Three of a Kind", cardString: "As Ac Ah Ks Qc Jd 8c 2h", expectedRank: ThreeOfAKind},
		{name: "Two Pair", cardString: "As Ac Ks Kc Qs Jd 7c 2h", expectedRank: TwoPair},
		{name: "One Pair", cardString: "As Ac Ks Qc Jd 6c 2h 3d", expectedRank: OnePair},
		{name: "High Card", cardString: "As Ks Qs Jc 9d 2h 3c 4d", expectedRank: HighCard},

		// New: Skip Straight Flush
		{name: "Skip Straight Flush (A-Low)", cardString: "As 3s 5s 7s 9s Kd Qc Jc", expectedRank: SkipStraightFlush},
		{name: "Skip Straight Flush (K-High)", cardString: "Ks Js 9s 7s 5s Ad Qc Th", expectedRank: SkipStraightFlush},

		// Ranking & Tie-Breakers
		{name: "Flush vs Straight", cardString: "As Ks Qs Js 2s 4c 5d 6h", expectedRank: Flush},

		// Hand Composition
		{name: "Board Play (Straight)", cardString: "2c 3d 4h Ah Ks Qd Jc Tc", expectedRank: Straight},
		{name: "1 Card Play (Royal Flush)", cardString: "As 2c 3d Ks Qs Js Ts 4h", expectedRank: RoyalFlush},
		{name: "2 Card Play (Full House)", cardString: "As Ks 2d Ac Ah Kd 3h 4c", expectedRank: FullHouse},
		{name: "3 Card Play (Full House)", cardString: "As Ac Ah Ks Kc 2d 3c 4d", expectedRank: FullHouse},
	}

	gameRules := &GameRules{
		LowHand: LowHandRules{
			Enabled: false, // Low hands are not enabled for these tests
			MaxRank: 0,     // No low hand rules apply
		},
		HandRankings: HandRankingsRules{
			UseStandardRankings: false,
			CustomRankings: []CustomHandRanking{
				{Name: "skip_straight_flush", InsertAfterRank: "royal_flush"},
				{Name: "skip_straight", InsertAfterRank: "flush"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := CardsFromStrings(tc.cardString)
			highHand, _ := EvaluateHand(pool[:3], pool[3:], gameRules)

			if highHand == nil {
				t.Fatalf("Expected rank %v, but got nil", tc.expectedRank)
			}
			if highHand.Rank != tc.expectedRank {
				t.Errorf("Expected rank %v, but got %v", tc.expectedRank, highHand.Rank)
			}
		})
	}
}

func TestPLS7LowHands(t *testing.T) {
	testCases := []struct {
		name           string
		cardString     string
		lowGameEnabled bool   // New flag for testing
		expectLowHand  bool   // Does a low hand exist?
		expectedValues string // Expected the best low hand, e.g., "7 6 4 2 A"
	}{
		{name: "Nut Low (A-5)", cardString: "As 2c 3d 4h 5s 8s 9s Ts", lowGameEnabled: true, expectLowHand: true, expectedValues: "5 4 3 2 A"},
		{name: "7-High Low", cardString: "As 2c 4d 6h 7s 8s 9s Ts", lowGameEnabled: true, expectLowHand: true, expectedValues: "7 6 4 2 A"},
		{name: "No Low (Not enough cards)", cardString: "As 2c 3d 4h 8s 9s Ts Js", lowGameEnabled: true, expectLowHand: false},
		{name: "No Low (Pair exists)", cardString: "As Ac 2d 3h 4s 8s 9s Ts", lowGameEnabled: true, expectLowHand: false},
		{name: "High/Low Combo (Straight Flush and Low)", cardString: "As 2s 3s 4s 5s 8c 9d Th", lowGameEnabled: true, expectLowHand: true, expectedValues: "5 4 3 2 A"},
		// when lowGameEnabled is false, we should not expect a low hand
		{name: "LowGameEnabled set to false - No Low Hand Expected", cardString: "As 2c 3d 4h 5s 8s 9s Ts", lowGameEnabled: false, expectLowHand: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := CardsFromStrings(tc.cardString)
			gameRules := &GameRules{
				LowHand: LowHandRules{
					Enabled: tc.lowGameEnabled,
					MaxRank: 7, // Assuming 7-or-better for low hands
				},
			}
			_, lowHand := EvaluateHand(pool[:3], pool[3:], gameRules)

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

func TestPLS7FindBestStraight(t *testing.T) {
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
			cardString:       "As Kd Qc Jh Ts 2s 3d 4c",
			expectStraight:   true,
			expectedTopRank:  Ace,
			expectedCardsStr: "[As][Kd][Qc][Jh][Ts]",
		},
		{
			name:             "Ace Low Straight (Wheel)",
			cardString:       "As 2d 3c 4h 5s Ks Qd Jc",
			expectStraight:   true,
			expectedTopRank:  Five, // In a wheel, Five is the high card for ranking purposes
			expectedCardsStr: "[5s][4h][3c][2d][As]",
		},
		{
			name:           "No Straight",
			cardString:     "As Ks Qs Js 9s 2c 3d 4h",
			expectStraight: false,
		},
		{
			name:             "Straight with Pairs",
			cardString:       "As Ac 5d 4c 3h 2s Ks Qd",
			expectStraight:   true,
			expectedTopRank:  Five,
			expectedCardsStr: "[5d][4c][3h][2s][As]",
		},
		{
			name:             "Longer than 5 card straight",
			cardString:       "9s 8d 7c 6h 5s 4d 3c 2h",
			expectStraight:   true,
			expectedTopRank:  Nine, // Should find the highest possible straight
			expectedCardsStr: "[9s][8d][7c][6h][5s]",
		},
		{
			name:           "Broken Straight",
			cardString:     "As Ks Qs Jc 9d 8h 7c 6s",
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

// TestPLS7LowHandComparison specifically tests the comparison logic between two low hands.
func TestPLS7LowHandComparison(t *testing.T) {
	util.InitLogger(true)

	// compare is a helper to simulate the comparison logic.
	// Returns 1 if h1 is better (lower), -1 if h2 is better, 0 if tie.
	compare := func(h1, h2 *HandResult) int {
		for i := 0; i < 5; i++ {
			v1 := getLowRankValue(h1.HighValues[i])
			v2 := getLowRankValue(h2.HighValues[i])
			if v1 < v2 {
				return 1 // h1 is better
			}
			if v1 > v2 {
				return -1 // h2 is better
			}
		}
		return 0 // Tie
	}

	testCases := []struct {
		name           string
		hand1Str       string // Pool for hand 1
		hand2Str       string // Pool for hand 2
		expectedWinner int    // 1 for hand1, -1 for hand2
	}{
		{
			name:           "7-6-5-3-A should beat 7-6-5-4-3",
			hand1Str:       "As 7d 6s 5c 3h Ks Qs Js", // Makes 7-6-5-3-A
			hand2Str:       "7d 6s 5c 4d 3h Ks Qs Js", // Makes 7-6-5-4-3
			expectedWinner: 1,
		},
		{
			name:           "6-5-4-3-2 should beat 7-5-4-3-2",
			hand1Str:       "6s 5d 4c 3h 2s Ks Qs Js", // Makes 6-5-4-3-2
			hand2Str:       "7s 5d 4c 3h 2s Ks Qs Js", // Makes 7-5-4-3-2
			expectedWinner: 1,
		},
		{
			name:           "Nut low should beat 6-low",
			hand1Str:       "As 2d 3c 4h 5s Ks Qs Js", // Makes A-2-3-4-5
			hand2Str:       "As 2d 3c 4h 6s Ks Qs Js", // Makes A-2-3-4-6
			expectedWinner: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool1 := CardsFromStrings(tc.hand1Str)
			pool2 := CardsFromStrings(tc.hand2Str)
			gameRules := &GameRules{
				LowHand: LowHandRules{
					Enabled: true,
					MaxRank: 7,
				},
			}

			_, lowHand1 := EvaluateHand(pool1[:3], pool1[3:], gameRules)
			_, lowHand2 := EvaluateHand(pool2[:3], pool2[3:], gameRules)

			if lowHand1 == nil || lowHand2 == nil {
				t.Fatal("Both hands should qualify for a low hand in this test")
			}

			winner := compare(lowHand1, lowHand2)
			if winner != tc.expectedWinner {
				t.Errorf("Expected winner to be %d, but got %d. Hand1: %v, Hand2: %v",
					tc.expectedWinner, winner, lowHand1.HighValues, lowHand2.HighValues)
			}
		})
	}
}

// TestHandRankOrder tests the order of hand ranks by given game rules.
func TestHandRankOrder(t *testing.T) {
	util.InitLogger(true)

	testCases := []struct {
		name         string
		gameRules    *HandRankingsRules
		expectedRank []HandRank
	}{
		{
			name: "Standard Hand Rankings",
			gameRules: &HandRankingsRules{
				UseStandardRankings: true,
				CustomRankings:      []CustomHandRanking{},
			},
			expectedRank: []HandRank{
				RoyalFlush,
				StraightFlush,
				FourOfAKind,
				FullHouse,
				Flush,
				Straight,
				ThreeOfAKind,
				TwoPair,
				OnePair,
				HighCard,
			},
		},
		{
			name: "Custom Hand Rankings with Skip Straight Flush and Skip Straight",
			gameRules: &HandRankingsRules{
				UseStandardRankings: false,
				CustomRankings: []CustomHandRanking{
					{Name: "skip_straight_flush", InsertAfterRank: "royal_flush"},
					{Name: "skip_straight", InsertAfterRank: "flush"},
				},
			},
			expectedRank: []HandRank{
				RoyalFlush,
				SkipStraightFlush,
				StraightFlush,
				FourOfAKind,
				FullHouse,
				Flush,
				SkipStraight,
				Straight,
				ThreeOfAKind,
				TwoPair,
				OnePair,
				HighCard,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := getHandRanks(tc.gameRules)
			if len(actual) != len(tc.expectedRank) {
				t.Fatalf("Expected %d hand ranks, but got %d", len(tc.expectedRank), len(actual))
			}
			for i, rank := range actual {
				if rank != tc.expectedRank[i] {
					t.Errorf("Expected rank %v at position %d, but got %v", tc.expectedRank[i], i, rank)
				}
			}
		})
	}
}
