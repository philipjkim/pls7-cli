package poker

// CalculateOuts calculates the number of cards that can improve the player's hand.
func CalculateOuts(holeCards []Card, communityCards []Card, lowlessMode bool) int {
	currentHand, _ := EvaluateHand(holeCards, communityCards, lowlessMode)
	if currentHand == nil {
		return 0
	}

	outs := 0
	deck := NewDeck()

	// Remove known cards from the deck
	seenCards := make(map[Card]bool)
	for _, c := range holeCards {
		seenCards[c] = true
	}
	for _, c := range communityCards {
		seenCards[c] = true
	}

	var deckCards []Card
	for _, card := range deck.Cards {
		if !seenCards[card] {
			deckCards = append(deckCards, card)
		}
	}

	for _, card := range deckCards {
		// Simulate adding the card to the community cards
		newCommunityCards := append(communityCards, card)
		improvedHand, _ := EvaluateHand(holeCards, newCommunityCards, lowlessMode)

		if improvedHand != nil && compareHandResults(improvedHand, currentHand) == 1 {
			outs++
		}
	}

	return outs
}
