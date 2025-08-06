package poker

import (
	"github.com/sirupsen/logrus"
)

// OutsInfo stores the detailed results of an outs calculation.
type OutsInfo struct {
	AllOuts         []Card
	OutsPerHandRank map[HandRank][]Card
}

// CalculateOuts calculates the number of cards that can improve the player's hand.
// We define "outs" as the cards that can complete the following draws:
//
// - Flush Draw: 4 cards of the same suit
// - Straight Draw: 4 cards in sequence (including gutshot)
// - Trips Draw: 2 cards of the same rank (only if 2 hole cards are of the same rank a.k.a. pocket pair)
// - Full House Draw: trips or two pair made (only if trips or two pair made by pocket pair)
// - Quad Draw: 3 cards of the same rank (only if trips made by pocket pair)
// - Skip Straight Draw: 4 cards in skip-sequence (including gutshot, e.g. 2-4-8-T, 3-5-7-9)
func CalculateOuts(holeCards []Card, communityCards []Card, lowGameEnabled bool) (bool, *OutsInfo) {
	currentHand, _ := EvaluateHand(holeCards, communityCards, lowGameEnabled)
	if currentHand == nil {
		return false, &OutsInfo{
			OutsPerHandRank: make(map[HandRank][]Card),
		}
	}

	outsInfo := &OutsInfo{
		OutsPerHandRank: make(map[HandRank][]Card),
	}
	allOutsMap := make(map[Card]bool)

	// Remove known cards from the deck
	seenCards := make(map[Card]bool)
	for _, c := range holeCards {
		seenCards[c] = true
	}
	for _, c := range communityCards {
		seenCards[c] = true
	}

	// The order of checks is from highest rank to lowest rank.
	// We check for all possible draws for hands that are better than the current hand.

	// --- Skip Straight Flush ---
	if currentHand.Rank < SkipStraightFlush {
		if hasDraw, outs := hasSkipStraightFlushDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[SkipStraightFlush] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Straight Flush ---
	if currentHand.Rank < StraightFlush {
		if hasDraw, outs := hasStraightFlushDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[StraightFlush] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Four of a Kind ---
	if currentHand.Rank < FourOfAKind {
		if hasDraw, outs := hasFourOfAKindDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[FourOfAKind] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Full House ---
	if currentHand.Rank < FullHouse {
		if hasDraw, outs := hasFullHouseDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[FullHouse] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Flush ---
	if currentHand.Rank < Flush {
		if hasDraw, outs := hasFlushDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[Flush] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Skip Straight ---
	if currentHand.Rank < SkipStraight {
		if hasDraw, outs := hasSkipStraightDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[SkipStraight] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Straight ---
	if currentHand.Rank < Straight {
		if hasDraw, outs := hasStraightDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[Straight] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Three of a Kind ---
	if currentHand.Rank < ThreeOfAKind {
		if hasDraw, outs := hasThreeOfAKindDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[ThreeOfAKind] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// --- Low Hands (only if lowGameEnabled is true) ---
	logrus.Debugf("CalculateOuts: Checking for low hands draws, lowGameEnabled: %v", lowGameEnabled)
	if lowGameEnabled {
		logrus.Debugf("CalculateOuts: Low game enabled, checking for low hand draws")
		if hasDraw, outs := hasLowHandDraw(holeCards, communityCards, seenCards); hasDraw {
			outsInfo.OutsPerHandRank[HighCard] = outs
			logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
			for _, out := range outs {
				allOutsMap[out] = true
			}
		}
	}

	// Consolidate all unique outs
	for card := range allOutsMap {
		outsInfo.AllOuts = append(outsInfo.AllOuts, card)
	}

	return len(outsInfo.AllOuts) > 0, outsInfo
}

func hasSkipStraightFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	suitCounts := make(map[Suit]int)
	for _, c := range pool {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count >= 4 { // Need at least 4 cards of the same suit for a SSF draw
			suitedCards := []Card{}
			for _, c := range pool {
				if c.Suit == suit {
					suitedCards = append(suitedCards, c)
				}
			}

			// Now check for a skip straight draw within the suited cards
			uniqueRanks := make(map[Rank]bool)
			for _, c := range suitedCards {
				uniqueRanks[c.Rank] = true
			}

			var outs []Card
			for r := Rank(2); r <= Ace; r++ {
				if !uniqueRanks[r] {
					outCard := Card{Rank: r, Suit: suit}
					if seenCards[outCard] {
						continue
					}

					empPool := append(suitedCards, outCard)
					analysis := newHandAnalysis(empPool)
					if skipStraightCards, ok := findSkipStraight(analysis); ok {
						// Check if the skip straight was completed by the card we added
						found := false
						for _, sc := range skipStraightCards {
							if sc.Rank == r {
								found = true
								break
							}
						}
						if found {
							outs = append(outs, outCard)
						}
					}
				}
			}
			if len(outs) > 0 {
				return true, outs
			}
		}
	}

	return false, nil
}

func hasStraightFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	suitCounts := make(map[Suit]int)
	for _, c := range pool {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count >= 4 { // Need at least 4 cards of the same suit for a SF draw
			suitedCards := []Card{}
			for _, c := range pool {
				if c.Suit == suit {
					suitedCards = append(suitedCards, c)
				}
			}

			// Now check for a straight draw within the suited cards
			uniqueRanks := make(map[Rank]bool)
			for _, c := range suitedCards {
				uniqueRanks[c.Rank] = true
			}

			var outs []Card
			for r := Rank(2); r <= Ace; r++ {
				if !uniqueRanks[r] {
					outCard := Card{Rank: r, Suit: suit}
					if seenCards[outCard] {
						continue
					}

					empPool := append(suitedCards, outCard)
					analysis := newHandAnalysis(empPool)
					if sfCards, ok := findBestStraight(analysis); ok {
						// Check if the straight was completed by the card we added
						found := false
						for _, sc := range sfCards {
							if sc.Rank == r {
								found = true
								break
							}
						}
						if found {
							outs = append(outs, outCard)
						}
					}
				}
			}
			if len(outs) > 0 {
				return true, outs
			}
		}
	}

	return false, nil
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

func hasSkipStraightDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	uniqueRanks := make(map[Rank]bool)
	for _, c := range pool {
		uniqueRanks[c.Rank] = true
	}
	logrus.Debugf("hasSkipStraightDraw: Unique ranks in pool: %+v", uniqueRanks)

	var outs []Card
	for r := Rank(2); r <= Ace; r++ {
		if !uniqueRanks[r] {
			// Check if adding this rank would complete a skip straight
			empPool := append(pool, Card{Rank: r, Suit: Spade}) // Suit doesn't matter for skip straight check
			analysis := newHandAnalysis(empPool)
			if skipStraightCards, ok := findSkipStraight(analysis); ok {
				// Check if the skip straight was completed by the card we added
				found := false
				for _, sc := range skipStraightCards {
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
							logrus.Debugf("hasSkipStraightDraw: Found skip straight draw out: %v, current outs: %v", outCard, outs)
						}
					}
				}
			}
		}
	}

	logrus.Debugf("hasSkipStraightDraw: Final outs for skip straight draw: %v", outs)
	return len(outs) > 0, outs
}

// hasThreeOfAKindDraw checks if the player has a trips draw (requirement: two of three hole cards of the same rank).
func hasThreeOfAKindDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	ppFound, ppRank := findPocketPair(holeCards)

	// If we don't have a pocket pair, we can't have a trips draw
	if !ppFound {
		logrus.Debugf("hasThreeOfAKindDraw: No pocket pair found in hole cards: %v", holeCards)
		return false, []Card{}
	}

	// Check if we already have trips made in the community cards
	for _, c := range communityCards {
		if c.Rank == ppRank {
			logrus.Debugf("hasThreeOfAKindDraw: Found a community card with rank %v, trips made already", c.Rank)
			return false, []Card{}
		}
	}

	pool := append(holeCards, communityCards...)
	outs := make([]Card, 0)
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		outCard := Card{Rank: ppRank, Suit: suit}
		if !seenCards[outCard] {
			outs = append(outs, outCard)
			logrus.Debugf(
				"hasThreeOfAKindDraw: Found trips draw out: %v, current outs: %v, pool: %v",
				outCard, outs, pool,
			)
		}
	}

	return len(outs) > 0, outs
}

func hasFullHouseDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	// Check if we already have a full house made
	handRank, _ := EvaluateHand(holeCards, communityCards, false)
	if handRank.Rank == FullHouse {
		logrus.Debugf("hasFullHouseDraw: Already have a full house: %v, holeCards: %v, communityCards: %v",
			handRank, holeCards, communityCards)
		return false, []Card{}
	}

	// holeCards + communityCards should one of the following HandRanks to draw a full house:
	// - Trips (3 of a kind)
	// - Two Pair (2 pairs)

	pool := append(holeCards, communityCards...)
	uniqueRankCounts := make(map[Rank]int, 0)
	for _, c := range pool {
		uniqueRankCounts[c.Rank] += 1
	}
	logrus.Debugf("hasFullHouseDraw: Unique rank counts in pool: %+v", uniqueRankCounts)

	// If handRank is Trips, we can draw a full house by adding any of non-pocket pair ranks
	if handRank.Rank == ThreeOfAKind {
		outs := make([]Card, 0)
		for rank := range uniqueRankCounts {
			if rank == handRank.HighValues[0] {
				// Skip the rank already made into trips
				continue
			}
			for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
				outCard := Card{Rank: rank, Suit: suit}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf(
						"hasFullHouseDraw: Found full house draw out: %v, current outs: %v, pool: %v",
						outCard, outs, pool,
					)
				}
			}
		}
		return len(outs) > 0, outs
	}

	// If handRank is Two Pair, we can draw a full house by adding any of pair-made ranks
	if handRank.Rank == TwoPair {
		outs := make([]Card, 0)
		for rank, count := range uniqueRankCounts {
			if count < 2 {
				// non-pair ranks cannot be used to draw a full house, it only makes three pairs (which is just two pair)
				continue
			}
			for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
				outCard := Card{Rank: rank, Suit: suit}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf(
						"hasFullHouseDraw: Found full house draw out: %v, current outs: %v, pool: %v",
						outCard, outs, pool,
					)
				}
			}
		}
		return len(outs) > 0, outs
	}

	logrus.Debugf(
		"hasFullHouseDraw: current hand rank is not suitable for full house draw: %v, "+
			"holeCards: %v, communityCards: %v, handRank: %v",
		handRank, holeCards, communityCards, handRank.Rank,
	)
	return false, []Card{}
}

