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

	// --- High Hand Evaluation ---
	// This part remains the same. We find the best high hand first.
	// (The code is shortened for brevity, but it's the same as the previous step)
	if sfCards, ok := findStraightFlush(analysis); ok {
		rank := StraightFlush
		if sfCards[0].Rank == Ace {
			rank = RoyalFlush
		}
		highResult = &HandResult{Rank: rank, Cards: sfCards, HighValues: []Rank{sfCards[0].Rank}}
	} else if qpCards, ok := findQuadPair(analysis); ok {
		highResult = &HandResult{Rank: QuadPair, Cards: qpCards}
	} else if dtCards, ok := findDoubleTriple(analysis); ok {
		highResult = &HandResult{Rank: DoubleTriple, Cards: dtCards}
	} else if quadRank, ok := findBestNOfAKind(analysis.rankCounts, 4); ok {
		kickers := findKickers(analysis.cards, []Rank{quadRank}, 1)
		quadCards := findCardsByRank(pool, quadRank, 4)
		highResult = &HandResult{Rank: FourOfAKind, Cards: append(quadCards, kickers...), HighValues: []Rank{quadRank, kickers[0].Rank}}
	} else if tripleRank, pairRank, ok := findBestFullHouse(analysis.rankCounts); ok {
		tripleCards := findCardsByRank(pool, tripleRank, 3)
		pairCards := findCardsByRank(pool, pairRank, 2)
		highResult = &HandResult{Rank: FullHouse, Cards: append(tripleCards, pairCards...), HighValues: []Rank{tripleRank, pairRank}}
	} else if flushCards, ok := findBestFlush(analysis); ok {
		highResult = &HandResult{Rank: Flush, Cards: flushCards, HighValues: []Rank{flushCards[0].Rank, flushCards[1].Rank, flushCards[2].Rank, flushCards[3].Rank, flushCards[4].Rank}}
	} else if ssCards, ok := findSkipStraight(analysis); ok {
		highResult = &HandResult{Rank: SkipStraight, Cards: ssCards, HighValues: []Rank{ssCards[0].Rank}}
	} else if straightCards, ok := findBestStraight(analysis); ok {
		highResult = &HandResult{Rank: Straight, Cards: straightCards, HighValues: []Rank{straightCards[0].Rank}}
	} else if tpCards, ok := findTriPair(analysis); ok {
		highResult = &HandResult{Rank: TriPair, Cards: tpCards}
	} else if tripleRank, ok := findBestNOfAKind(analysis.rankCounts, 3); ok {
		kickers := findKickers(analysis.cards, []Rank{tripleRank}, 2)
		tripleCards := findCardsByRank(pool, tripleRank, 3)
		highResult = &HandResult{Rank: ThreeOfAKind, Cards: append(tripleCards, kickers...), HighValues: []Rank{tripleRank, kickers[0].Rank, kickers[1].Rank}}
	} else if highPair, lowPair, ok := findBestTwoPair(analysis.rankCounts); ok {
		kickers := findKickers(analysis.cards, []Rank{highPair, lowPair}, 1)
		highPairCards := findCardsByRank(pool, highPair, 2)
		lowPairCards := findCardsByRank(pool, lowPair, 2)
		allCards := append(highPairCards, lowPairCards...)
		allCards = append(allCards, kickers...)
		highResult = &HandResult{Rank: TwoPair, Cards: allCards, HighValues: []Rank{highPair, lowPair, kickers[0].Rank}}
	} else if pairRank, ok := findBestNOfAKind(analysis.rankCounts, 2); ok {
		kickers := findKickers(analysis.cards, []Rank{pairRank}, 3)
		pairCards := findCardsByRank(pool, pairRank, 2)
		highResult = &HandResult{Rank: OnePair, Cards: append(pairCards, kickers...), HighValues: []Rank{pairRank, kickers[0].Rank, kickers[1].Rank, kickers[2].Rank}}
	} else {
		highResult = &HandResult{Rank: HighCard, Cards: analysis.cards[:5], HighValues: []Rank{analysis.cards[0].Rank, analysis.cards[1].Rank, analysis.cards[2].Rank, analysis.cards[3].Rank, analysis.cards[4].Rank}}
	}

	// --- Low Hand Evaluation ---
	if lowHand, ok := findBestLowHand(analysis); ok {
		lowResult = lowHand
	}

	return highResult, lowResult
}

// --- New Helper Function for Low Hand ---

func findBestLowHand(analysis *handAnalysis) (*HandResult, bool) {
	uniqueLowCards := make([]Card, 0, 8)
	usedRanks := make(map[Rank]bool)

	// Find all unique cards with rank 7 or lower, treating Ace as a low card.
	for _, card := range analysis.cards {
		// FIX: Ace must be included as a low card candidate.
		isLowCard := card.Rank <= Seven || card.Rank == Ace
		if isLowCard && !usedRanks[card.Rank] {
			uniqueLowCards = append(uniqueLowCards, card)
			usedRanks[card.Rank] = true
		}
	}

	// A low hand must have at least 5 unique cards of rank 7 or lower
	if len(uniqueLowCards) < 5 {
		return nil, false
	}

	// Sort the unique low cards by rank ascending to find the best (lowest) hand
	sort.Slice(uniqueLowCards, func(i, j int) bool {
		// Special handling for Ace as the lowest card
		rankI, rankJ := uniqueLowCards[i].Rank, uniqueLowCards[j].Rank
		if rankI == Ace {
			rankI = 1
		}
		if rankJ == Ace {
			rankJ = 1
		}
		return rankI < rankJ
	})

	// The best low hand consists of the 5 lowest cards
	bestLowCards := uniqueLowCards[:5]

	// Sort descending for HighValues comparison (e.g., 7-5-4-3-2 is better than 7-6-3-2-A)
	sort.Slice(bestLowCards, func(i, j int) bool {
		return bestLowCards[i].Rank > bestLowCards[j].Rank
	})

	return &HandResult{
		Rank:       HighCard, // Rank is not relevant for low hands in this context
		Cards:      bestLowCards,
		HighValues: []Rank{bestLowCards[0].Rank, bestLowCards[1].Rank, bestLowCards[2].Rank, bestLowCards[3].Rank, bestLowCards[4].Rank},
	}, true
}

