package poker

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"pls7-cli/internal/util"
	"sort"
)

// HandRank defines the ranking of a poker hand.
// The order is important, from lowest to highest rank.
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	SkipStraight // PLS7 Special
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	SkipStraightFlush // New rank, 2nd highest
	RoyalFlush        // Highest rank
)

// String makes HandRank implement the Stringer interface for easy printing.
func (hr HandRank) String() string {
	return []string{
		"High Card",
		"One Pair",
		"Two Pair",
		"Three of a Kind",
		"Straight",
		"Skip Straight",
		"Flush",
		"Full House",
		"Four of a Kind",
		"Straight Flush",
		"Skip Straight Flush",
		"Royal Flush",
	}[hr]
}

// HandResult stores the outcome of a hand evaluation.
type HandResult struct {
	Rank       HandRank
	Cards      []Card
	HighValues []Rank // For tie-breaking (e.g., [Ace, King] for A-high flush)
}

// String makes HandResult implement the Stringer interface for detailed printing.
func (hr *HandResult) String() string {
	if hr == nil {
		return "N/A"
	}

	switch hr.Rank {
	case RoyalFlush, SkipStraightFlush, StraightFlush, Straight, SkipStraight, FullHouse, Flush, OnePair:
		return fmt.Sprintf("%s, %s", hr.Rank.String(), hr.CardsString())
	case FourOfAKind:
		quadRank := hr.HighValues[0].String()
		return fmt.Sprintf("%s Four of a Kind, %s", quadRank, hr.CardsString())
	case ThreeOfAKind:
		tripleRank := hr.HighValues[0].String()
		return fmt.Sprintf("%s Three of a Kind, %s", tripleRank, hr.CardsString())
	case TwoPair:
		highPair := hr.HighValues[0].String()
		lowPair := hr.HighValues[1].String()
		return fmt.Sprintf("Two Pair, %ss and %ss, %s", highPair, lowPair, hr.CardsString())
	case HighCard:
		topCard := hr.HighValues[0].String()
		return fmt.Sprintf("%s-High, %s", topCard, hr.CardsString())
	default:
		return "Unknown Hand"
	}
}

// CardsString returns a string representation of the cards in the hand.
func (hr *HandResult) CardsString() string {
	if hr == nil || len(hr.Cards) == 0 {
		return "No Cards"
	}
	cards := make([]string, len(hr.Cards))
	for i, c := range hr.Cards {
		cards[i] = c.Rank.String()
		cards[i] += c.Suit.String() + " "
	}
	return util.JoinStrings(cards)
}

// handAnalysis is a helper struct to hold counts of ranks and suits.
type handAnalysis struct {
	rankCounts map[Rank]int
	suitCounts map[Suit]int
	cards      []Card // Original 8 cards, sorted by rank descending
}

