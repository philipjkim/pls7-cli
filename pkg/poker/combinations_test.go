package poker

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

// cardSliceSorter implements sort.Interface for [][]Card to allow for deterministic sorting.
type cardSliceSorter [][]Card

func (s cardSliceSorter) Len() int      { return len(s) }
func (s cardSliceSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s cardSliceSorter) Less(i, j int) bool {
	// This implementation compares card slices lexicographically in descending order
	// to match the test case expectations.
	sliceI, sliceJ := s[i], s[j]
	lenI, lenJ := len(sliceI), len(sliceJ)
	minLen := lenI
	if lenJ < minLen {
		minLen = lenJ
	}

	for k := 0; k < minLen; k++ {
		cardI, cardJ := sliceI[k], sliceJ[k]
		if cardI.Rank != cardJ.Rank {
			return cardI.Rank > cardJ.Rank
		}
		if cardI.Suit != cardJ.Suit {
			return cardI.Suit > cardJ.Suit
		}
	}

	return lenI > lenJ
}

func TestCombinations(t *testing.T) {
	cards := []Card{
		{Suit: Heart, Rank: Ace},
		{Suit: Diamond, Rank: King},
		{Suit: Club, Rank: Queen},
		{Suit: Spade, Rank: Jack},
	}

	testCases := []struct {
		name     string
		n        int
		expected [][]Card
	}{
		{
			name: "Combinations of 2",
			n:    2,
			expected: [][]Card{
				{{Suit: Heart, Rank: Ace}, {Suit: Diamond, Rank: King}},
				{{Suit: Heart, Rank: Ace}, {Suit: Club, Rank: Queen}},
				{{Suit: Heart, Rank: Ace}, {Suit: Spade, Rank: Jack}},
				{{Suit: Diamond, Rank: King}, {Suit: Club, Rank: Queen}},
				{{Suit: Diamond, Rank: King}, {Suit: Spade, Rank: Jack}},
				{{Suit: Club, Rank: Queen}, {Suit: Spade, Rank: Jack}},
			},
		},
		{
			name: "Combinations of 3",
			n:    3,
			expected: [][]Card{
				{{Suit: Heart, Rank: Ace}, {Suit: Diamond, Rank: King}, {Suit: Club, Rank: Queen}},
				{{Suit: Heart, Rank: Ace}, {Suit: Diamond, Rank: King}, {Suit: Spade, Rank: Jack}},
				{{Suit: Heart, Rank: Ace}, {Suit: Club, Rank: Queen}, {Suit: Spade, Rank: Jack}},
				{{Suit: Diamond, Rank: King}, {Suit: Club, Rank: Queen}, {Suit: Spade, Rank: Jack}},
			},
		},
		{
			name:     "Combinations of 1",
			n:        1,
			expected: [][]Card{{{Suit: Heart, Rank: Ace}}, {{Suit: Diamond, Rank: King}}, {{Suit: Club, Rank: Queen}}, {{Suit: Spade, Rank: Jack}}},
		},
		{
			name:     "Combinations of 4",
			n:        4,
			expected: [][]Card{cards},
		},
		{
			name:     "N greater than length",
			n:        5,
			expected: nil,
		},
		{
			name:     "N is zero",
			n:        0,
			expected: [][]Card{{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := combinations(cards, tc.n)

			// Sort both slices for deterministic comparison
			sort.Sort(cardSliceSorter(result))
			sort.Sort(cardSliceSorter(tc.expected))

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
