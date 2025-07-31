package poker

import "github.com/sirupsen/logrus"

// CalculateOuts calculates the number of cards that can improve the player's hand.
func CalculateOuts(holeCards []Card, communityCards []Card, lowlessMode bool) int {
	currentHand, _ := EvaluateHand(holeCards, communityCards, lowlessMode)
	if currentHand == nil {
		return 0
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

	return len(outcomes)
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