// String makes handAnalysis implement the Stringer interface for debugging.
func (ha *handAnalysis) String() string {
	if ha == nil {
		return "N/A"
	}
	rankStr := "Rank Counts: "
	for rank, count := range ha.rankCounts {
		rankStr += fmt.Sprintf("%s(%d) ", rank.String(), count)
	}
	suitStr := "Suit Counts: "
	for suit, count := range ha.suitCounts {
		suitStr += fmt.Sprintf("%s (%d) ", suit.String(), count)
	}
	return fmt.Sprintf("%s, %s, Cards: %v", rankStr, suitStr, ha.cards)
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
func EvaluateHand(holeCards []Card, communityCards []Card, isLowless bool) (highResult *HandResult, lowResult *HandResult) {
	pool := make([]Card, 0, 8)
	pool = append(pool, holeCards...)
	pool = append(pool, communityCards...)

	analysis := newHandAnalysis(pool)

	// --- High Hand Evaluation ---
	if sfCards, ok := findStraightFlush(analysis); ok {
		rank := StraightFlush
		if sfCards[0].Rank == Ace {
			rank = RoyalFlush
		}
		highResult = &HandResult{Rank: rank, Cards: sfCards, HighValues: []Rank{sfCards[0].Rank}}
	} else if ssfCards, ok := findSkipStraightFlush(analysis); ok {
		highResult = &HandResult{Rank: SkipStraightFlush, Cards: ssfCards, HighValues: []Rank{ssfCards[0].Rank}}
	} else if quadRank, ok := findBestNOfAKind(analysis.rankCounts, 4); ok {
		kickers := findKickers(analysis.cards, []Rank{quadRank}, 1)
		quadCards := findCardsByRank(pool, quadRank, 4)
		highResult = &HandResult{Rank: FourOfAKind, Cards: append(quadCards, kickers...), HighValues: []Rank{quadRank, kickers[0].Rank}}
	} else if tripleRank, pairRank, ok := findBestFullHouse(analysis.rankCounts); ok {
		tripleCards := findCardsByRank(pool, tripleRank, 3)
		pairCards := findCardsByRank(pool, pairRank, 2)
		highResult = &HandResult{Rank: FullHouse, Cards: append(tripleCards, pairCards...), HighValues: []Rank{tripleRank, pairRank}}
	} else if flushCards, ok := findBestFlush(analysis); ok {
		highResult = &HandResult{
			Rank:  Flush,
			Cards: flushCards,
			HighValues: []Rank{
				flushCards[0].Rank,
				flushCards[1].Rank,
				flushCards[2].Rank,
				flushCards[3].Rank,
				flushCards[4].Rank,
			},
		}
	} else if ssCards, ok := findSkipStraight(analysis); ok {
		highResult = &HandResult{Rank: SkipStraight, Cards: ssCards, HighValues: []Rank{ssCards[0].Rank}}
	} else if straightCards, ok := findBestStraight(analysis); ok {
		highResult = &HandResult{Rank: Straight, Cards: straightCards, HighValues: []Rank{straightCards[0].Rank}}
	} else if tripleRank, ok := findBestNOfAKind(analysis.rankCounts, 3); ok {
		kickers := findKickers(analysis.cards, []Rank{tripleRank}, 2)
		tripleCards := findCardsByRank(pool, tripleRank, 3)
		highResult = &HandResult{
			Rank:       ThreeOfAKind,
			Cards:      append(tripleCards, kickers...),
			HighValues: []Rank{tripleRank, kickers[0].Rank, kickers[1].Rank},
		}
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
		highResult = &HandResult{
			Rank:       OnePair,
			Cards:      append(pairCards, kickers...),
			HighValues: []Rank{pairRank, kickers[0].Rank, kickers[1].Rank, kickers[2].Rank},
		}
	} else {
		if len(analysis.cards) < 5 {
			logrus.Warnf(
				"Not enough cards to evaluate a high hand: %d cards available - holeCards: %+v, communityCards: %+v",
				len(analysis.cards), holeCards, communityCards,
			)
			return nil, nil // Not enough cards for a valid hand
		}
		highResult = &HandResult{
			Rank:  HighCard,
			Cards: analysis.cards[:5],
			HighValues: []Rank{
				analysis.cards[0].Rank,
				analysis.cards[1].Rank,
				analysis.cards[2].Rank,
				analysis.cards[3].Rank,
				analysis.cards[4].Rank,
			},
		}
	}

	// --- Low Hand Evaluation ---
	if !isLowless {
		if lowHand, ok := findBestLowHand(analysis); ok {
			lowResult = lowHand
		}
	}

	return highResult, lowResult
}

// findSkipStraightFlush finds a Skip Straight Flush from the pool.
func findSkipStraightFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			flushAnalysis := newHandAnalysis(flushCards)
			if ssfCards, ok := findSkipStraight(flushAnalysis); ok {
				return ssfCards, true
			}
		}
	}
	return nil, false
}

// --- New Helper Function for Low Hand ---

// findBestLowHand finds the best possible 7-or-better low hand from the pool.
func findBestLowHand(analysis *handAnalysis) (*HandResult, bool) {
	uniqueLowCards := make([]Card, 0, 8)
	usedRanks := make(map[Rank]bool)

	// Find all unique cards with rank 7 or lower, treating Ace as a low card.
	for _, card := range analysis.cards {
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
		return getLowRankValue(uniqueLowCards[i].Rank) < getLowRankValue(uniqueLowCards[j].Rank)
	})

	// The best low hand consists of the 5 lowest cards
	bestLowCards := uniqueLowCards[:5]

	// Sort descending for HighValues comparison, treating Ace as the lowest rank.
	sort.Slice(bestLowCards, func(i, j int) bool {
		return getLowRankValue(bestLowCards[i].Rank) > getLowRankValue(bestLowCards[j].Rank)
	})

	return &HandResult{
		Rank:  HighCard, // Rank is not relevant for low hands in this context
		Cards: bestLowCards,
		HighValues: []Rank{
			bestLowCards[0].Rank,
			bestLowCards[1].Rank,
			bestLowCards[2].Rank,
			bestLowCards[3].Rank,
			bestLowCards[4].Rank,
		},
	}, true
}

