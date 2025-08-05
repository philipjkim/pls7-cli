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
		lowlessMode         bool
		expectedAllOuts     []Card
		expectedOutsPerRank map[HandRank][]Card
	}{
		{
			name:            "Open-ended Straight Draw",
			holeCards:       cardsFromStrings("8s 7s Kc"),
			communityCards:  cardsFromStrings("6c 5h 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "OESD with Ace high",
			holeCards:       cardsFromStrings("As Kh Qs"),
			communityCards:  cardsFromStrings("Jh 5c 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ts Th Td Tc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("Ts Th Td Tc"),
			},
		},
		{
			name:            "OESD with Ace low",
			holeCards:       cardsFromStrings("4s 3d 2h"),
			communityCards:  cardsFromStrings("As Qc Tc"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("5s 5h 5d 5c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("5s 5h 5d 5c"),
			},
		},
		{
			name:            "Gutshot 8654",
			holeCards:       cardsFromStrings("8s 6s 5c"),
			communityCards:  cardsFromStrings("Ad Kh 4h"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("7s 7h 7d 7c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("7s 7h 7d 7c"),
			},
		},
		{
			name:            "Gutshot with Ace high",
			holeCards:       cardsFromStrings("As Qd Jc"),
			communityCards:  cardsFromStrings("Th 7c 3d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ks Kh Kd Kc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("Ks Kh Kd Kc"),
			},
		},
		{
			name:            "Gutshot with Ace low",
			holeCards:       cardsFromStrings("4h 3c As"),
			communityCards:  cardsFromStrings("Jh 8c 5d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("2s 2h 2d 2c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: cardsFromStrings("2s 2h 2d 2c"),
			},
		},
		{
			name:            "Flush-only Draw",
			holeCards:       cardsFromStrings("As Js 5h"),
			communityCards:  cardsFromStrings("8s 7s 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush: cardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			},
		},
		{
			name:            "Straight or Flush Draw",
			holeCards:       cardsFromStrings("8s 7s Kc"),
			communityCards:  cardsFromStrings("6s 5s 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s 9h 9d 9c 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush:    cardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s"),
				Straight: cardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "Triple Draw with Pocket Pair",
			holeCards:       cardsFromStrings("8s 8h 6c"),
			communityCards:  cardsFromStrings("As Js 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				ThreeOfAKind: cardsFromStrings("8d 8c"),
			},
		},
		{
			name:                "Non-Draw for Triple because board is paired, not pocket pair",
			holeCards:           cardsFromStrings("8s 6d 5c"),
			communityCards:      cardsFromStrings("As Jh Jd"),
			lowlessMode:         true,
			expectedAllOuts:     []Card{},
			expectedOutsPerRank: map[HandRank][]Card{},
		},
		{
			name:            "Full House Draw from Two Pair with Pocket Pair",
			holeCards:       cardsFromStrings("8s 8h 6c"),
			communityCards:  cardsFromStrings("As Ah 5d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ad Ac 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: cardsFromStrings("Ad Ac 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips with Pocket Pair",
			holeCards:       cardsFromStrings("8s 8h 6c"),
			communityCards:  cardsFromStrings("As 8d 5d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ah Ad Ac 6s 6h 6d 5s 5h 5c 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   cardsFromStrings("Ah Ad Ac 6s 6h 6d 5s 5h 5c"),
				FourOfAKind: cardsFromStrings("8c"),
			},
		},
		{
			name:            "Full House Draw from Two Pair without pocket pair",
			holeCards:       cardsFromStrings("As 8h 6c"),
			communityCards:  cardsFromStrings("Ac 8s 5d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ad Ah 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: cardsFromStrings("Ad Ah 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips without pocket pair",
			holeCards:       cardsFromStrings("As 8h 6c"),
			communityCards:  cardsFromStrings("Ah Ad 5d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ac 6s 6h 6d 5s 5h 5c 8s 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   cardsFromStrings("6s 6h 6d 5s 5h 5c 8s 8d 8c"),
				FourOfAKind: cardsFromStrings("Ac"),
			},
		},
		{
			name:            "Skip Straight Draw (Gutshot)",
			holeCards:       cardsFromStrings("8s 6s 3c"),
			communityCards:  cardsFromStrings("Ad Qh 4h"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Ts Th Td Tc"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraight: cardsFromStrings("Ts Th Td Tc"),
			},
		},
		{
			name:            "Skip Straight Draw (Open-ended)",
			holeCards:       cardsFromStrings("3s 5s 7c"),
			communityCards:  cardsFromStrings("9d Qh Qh"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("As Ah Ad Ac Js Jh Jd Jc"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraight: cardsFromStrings("As Ah Ad Ac Js Jh Jd Jc"),
			},
		},
		{
			name:            "Straight Flush Draw (Open-ended)",
			holeCards:       cardsFromStrings("8s 7s Kc"),
			communityCards:  cardsFromStrings("6s 5s 2d"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Js 4d 4c 9h 2s Qs Ks 9c 4s 9s Ts As 4h 9d 3s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight:      cardsFromStrings("4s 4h 4d 4c 9s 9h 9d 9c"),
				Flush:         cardsFromStrings("2s 3s 4s 9s Ts Js Qs Ks As"),
				StraightFlush: cardsFromStrings("9s 4s"),
			},
		},
		{
			name:            "Skip Straight Flush Draw (Open-ended)",
			holeCards:       cardsFromStrings("Ts 8s 6s"),
			communityCards:  cardsFromStrings("Kh Kd 4s"),
			lowlessMode:     true,
			expectedAllOuts: cardsFromStrings("Qs Qh Qd Qc 2s 2h 2d 2c 3s 5s 7s 9s Js Ks As"),
			expectedOutsPerRank: map[HandRank][]Card{
				SkipStraightFlush: cardsFromStrings("Qs 2s"),
				Flush:             cardsFromStrings("2s 3s 5s 7s 9s Js Qs Ks As"),
				SkipStraight:      cardsFromStrings("2s 2h 2d 2c Qs Qh Qd Qc"),
			},
		},
		{
			name:            "Low Hand Draw",
			holeCards:       cardsFromStrings("2s 3c 6h"),
			communityCards:  cardsFromStrings("Kh Kd 7s"),
			lowlessMode:     false,
			expectedAllOuts: cardsFromStrings("As Ah Ad Ac 4s 4h 4d 4c 5s 5h 5d 5c"),
			expectedOutsPerRank: map[HandRank][]Card{
				HighCard: cardsFromStrings("As Ah Ad Ac 4s 4h 4d 4c 5s 5h 5d 5c"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasOuts, outsInfo := CalculateOuts(tc.holeCards, tc.communityCards, tc.lowlessMode)
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
	testCases := []struct {
		name           string
		holeCards      []Card
		communityCards []Card
		expectedEquity float64
	}{
		{
			name:           "Flush Draw on Flop",
			holeCards:      cardsFromStrings("As Js 5h"),
			communityCards: cardsFromStrings("8s 7s 2d"),
			expectedEquity: 0.36, // 9 outs * 4 = 36%
		},
		{
			name:           "OESD on Turn",
			holeCards:      cardsFromStrings("8s 7s Kc"),
			communityCards: cardsFromStrings("6c 5h 2d 2h"),
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
