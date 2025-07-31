package poker

import (
	"testing"
)

func TestCalculateOuts(t *testing.T) {
	testCases := []struct {
		name           string
		holeCards      []Card
		communityCards []Card
		lowlessMode    bool
		expectedOuts   int
	}{
		{
			name:           "Open-ended Straight Draw",
			holeCards:      cardsFromStrings("8s 7s"),
			communityCards: cardsFromStrings("6c 5h 2d"),
			lowlessMode:    true,
			expectedOuts:   8, // 4 nines, 4 fours
		},
		{
			name:           "Flush Draw",
			holeCards:      cardsFromStrings("As Ks"),
			communityCards: cardsFromStrings("Qs Js 2s"),
			lowlessMode:    true,
			expectedOuts:   9, // 9 remaining spades
		},
		{
			name:           "Straight Flush Draw",
			holeCards:      cardsFromStrings("8s 7s"),
			communityCards: cardsFromStrings("6s 5s 2d"),
			lowlessMode:    true,
			expectedOuts:   15, // 9 spades for flush, 6 non-spade cards for straight (4s, 4h, 4d, 9s, 9h, 9d)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outs := CalculateOuts(tc.holeCards, tc.communityCards, tc.lowlessMode)
			if outs != tc.expectedOuts {
				t.Errorf("Expected %d outs, but got %d", tc.expectedOuts, outs)
			}
		})
	}
}