// getLowRankValue returns the integer value of a rank for low hand comparisons, treating Ace as 1.
func getLowRankValue(r Rank) int {
	if r == Ace {
		return 1
	}
	return int(r)
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

func findSkipStraight(analysis *handAnalysis) ([]Card, bool) {
	logrus.Debugf("findSkipStraight: Analyzing handAnalysis: %+v", analysis)

	uniqueRanksAceHigh := make([]Rank, 0)
	seenRanks := make(map[Rank]bool)
	hasAce := false

	for _, card := range analysis.cards {
		if !seenRanks[card.Rank] {
			uniqueRanksAceHigh = append(uniqueRanksAceHigh, card.Rank)
			seenRanks[card.Rank] = true
			if card.Rank == Ace {
				hasAce = true
			}
		}
	}
	sort.Slice(uniqueRanksAceHigh, func(i, j int) bool { return uniqueRanksAceHigh[i] > uniqueRanksAceHigh[j] })
	listOfUniqueRanks := [][]Rank{uniqueRanksAceHigh}

	// If Ace is present, create a second list treating Ace as 1 (14 in PLS7)
	if hasAce {
		logrus.Debugf("findSkipStraight: Ace found, creating alternative rank list treating Ace as 1.")
		uniqueRanksAceLow := make([]Rank, 0)
		uniqueRanksAceLow = append(uniqueRanksAceLow, uniqueRanksAceHigh[1:]...) // Copy all except Ace
		uniqueRanksAceLow = append(uniqueRanksAceLow, uniqueRanksAceHigh[0])     // Add Ace at the end
		listOfUniqueRanks = append(listOfUniqueRanks, uniqueRanksAceLow)
	}
	logrus.Debugf("findSkipStraight: listOfUniqueRanks: %+v", listOfUniqueRanks)

	for _, uniqueRanks := range listOfUniqueRanks {
		// In Skip Straight, the highest rank must be at least 9
		if uniqueRanks[0] < 9 {
			logrus.Debugf(
				"findSkipStraight: Skipping analysis for uniqueRanks starting with %v, as it is less than 9.",
				uniqueRanks[0],
			)
			continue // Skip analysis if the highest rank is less than 9
		}
		for i := 0; i <= len(uniqueRanks)-5; i++ {
			smallest := uniqueRanks[i] - 8 // The smallest rank in a Skip Straight is 8 ranks below the top rank
			if smallest < Two {
				smallest = Ace // If the smallest rank is less than Two, treat it as Ace
				logrus.Debugf("findSkipStraight: Adjusting smallest rank to Ace as it is less than Two.")
			}
			possibleSkipStraight := []Rank{
				uniqueRanks[i],
				uniqueRanks[i] - 2,
				uniqueRanks[i] - 4,
				uniqueRanks[i] - 6,
				smallest,
			}
			logrus.Debugf("findSkipStraight: Checking possible Skip Straight: %v", possibleSkipStraight)
			isSkipStraight := true
			for _, c := range possibleSkipStraight {
				if !containsRank(uniqueRanks, c) {
					isSkipStraight = false
					logrus.Debugf(
						"findSkipStraight: Rank %v not found in uniqueRanks: %v, setting isSkipStraight to false.",
						c, uniqueRanks,
					)
					break
				}
			}
			if isSkipStraight {
				topRank := uniqueRanks[i]
				logrus.Debugf("findSkipStraight: Found Skip Straight! topRank: %v, ranks: %v", topRank, possibleSkipStraight)
				return findCardsForStraight(analysis.cards, possibleSkipStraight), true
			}
		}
	}
	logrus.Debugf("findSkipStraight: No Skip Straight found.")
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
	if uniqueRanks[0] == Ace &&
		containsRank(uniqueRanks, Five) &&
		containsRank(uniqueRanks, Four) &&
		containsRank(uniqueRanks, Three) &&
		containsRank(uniqueRanks, Two) {
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

// compareHandResults is a helper function to compare two hand results.
// It returns 1 if h1 is better, -1 if h2 is better, and 0 if they are equal.
func compareHandResults(h1, h2 *HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0
}
