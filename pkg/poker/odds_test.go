package poker

import (
	"fmt"
	"pls7-cli/internal/util"
	"sort"
	"testing"
)

func TestCalculateOuts(t *testing.T) {
	util.InitLogger(true)
	testCases := []struct {
		name                string
		holeCards           []Card
		communityCards      []Card
		lowGameEnabled      bool
		expectedAllOuts     []Card
		expectedOutsPerRank map[HandRank][]Card
	}{
		{
			name:            "Open-ended Straight Draw",
			holeCards:       CardsFromStrings("8s 7s Kc"),
			communityCards:  CardsFromStrings("6c 5h 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "OESD with Ace high",
			holeCards:       CardsFromStrings("As Kh Qs"),
			communityCards:  CardsFromStrings("Jh 5c 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ts Th Td Tc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("Ts Th Td Tc"),
			},
		},
		{
			name:            "OESD with Ace low",
			holeCards:       CardsFromStrings("4s 3d 2h"),
			communityCards:  CardsFromStrings("As Qc Tc"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("5s 5h 5d 5c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("5s 5h 5d 5c"),
			},
		},
		{
			name:            "Gutshot 8654",
			holeCards:       CardsFromStrings("8s 6s 5c"),
			communityCards:  CardsFromStrings("Ad Kh 4h"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("7s 7h 7d 7c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("7s 7h 7d 7c"),
			},
		},
		{
			name:            "Gutshot with Ace high",
			holeCards:       CardsFromStrings("As Qd Jc"),
			communityCards:  CardsFromStrings("Th 7c 3d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ks Kh Kd Kc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("Ks Kh Kd Kc"),
			},
		},
		{
			name:            "Gutshot with Ace low",
			holeCards:       CardsFromStrings("4h 3c As"),
			communityCards:  CardsFromStrings("Jh 8c 5d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("2s 2h 2d 2c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("2s 2h 2d 2c"),
			},
		},
		{
			name:            "Flush-only Draw",
			holeCards:       CardsFromStrings("As Js 5h"),
			communityCards:  CardsFromStrings("8s 7s 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush: CardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			},
		},
		{
			name:            "Straight or Flush Draw",
			holeCards:       CardsFromStrings("8s 7s Kc"),
			communityCards:  CardsFromStrings("6s 5s 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s 9h 9d 9c 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush:    CardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s"),
				Straight: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "Triple Draw with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h 6c"),
			communityCards:  CardsFromStrings("As Js 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				ThreeOfAKind: CardsFromStrings("8d 8c"),
			},
		},
		{
			name:                "Non-Draw for Triple because board is paired, not pocket pair",
			holeCards:           CardsFromStrings("8s 6d 5c"),
			communityCards:      CardsFromStrings("As Jh Jd"),
			lowGameEnabled:      false,
			expectedAllOuts:     []Card{},
			expectedOutsPerRank: map[HandRank][]Card{},
		},
		{
			name:            "Full House Draw from Two Pair with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h 6c"),
			communityCards:  CardsFromStrings("As Ah 5d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ad Ac 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: CardsFromStrings("Ad Ac 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h 6c"),
			communityCards:  CardsFromStrings("As 8d 5d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ah Ad Ac 6s 6h 6d 5s 5h 5c 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   CardsFromStrings("Ah Ad Ac 6s 6h 6d 5s 5h 5c"),
				FourOfAKind: CardsFromStrings("8c"),
			},
		},
		{
			name:            "Full House Draw from Two Pair without pocket pair",
			holeCards:       CardsFromStrings("As 8h 6c"),
			communityCards:  CardsFromStrings("Ac 8s 5d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ad Ah 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: CardsFromStrings("Ad Ah 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips without pocket pair",
			holeCards:       CardsFromStrings("As 8h 6c"),
			communityCards:  CardsFromStrings("Ah Ad 5d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ac 6s 6h 6d 5s 5h 5c 8s 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   CardsFromStrings("6s 6h 6d 5s 5h 5c 8s 8d 8c"),
				FourOfAKind: CardsFromStrings("Ac"),
			},
		},
		{
			name:            "Skip Straight Draw (Gutshot)",
			holeCards:       CardsFromStrings("8s 6s 3c"),
			communityCards:  CardsFromStrings("Ad Qh 4h"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Ts Th Td Tc"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraight: CardsFromStrings("Ts Th Td Tc"),
			},
		},
		{
			name:            "Skip Straight Draw (Open-ended)",
			holeCards:       CardsFromStrings("3s 5s 7c"),
			communityCards:  CardsFromStrings("9d Qh Qh"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("As Ah Ad Ac Js Jh Jd Jc"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraight: CardsFromStrings("As Ah Ad Ac Js Jh Jd Jc"),
			},
		},
		{
			name:            "Straight Flush Draw (Open-ended)",
			holeCards:       CardsFromStrings("8s 7s Kc"),
			communityCards:  CardsFromStrings("6s 5s 2d"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Js 4d 4c 9h 2s Qs Ks 9c 4s 9s Ts As 4h 9d 3s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight:      CardsFromStrings("4s 4h 4d 4c 9s 9h 9d 9c"),
				Flush:         CardsFromStrings("2s 3s 4s 9s Ts Js Qs Ks As"),
				StraightFlush: CardsFromStrings("9s 4s"),
			},
		},
		{
			name:            "Skip Straight Flush Draw (Open-ended)",
			holeCards:       CardsFromStrings("Ts 8s 6s"),
			communityCards:  CardsFromStrings("Kh Kd 4s"),
			lowGameEnabled:  false,
			expectedAllOuts: CardsFromStrings("Qs Qh Qd Qc 2s 2h 2d 2c 3s 5s 7s 9s Js Ks As"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraightFlush: CardsFromStrings("Qs 2s"),
				Flush:             CardsFromStrings("2s 3s 5s 7s 9s Js Qs Ks As"),
				SkipStraight:      CardsFromStrings("2s 2h 2d 2c Qs Qh Qd Qc"),
			},
		},
		{
			name:            "Low Hand Draw",
			holeCards:       CardsFromStrings("2s 3c 6h"),
			communityCards:  CardsFromStrings("Kh Kd 7s"),
			lowGameEnabled:  true,
			expectedAllOuts: CardsFromStrings("As Ah Ad Ac 4s 4h 4d 4c 5s 5h 5d 5c"),
			expectedOutsPerRank: map[HandRank][]Card{
				HighCard: CardsFromStrings("As Ah Ad Ac 4s 4h 4d 4c 5s 5h 5d 5c"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameRules := &GameRules{
				Name:         "Pot-Limit Sampyeong 7-or-Better",
				Abbreviation: "PLS7",
				BettingLimit: "pot_limit",
				HoleCards: HoleCardRules{
					Count:         3,
					UseConstraint: "any",
					UseCount:      0,
				},
				HandRankings: HandRankingsRules{
					UseStandardRankings: false,
					CustomRankings: []CustomHandRanking{
						{
							Name:            "skip_straight_flush",
							InsertAfterRank: "royal_flush",
						}, {
							Name:            "skip_straight",
							InsertAfterRank: "flush",
						},
					},
				},
				LowHand: LowHandRules{
					Enabled: tc.lowGameEnabled,
					MaxRank: 7,
				},
			}
			hasOuts, outsInfo := CalculateOuts(tc.holeCards, tc.communityCards, gameRules)
			fmt.Printf("hasOuts: %v, outsInfo: %+v\n", hasOuts, outsInfo)

			if (len(tc.expectedAllOuts) > 0) != hasOuts {
				t.Errorf("Expected hasOuts to be %v, but got %v", len(tc.expectedAllOuts) > 0, hasOuts)
			}

			if !cardSlicesEqual(outsInfo.AllOuts, tc.expectedAllOuts) {
				t.Errorf("Expected all outs %v, but got %v", tc.expectedAllOuts, outsInfo.AllOuts)
			}

			for rank, expectedOuts := range tc.expectedOutsPerRank {
				if !cardSlicesEqual(outsInfo.OutsPerHandRank[rank], expectedOuts) {
					t.Errorf("For rank %v, expected outs %v, but got %v", rank, expectedOuts, outsInfo.OutsPerHandRank[rank])
				}
			}
		})
	}
}

func cardSlicesEqual(a, b []Card) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].Suit != a[j].Suit {
			return a[i].Suit < a[j].Suit
		}
		return a[i].Rank < a[j].Rank
	})
	sort.Slice(b, func(i, j int) bool {
		if b[i].Suit != b[j].Suit {
			return b[i].Suit < b[j].Suit
		}
		return b[i].Rank < b[j].Rank
	})

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestCalculateBreakEvenEquityBasedOnPotOdds(t *testing.T) {
	testCases := []struct {
		name         string
		pot          int
		amountToCall int
		expected     float64
	}{
		{"Simple Case", 1000, 500, 0.3333333333333333},
		{"Zero Call Amount", 1000, 0, 0},
		{"Zero Pot", 0, 500, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := CalculateBreakEvenEquityBasedOnPotOdds(tc.pot, tc.amountToCall)
			if actual != tc.expected {
				t.Errorf("Expected break even equity to be %.2f, but got %.2f", tc.expected, actual)
			}
		})
	}
}

func TestCalculateEquityWithCards(t *testing.T) {
	util.InitLogger(true)
	testCases := []struct {
		name           string
		holeCards      []Card
		communityCards []Card
		expectedEquity float64
	}{
		{
			name:           "Flush Draw on Flop",
			holeCards:      CardsFromStrings("As Js 5h"),
			communityCards: CardsFromStrings("8s 7s 2d"),
			expectedEquity: 0.36, // 9 outs * 4 = 36%
		},
		{
			name:           "OESD on Turn",
			holeCards:      CardsFromStrings("8s 7s Kc"),
			communityCards: CardsFromStrings("6c 5h 2d 2h"),
			expectedEquity: 0.16, // 8 outs * 2 = 16%
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := CalculateEquityWithCards(tc.holeCards, tc.communityCards)
			if actual != tc.expectedEquity {
				t.Errorf("Expected equity to be %.2f, but got %.2f", tc.expectedEquity, actual)
			}
		})
	}
}