// hasFourOfAKindDraw checks if the player has a four of a kind draw.
// Note that this does not require a pocket pair, but rather checks if the player has trips made.
func hasFourOfAKindDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	handRank, _ := EvaluateHand(holeCards, communityCards, false)
	if handRank.Rank != ThreeOfAKind {
		logrus.Debugf("hasFourOfAKindDraw: Current hand is not trips, cannot draw four of a kind: %v", handRank)
		return false, []Card{}
	}

	outs := make([]Card, 0)
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		outCard := Card{Rank: handRank.HighValues[0], Suit: suit}
		if !seenCards[outCard] {
			logrus.Debugf("hasFourOfAKindDraw: Found four of a kind draw out: %v", outCard)
			outs = append(outs, outCard)
		}
	}
	return len(outs) > 0, outs
}

// hasLowHandDraw checks if the player has a low hand draw.
// A low hand draw is possible if the player has at least one low card (A-7) in their hole cards
func hasLowHandDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	lowCards := make(map[Rank]bool)
	for _, c := range holeCards {
		if isLowCard(c) {
			lowCards[c.Rank] = true
		}
	}

	// At least one low card is required to consider low draws.
	if len(lowCards) < 1 {
		logrus.Debugf("hasLowHandDraw: No low cards in hole cards: %v", holeCards)
		return false, []Card{}
	}

	for _, c := range communityCards {
		if c.Rank <= Seven {
			lowCards[c.Rank] = true
		}
	}

	if len(lowCards) != 4 {
		logrus.Debugf("hasLowHandDraw: cannot draw since lowCards count is not 4: %v", lowCards)
		return false, []Card{}
	}

	// If we have 4 low cards, we can draw a low hand
	var outs []Card
	for _, r := range []Rank{Ace, Two, Three, Four, Five, Six, Seven} {
		if !lowCards[r] {
			for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
				outCard := Card{Rank: r, Suit: suit}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf("hasLowHandDraw: Found low draw out: %v, current outs: %v", outCard, outs)
				}
			}
		}
	}

	return len(outs) > 0, outs
}

