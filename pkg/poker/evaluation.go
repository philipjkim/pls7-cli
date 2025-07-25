package poker

import (
	"sort"
)

// HandRank defines the ranking of a poker hand.
// The order is important, from lowest to highest rank.
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	TriPair // PLS7 Special
	ThreeOfAKind
	Straight
	SkipStraight // PLS7 Special
	Flush
	FullHouse
	DoubleTriple // PLS7 Special
	FourOfAKind
	QuadPair // PLS7 Special
	StraightFlush
	RoyalFlush
)

// HandResult stores the outcome of a hand evaluation.
type HandResult struct {
	Rank       HandRank
	Cards      []Card
	HighValues []Rank // For tie-breaking (e.g., [Ace, King] for A-high flush)
}

// handAnalysis is a helper struct to hold counts of ranks and suits.
type handAnalysis struct {
	rankCounts map[Rank]int
	suitCounts map[Suit]int
	cards      []Card // Original 8 cards, sorted by rank descending
}

// newHandAnalysis creates an analysis object from an 8-card pool.
func newHandAnalysis(pool []Card) *handAnalysis {
	analysis := &handAnalysis{
		rankCounts: make(map[Rank]int),
		suitCounts: make(map[Suit]int),
		cards:      make([]Card, len(pool)),
	}
	copy(analysis.cards, pool)

	sort.Slice(analysis.cards, func(i, j int) bool {
		return analysis.cards[i].Rank > analysis.cards[j].Rank
	})

	for _, c := range analysis.cards {
		analysis.rankCounts[c.Rank]++
		analysis.suitCounts[c.Suit]++
	}
	return analysis
}

// EvaluateHand analyzes a full 8-card pool and determines the best high and low hands.
func EvaluateHand(holeCards []Card, communityCards []Card) (highResult *HandResult, lowResult *HandResult) {
	pool := make([]Card, 0, 8)
	pool = append(pool, holeCards...)
	pool = append(pool, communityCards...)

	analysis := newHandAnalysis(pool)

	// Check for hands in descending order of rank

	// Check for Four of a Kind
	if quadRank, ok := findBestNOfAKind(analysis.rankCounts, 4); ok {
		kickers := findKickers(analysis.cards, []Rank{quadRank}, 1)
		quadCards := findCardsByRank(pool, quadRank, 4)

		highResult = &HandResult{
			Rank:       FourOfAKind,
			Cards:      append(quadCards, kickers...),
			HighValues: []Rank{quadRank, kickers[0].Rank},
		}
		return highResult, nil
	}

	// Check for Full House
	if tripleRank, pairRank, ok := findBestFullHouse(analysis.rankCounts); ok {
		tripleCards := findCardsByRank(pool, tripleRank, 3)
		pairCards := findCardsByRank(pool, pairRank, 2)

		highResult = &HandResult{
			Rank:       FullHouse,
			Cards:      append(tripleCards, pairCards...),
			HighValues: []Rank{tripleRank, pairRank},
		}
		return highResult, nil
	}

	// Check for Flush
	if flushCards, ok := findBestFlush(analysis); ok {
		highResult = &HandResult{
			Rank:       Flush,
			Cards:      flushCards,
			HighValues: []Rank{flushCards[0].Rank, flushCards[1].Rank, flushCards[2].Rank, flushCards[3].Rank, flushCards[4].Rank},
		}
		return highResult, nil
	}

	// Check for Straight
	if straightCards, ok := findBestStraight(analysis); ok {
		highResult = &HandResult{
			Rank:       Straight,
			Cards:      straightCards,
			HighValues: []Rank{straightCards[0].Rank},
		}
		return highResult, nil
	}

	// Check for Three of a Kind
	if tripleRank, ok := findBestNOfAKind(analysis.rankCounts, 3); ok {
		kickers := findKickers(analysis.cards, []Rank{tripleRank}, 2)
		tripleCards := findCardsByRank(pool, tripleRank, 3)

		highResult = &HandResult{
			Rank:       ThreeOfAKind,
			Cards:      append(tripleCards, kickers...),
			HighValues: []Rank{tripleRank, kickers[0].Rank, kickers[1].Rank},
		}
		return highResult, nil
	}

	// Check for Two Pair
	if highPair, lowPair, ok := findBestTwoPair(analysis.rankCounts); ok {
		kickers := findKickers(analysis.cards, []Rank{highPair, lowPair}, 1)
		highPairCards := findCardsByRank(pool, highPair, 2)
		lowPairCards := findCardsByRank(pool, lowPair, 2)

		allCards := append(highPairCards, lowPairCards...)
		allCards = append(allCards, kickers...)

		highResult = &HandResult{
			Rank:       TwoPair,
			Cards:      allCards,
			HighValues: []Rank{highPair, lowPair, kickers[0].Rank},
		}
		return highResult, nil
	}

	// Check for One Pair
	if pairRank, ok := findBestNOfAKind(analysis.rankCounts, 2); ok {
		kickers := findKickers(analysis.cards, []Rank{pairRank}, 3)
		pairCards := findCardsByRank(pool, pairRank, 2)

		highResult = &HandResult{
			Rank:       OnePair,
			Cards:      append(pairCards, kickers...),
			HighValues: []Rank{pairRank, kickers[0].Rank, kickers[1].Rank, kickers[2].Rank},
		}
		return highResult, nil
	}

	// Default to High Card
	highResult = &HandResult{
		Rank:       HighCard,
		Cards:      analysis.cards[:5],
		HighValues: []Rank{analysis.cards[0].Rank, analysis.cards[1].Rank, analysis.cards[2].Rank, analysis.cards[3].Rank, analysis.cards[4].Rank},
	}

	return highResult, nil
}

