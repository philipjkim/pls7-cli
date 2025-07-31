package poker

import (
	"pls7-cli/internal/util"
	"sort"
	"testing"
)

func TestCalculateOuts(t *testing.T) {
	util.InitLogger(true)
	testCases := []struct {
		name           string
		holeCards      []Card
		communityCards []Card
		lowlessMode    bool
		expectedOuts   []Card
	}{
		{
			name:           "Open-ended Straight Draw",
			holeCards:      cardsFromStrings("8s 7s Kc"),
			communityCards: cardsFromStrings("6c 5h 2d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("9s 9h 9d 9c 4s 4h 4d 4c"),
		},
		{
			name:           "OESD with Ace high",
			holeCards:      cardsFromStrings("As Kh Qs"),
			communityCards: cardsFromStrings("Jh 5c 2d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("Ts Th Td Tc"),
		},
		{
			name:           "OESD with Ace low",
			holeCards:      cardsFromStrings("4s 3d 2h"),
			communityCards: cardsFromStrings("As Qc Tc"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("5s 5h 5d 5c"),
		},
		{
			name:           "Gutshot 8654",
			holeCards:      cardsFromStrings("8s 6s 5c"),
			communityCards: cardsFromStrings("Ad Qh 4h"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("7s 7h 7d 7c"),
		},
		{
			name:           "Gutshot with Ace high",
			holeCards:      cardsFromStrings("As Qd Jc"),
			communityCards: cardsFromStrings("Th 7c 6d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("Ks Kh Kd Kc"),
		},
		{
			name:           "Gutshot with Ace low",
			holeCards:      cardsFromStrings("4h 3c As"),
			communityCards: cardsFromStrings("Jh 8c 5d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("2s 2h 2d 2c"),
		},
		{
			name:           "Flush-only Draw",
			holeCards:      cardsFromStrings("As Js 5h"),
			communityCards: cardsFromStrings("8s 7s 2d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("Ks Qs Ts 9s 6s 5s 4s 3s 2s"),
		},
		{
			name:           "Straight or Flush Draw",
			holeCards:      cardsFromStrings("8s 7s Kc"),
			communityCards: cardsFromStrings("6s 5s 2d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("As Ks Qs Js Ts 9s 4s 3s 2s 9h 9d 9c 4h 4d 4c"),
		},
		{
			name:           "Triple Draw with Pocket Pair",
			holeCards:      cardsFromStrings("8s 8h 6c"),
			communityCards: cardsFromStrings("As Js 2d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("8d 8c"),
		},
		{
			name:           "Non-Draw for Triple because board is paired, not pocket pair",
			holeCards:      cardsFromStrings("8s 6d 5c"),
			communityCards: cardsFromStrings("As Jh Jd"),
			lowlessMode:    true,
			expectedOuts:   []Card{},
		},
		{
			name:           "Full House Draw from Two Pair with Pocket Pair",
			holeCards:      cardsFromStrings("8s 8h 6c"),
			communityCards: cardsFromStrings("As Ah 5d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("Ad Ac 8d 8c"),
		},
		{
			name:           "Full House and Quad Draw from Trips with Pocket Pair",
			holeCards:      cardsFromStrings("8s 8h 6c"),
			communityCards: cardsFromStrings("As 8d 5d"),
			lowlessMode:    true,
			expectedOuts:   cardsFromStrings("Ah Ad Ac 6s 6h 6d 5s 5h 5c 8c"),
		},
		{
			name:           "Full House Non-Draw from Two Pair because no pocket pair",
			holeCards:      cardsFromStrings("As 8h 6c"),
			communityCards: cardsFromStrings("Ac 8s 5d"),
			lowlessMode:    true,
			expectedOuts:   []Card{},
		},
		{
			name:           "Full House Non-Draw and Quad Non-Draw from Trips because no pocket pair",
			holeCards:      cardsFromStrings("As 8h 6c"),
			communityCards: cardsFromStrings("Ah Ad 5d"),
			lowlessMode:    true,
			expectedOuts:   []Card{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outs := CalculateOuts(tc.holeCards, tc.communityCards, tc.lowlessMode)
			if !cardSlicesEqual(outs, tc.expectedOuts) {
				t.Errorf("Expected outs %v, but got %v", tc.expectedOuts, outs)
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