// isLowCard checks if the card is a low card (A-7)
func isLowCard(c Card) bool {
	return c.Rank <= Seven || c.Rank == Ace
}

// findPocketPair checks if the player has a pocket pair in their hole cards.
//
// Returns true if a pocket pair is found, along with the rank of the pocket pair.
// If no pocket pair is found, returns false and Rank(0).
func findPocketPair(holeCards []Card) (bool, Rank) {
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

	logrus.Debugf("findPocketPair: Hole cards: %v, Pocket pair rank: %d", holeCards, pocketPairRank)
	return pocketPairRank != 0, Rank(pocketPairRank)
}

// CalculateBreakEvenEquityBasedOnPotOdds calculates the break-even equity based on the pot size and the amount to call.
func CalculateBreakEvenEquityBasedOnPotOdds(pot int, amountToCall int) float64 {
	if amountToCall == 0 {
		return 0
	}
	totalPot := pot + amountToCall
	return float64(amountToCall) / float64(totalPot)
}

// CalculateEquityWithCards calculates the equity of our hand based on the number of outs and the number of opponents.
func CalculateEquityWithCards(ourHand, communityCards []Card) float64 {
	// Note that outs are calculated without low hands draw.
	hasOuts, outsInfo := CalculateOuts(ourHand, communityCards, false)
	if !hasOuts {
		return 0
	}

	numOuts := len(outsInfo.AllOuts)

	if len(communityCards) == 3 { // Flop
		return float64(numOuts*4) / 100
	} else if len(communityCards) == 4 { // Turn
		return float64(numOuts*2) / 100
	}

	return 0
}

// CalculateEquity calculates the equity of our hand based on the game phase and the number of outs.
func CalculateEquity(numCommunityCards, numOuts int) float64 {
	if numOuts == 0 || (numCommunityCards < 3 || numCommunityCards > 5) {
		logrus.Warnf("CalculateEquity: Invalid number of outs (%d) or community cards (%d)", numOuts, numCommunityCards)
		return 0
	}

	if numCommunityCards == 3 { // Flop phase
		return float64(numOuts*4) / 100
	}

	// Turn phase
	return float64(numOuts*2) / 100
}
