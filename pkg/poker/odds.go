package poker

import "github.com/sirupsen/logrus"

// CalculateOuts calculates the number of cards that can improve the player's hand.
// We define "outs" as the cards that can complete the following draws:
//
// - Flush Draw: 4 cards of the same suit
// - Straight Draw: 4 cards in sequence (including gutshot)
// - Trips Draw: 2 cards of the same rank (only if 2 hole cards are of the same rank a.k.a. pocket pair)
// - Full House Draw: trips or two pair made (only if trips or two pair made by pocket pair)
// - Quad Draw: 3 cards of the same rank (only if trips made by pocket pair)
// - Skip Straight Draw: 4 cards in skip-sequence (including gutshot, e.g. 2-4-8-T, 3-5-7-9)
func CalculateOuts(holeCards []Card, communityCards []Card, lowlessMode bool) []Card {
	currentHand, _ := EvaluateHand(holeCards, communityCards, lowlessMode)
	if currentHand == nil {
		return []Card{}
	}

	outcomes := make(map[Card]bool)

	// Remove known cards from the deck
	seenCards := make(map[Card]bool)
	for _, c := range holeCards {
		seenCards[c] = true
	}
	for _, c := range communityCards {
		seenCards[c] = true
	}

	// Find outs for flush
	hasFlush, flushOuts := hasFlushDraw(holeCards, communityCards, seenCards)
	if hasFlush {
		for _, out := range flushOuts {
			outcomes[out] = true
		}
	}

	// Find outs for straight
	hasStraightDraw, straightOuts := hasStraightDraw(holeCards, communityCards, seenCards)
	if hasStraightDraw {
		for _, out := range straightOuts {
			outcomes[out] = true
		}
	}

	// Find outs for trips draw
	hasTrips, tripsOuts := hasTripsDraw(holeCards, communityCards, seenCards)
	if hasTrips {
		for _, out := range tripsOuts {
			outcomes[out] = true
		}
	}

	var result []Card
	for card := range outcomes {
		result = append(result, card)
	}

	return result
}

func hasFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	suitCounts := make(map[Suit]int)
	for _, c := range holeCards {
		suitCounts[c.Suit]++
	}
	for _, c := range communityCards {
		suitCounts[c.Suit]++
	}
	logrus.Debugf("hasFlushDraw: Suit counts: %+v", suitCounts)

	for suit, count := range suitCounts {
		if count == 4 {
			var outs []Card
			deck := NewDeck()
			for _, card := range deck.Cards {
				if card.Suit == suit && !seenCards[card] {
					outs = append(outs, card)
					logrus.Debugf("hasFlushDraw: Found flush draw out: %v, current outs: %v", card, outs)
				}
			}
			logrus.Debugf("hasFlushDraw: Final outs for flush draw: %v", outs)
			return true, outs
		}
	}

	return false, nil
}

func hasStraightDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	uniqueRanks := make(map[Rank]bool)
	for _, c := range pool {
		uniqueRanks[c.Rank] = true
	}
	logrus.Debugf("hasStraightDraw: Unique ranks in pool: %+v", uniqueRanks)

	var outs []Card
	for r := Rank(2); r <= Ace; r++ {
		if !uniqueRanks[r] {
			// Check if adding this rank would complete a straight
			empPool := append(pool, Card{Rank: r, Suit: Spade}) // Suit doesn't matter for straight check
			analysis := newHandAnalysis(empPool)
			if straightCards, ok := findBestStraight(analysis); ok {
				// Check if the straight was completed by the card we added
				found := false
				for _, sc := range straightCards {
					if sc.Rank == r {
						found = true
						break
					}
				}
				if found {
					for s := Suit(0); s <= Club; s++ {
						outCard := Card{Rank: r, Suit: s}
						if !seenCards[outCard] {
							outs = append(outs, outCard)
							logrus.Debugf("hasStraightDraw: Found straight draw out: %v, current outs: %v", outCard, outs)
						}
					}
				}
			}
		}
	}

	logrus.Debugf("hasStraightDraw: Final outs for straight draw: %v", outs)
	return len(outs) > 0, outs
}

// hasTripsDraw checks if the player has a trips draw (requirement: two of three hole cards of the same rank).
func hasTripsDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	// Check if we have a pocket pair in the hole cards
	pocketPairRank := 0
	holeRankCounts := make(map[Rank]int, 0)
	for _, c := range holeCards {
		holeRankCounts[c.Rank]++
		if holeRankCounts[c.Rank] == 2 {
			pocketPairRank = int(c.Rank)
			break
		}
	}

	// If we don't have a pocket pair, we can't have a trips draw
	if pocketPairRank == 0 {
		logrus.Debugf("hasTripsDraw: No pocket pair found in hole cards: %v", holeCards)
		return false, []Card{}
	}

	// Check if we already have trips made in the community cards
	for _, c := range communityCards {
		if c.Rank == Rank(pocketPairRank) {
			logrus.Debugf("hasTripsDraw: Found a community card with rank %v, trips made already", c.Rank)
			return false, []Card{}
		}
	}

	pool := append(holeCards, communityCards...)
	outs := make([]Card, 0)
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		outCard := Card{Rank: Rank(pocketPairRank), Suit: suit}
		if !seenCards[outCard] {
			outs = append(outs, outCard)
			logrus.Debugf(
				"hasTripsDraw: Found trips draw out: %v, current outs: %v, pool: %v",
				outCard, outs, pool,
			)
		}
	}

	return len(outs) > 0, outs
}