// findBestFullHouse checks for the best full house.
func findBestFullHouse(rankCounts map[Rank]int) (Rank, Rank, bool) {
	var bestTripleRank Rank = -1
	var bestPairRank Rank = -1

	// Find the highest triple
	for rank, count := range rankCounts {
		if count >= 3 {
			if rank > bestTripleRank {
				bestTripleRank = rank
			}
		}
	}

	if bestTripleRank == -1 {
		return -1, -1, false // No triple found
	}

	// Find the highest pair among the remaining cards
	for rank, count := range rankCounts {
		if count >= 2 && rank != bestTripleRank {
			if rank > bestPairRank {
				bestPairRank = rank
			}
		}
	}

	if bestPairRank == -1 {
		return -1, -1, false // No pair found to go with the triple
	}

	return bestTripleRank, bestPairRank, true
}

// findBestFlush checks for a flush.
func findBestFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards { // analysis.cards is already sorted
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			return flushCards[:5], true
		}
	}
	return nil, false
}

// findBestStraight checks for a straight.
func findBestStraight(analysis *handAnalysis) ([]Card, bool) {
	uniqueRanks := make([]Rank, 0, len(analysis.rankCounts))
	for rank := range analysis.rankCounts {
		uniqueRanks = append(uniqueRanks, rank)
	}
	sort.Slice(uniqueRanks, func(i, j int) bool {
		return uniqueRanks[i] > uniqueRanks[j]
	})

	// Check for wheel straight (A-2-3-4-5)
	hasAce := uniqueRanks[0] == Ace
	if hasAce && containsRank(uniqueRanks, Five) && containsRank(uniqueRanks, Four) && containsRank(uniqueRanks, Three) && containsRank(uniqueRanks, Two) {
		return findCardsForStraight(analysis.cards, []Rank{Five, Four, Three, Two, Ace}), true
	}

	// Check for other straights
	for i := 0; i <= len(uniqueRanks)-5; i++ {
		isStraight := true
		for j := 0; j < 4; j++ {
			if uniqueRanks[i+j] != uniqueRanks[i+j+1]+1 {
				isStraight = false
				break
			}
		}
		if isStraight {
			topRank := uniqueRanks[i]
			ranks := []Rank{topRank, topRank - 1, topRank - 2, topRank - 3, topRank - 4}
			return findCardsForStraight(analysis.cards, ranks), true
		}
	}

	return nil, false
}

// findCardsForStraight reconstructs the straight hand from the pool.
func findCardsForStraight(pool []Card, ranks []Rank) []Card {
	straightCards := make([]Card, 0, 5)
	usedRanks := make(map[Rank]bool)
	for _, rank := range ranks {
		for _, card := range pool {
			if card.Rank == rank && !usedRanks[rank] {
				straightCards = append(straightCards, card)
				usedRanks[rank] = true
				break
			}
		}
	}
	return straightCards
}

// containsRank is a helper to check for a rank in a slice.
func containsRank(ranks []Rank, target Rank) bool {
	for _, r := range ranks {
		if r == target {
			return true
		}
	}
	return false
}

// findBestNOfAKind finds the highest-ranked N-of-a-kind (e.g., pair, triple).
func findBestNOfAKind(rankCounts map[Rank]int, n int) (Rank, bool) {
	bestRank := Rank(-1)
	found := false
	for rank, count := range rankCounts {
		if count >= n {
			if rank > bestRank {
				bestRank = rank
				found = true
			}
		}
	}
	return bestRank, found
}

// findBestTwoPair finds the two highest-ranked pairs.
func findBestTwoPair(rankCounts map[Rank]int) (Rank, Rank, bool) {
	pairs := []Rank{}
	for rank, count := range rankCounts {
		if count >= 2 {
			pairs = append(pairs, rank)
		}
	}
	if len(pairs) < 2 {
		return -1, -1, false
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i] > pairs[j]
	})
	return pairs[0], pairs[1], true
}

// findCardsByRank finds the first N cards of a specific rank from a pool.
func findCardsByRank(pool []Card, rank Rank, n int) []Card {
	cards := make([]Card, 0, n)
	for _, c := range pool {
		if c.Rank == rank {
			cards = append(cards, c)
			if len(cards) == n {
				break
			}
		}
	}
	return cards
}

// findKickers finds the best N kickers, excluding certain ranks.
func findKickers(sortedPool []Card, excludeRanks []Rank, n int) []Card {
	kickers := make([]Card, 0, n)
	excludeMap := make(map[Rank]bool)
	for _, r := range excludeRanks {
		excludeMap[r] = true
	}

	for _, c := range sortedPool {
		if !excludeMap[c.Rank] {
			kickers = append(kickers, c)
			if len(kickers) == n {
				break
			}
		}
	}
	return kickers
}
