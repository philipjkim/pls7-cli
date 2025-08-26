package poker

import (
	"fmt"
	"pls7-cli/internal/util"
	"testing"
)

func TestNLHCalculateOuts(t *testing.T) {
	util.InitLogger(true)
	testCases := []struct {
		name                string
		holeCards           []Card
		communityCards      []Card
		expectedAllOuts     []Card
		expectedOutsPerRank map[HandRank][]Card
	}{
		{
			name:            "Open-ended Straight Draw",
			holeCards:       CardsFromStrings("8s 7s"),
			communityCards:  CardsFromStrings("6c 5h 2d"),
			expectedAllOuts: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "OESD with Ace high",
			holeCards:       CardsFromStrings("As Kh"),
			communityCards:  CardsFromStrings("Qh Jc 2d"),
			expectedAllOuts: CardsFromStrings("Ts Th Td Tc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("Ts Th Td Tc"),
			},
		},
		{
			name:            "OESD with Ace low",
			holeCards:       CardsFromStrings("4s 3d"),
			communityCards:  CardsFromStrings("As Qc 2c"),
			expectedAllOuts: CardsFromStrings("5s 5h 5d 5c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("5s 5h 5d 5c"),
			},
		},
		{
			name:            "Gutshot 8654",
			holeCards:       CardsFromStrings("8s 6s"),
			communityCards:  CardsFromStrings("Ad 5h 4h"),
			expectedAllOuts: CardsFromStrings("7s 7h 7d 7c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("7s 7h 7d 7c"),
			},
		},
		{
			name:            "Gutshot with Ace high",
			holeCards:       CardsFromStrings("As Qd"),
			communityCards:  CardsFromStrings("Jh Tc 3d"),
			expectedAllOuts: CardsFromStrings("Ks Kh Kd Kc"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("Ks Kh Kd Kc"),
			},
		},
		{
			name:            "Gutshot with Ace low",
			holeCards:       CardsFromStrings("4h 3c"),
			communityCards:  CardsFromStrings("Ah 8c 5d"),
			expectedAllOuts: CardsFromStrings("2s 2h 2d 2c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight: CardsFromStrings("2s 2h 2d 2c"),
			},
		},
		{
			name:            "Flush-only Draw",
			holeCards:       CardsFromStrings("As Js"),
			communityCards:  CardsFromStrings("8s 7s 2d"),
			expectedAllOuts: CardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush: CardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
			},
		},
		{
			name:            "Straight or Flush Draw",
			holeCards:       CardsFromStrings("8s 7s"),
			communityCards:  CardsFromStrings("6s 5s 2d"),
			expectedAllOuts: CardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s 9h 9d 9c 4h 4d 4c"),
			expectedOutsPerRank: map[HandRank][]Card{
				Flush:    CardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s"),
				Straight: CardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
			},
		},
		{
			name:            "Triple Draw with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h"),
			communityCards:  CardsFromStrings("As Js 2d"),
			expectedAllOuts: CardsFromStrings("8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				ThreeOfAKind: CardsFromStrings("8d 8c"),
			},
		},
		{
			name:                "Non-Draw for Triple because board is paired, not pocket pair",
			holeCards:           CardsFromStrings("8s 6d"),
			communityCards:      CardsFromStrings("As Jh Jd"),
			expectedAllOuts:     []Card{},
			expectedOutsPerRank: map[HandRank][]Card{},
		},
		{
			name:            "Full House Draw from Two Pair with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h"),
			communityCards:  CardsFromStrings("As Ah 5d"),
			expectedAllOuts: CardsFromStrings("Ad Ac 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: CardsFromStrings("Ad Ac 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips with Pocket Pair",
			holeCards:       CardsFromStrings("8s 8h"),
			communityCards:  CardsFromStrings("As 8d 5d"),
			expectedAllOuts: CardsFromStrings("Ah Ad Ac 5s 5h 5c 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   CardsFromStrings("Ah Ad Ac 5s 5h 5c"),
				FourOfAKind: CardsFromStrings("8c"),
			},
		},
		{
			name:            "Full House Draw from Two Pair without pocket pair",
			holeCards:       CardsFromStrings("As 8h"),
			communityCards:  CardsFromStrings("Ac 8s 5d"),
			expectedAllOuts: CardsFromStrings("Ad Ah 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse: CardsFromStrings("Ad Ah 8d 8c"),
			},
		},
		{
			name:            "Full House and Quad Draw from Trips without pocket pair",
			holeCards:       CardsFromStrings("As 8h"),
			communityCards:  CardsFromStrings("Ah Ad 5d"),
			expectedAllOuts: CardsFromStrings("Ac 5s 5h 5c 8s 8d 8c"),
			expectedOutsPerRank: map[HandRank][]Card{
				FullHouse:   CardsFromStrings("5s 5h 5c 8s 8d 8c"),
				FourOfAKind: CardsFromStrings("Ac"),
			},
		},
		{
			name:            "Straight Flush Draw (Open-ended)",
			holeCards:       CardsFromStrings("8s 7s"),
			communityCards:  CardsFromStrings("6s 5s 2d"),
			expectedAllOuts: CardsFromStrings("Js 4d 4c 9h 2s Qs Ks 9c 4s 9s Ts As 4h 9d 3s"),
			expectedOutsPerRank: map[HandRank][]Card{
				Straight:      CardsFromStrings("4s 4h 4d 4c 9s 9h 9d 9c"),
				Flush:         CardsFromStrings("2s 3s 4s 9s Ts Js Qs Ks As"),
				StraightFlush: CardsFromStrings("9s 4s"),
			},
		},
	}

	gameRules := &GameRules{
		LowHand: LowHandRules{Enabled: false, MaxRank: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

func TestNLHCalculateBreakEvenEquityBasedOnPotOdds(t *testing.T) {
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

func TestNLHCalculateEquityWithCards(t *testing.T) {
	util.InitLogger(true)
	testCases := []struct {
		name           string
		holeCards      []Card
		communityCards []Card
		expectedEquity float64
	}{
		{
			name:           "Flush Draw on Flop",
			holeCards:      CardsFromStrings("As Js"),
			communityCards: CardsFromStrings("8s 7s 2d"),
			expectedEquity: 0.36, // 9 outs * 4 = 36%
		},
		{
			name:           "OESD on Turn",
			holeCards:      CardsFromStrings("8s 7s"),
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