// --- Existing Helper Functions ---

func findStraightFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			flushAnalysis := newHandAnalysis(flushCards)
			if sfCards, ok := findBestStraight(flushAnalysis); ok {
				return sfCards, true
			}
		}
	}
	return nil, false
}

func findQuadPair(analysis *handAnalysis) ([]Card, bool) {
	pairCount := 0
	for _, count := range analysis.rankCounts {
		if count == 2 {
			pairCount++
		}
	}
	if pairCount == 4 {
		return analysis.cards, true
	}
	return nil, false
}

func findDoubleTriple(analysis *handAnalysis) ([]Card, bool) {
	tripleRanks := []Rank{}
	for rank, count := range analysis.rankCounts {
		if count >= 3 {
			tripleRanks = append(tripleRanks, rank)
		}
	}
	if len(tripleRanks) >= 2 {
		sort.Slice(tripleRanks, func(i, j int) bool { return tripleRanks[i] > tripleRanks[j] })
		cards := make([]Card, 0, 6)
		cards = append(cards, findCardsByRank(analysis.cards, tripleRanks[0], 3)...)
		cards = append(cards, findCardsByRank(analysis.cards, tripleRanks[1], 3)...)
		return cards, true
	}
	return nil, false
}

func findSkipStraight(analysis *handAnalysis) ([]Card, bool) {
	uniqueRanks := make([]Rank, 0, len(analysis.rankCounts))
	for rank := range analysis.rankCounts {
		uniqueRanks = append(uniqueRanks, rank)
	}
	sort.Slice(uniqueRanks, func(i, j int) bool { return uniqueRanks[i] > uniqueRanks[j] })
	for i := 0; i <= len(uniqueRanks)-5; i++ {
		isSkipStraight := true
		for j := 0; j < 4; j++ {
			if uniqueRanks[i+j] != uniqueRanks[i+j+1]+2 {
				isSkipStraight = false
				break
			}
		}
		if isSkipStraight {
			topRank := uniqueRanks[i]
			ranks := []Rank{topRank, topRank - 2, topRank - 4, topRank - 6, topRank - 8}
			return findCardsForStraight(analysis.cards, ranks), true
		}
	}
	return nil, false
}

func findTriPair(analysis *handAnalysis) ([]Card, bool) {
	pairRanks := []Rank{}
	for rank, count := range analysis.rankCounts {
		if count >= 2 {
			pairRanks = append(pairRanks, rank)
		}
	}
	if len(pairRanks) >= 3 {
		sort.Slice(pairRanks, func(i, j int) bool { return pairRanks[i] > pairRanks[j] })
		cards := make([]Card, 0, 6)
		cards = append(cards, findCardsByRank(analysis.cards, pairRanks[0], 2)...)
		cards = append(cards, findCardsByRank(analysis.cards, pairRanks[1], 2)...)
		cards = append(cards, findCardsByRank(analysis.cards, pairRanks[2], 2)...)
		return cards, true
	}
	return nil, false
}

func findBestFullHouse(rankCounts map[Rank]int) (Rank, Rank, bool) {
	var bestTripleRank Rank = -1
	var bestPairRank Rank = -1
	for rank, count := range rankCounts {
		if count >= 3 {
			if rank > bestTripleRank {
				bestTripleRank = rank
			}
		}
	}
	if bestTripleRank == -1 {
		return -1, -1, false
	}
	for rank, count := range rankCounts {
		if count >= 2 && rank != bestTripleRank {
			if rank > bestPairRank {
				bestPairRank = rank
			}
		}
	}
	if bestPairRank == -1 {
		return -1, -1, false
	}
	return bestTripleRank, bestPairRank, true
}

func findBestFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			return flushCards[:5], true
		}
	}
	return nil, false
}

func findBestStraight(analysis *handAnalysis) ([]Card, bool) {
	uniqueRanks := make([]Rank, 0, len(analysis.rankCounts))
	for rank := range analysis.rankCounts {
		uniqueRanks = append(uniqueRanks, rank)
	}
	sort.Slice(uniqueRanks, func(i, j int) bool { return uniqueRanks[i] > uniqueRanks[j] })
	if uniqueRanks[0] == Ace && containsRank(uniqueRanks, Five) && containsRank(uniqueRanks, Four) && containsRank(uniqueRanks, Three) && containsRank(uniqueRanks, Two) {
		return findCardsForStraight(analysis.cards, []Rank{Five, Four, Three, Two, Ace}), true
	}
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

func containsRank(ranks []Rank, target Rank) bool {
	for _, r := range ranks {
		if r == target {
			return true
		}
	}
	return false
}

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
	sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
	return pairs[0], pairs[1], true
}

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
