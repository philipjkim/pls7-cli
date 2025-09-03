package poker

// HandIterator defines the interface for strategies that generate all possible
// 5-card hand combinations from a player's hole cards and the community cards,
// based on a specific set of rules.
// This allows for different poker variants (e.g., NLH vs. PLO) to use different
// rules for forming a hand.
type HandIterator interface {
	// Generate creates and returns a slice of all valid 5-card hands ([][]Card).
	Generate(holeCards, communityCards []Card, rules *GameRules) [][]Card
}

// AnyCombinationGenerator is a strategy that generates 5-card hands by taking
// the best 5 cards from the combined pool of hole and community cards. It implements
// the "any" UseConstraint, typical for games like No-Limit Hold'em.
type AnyCombinationGenerator struct{}

func (g *AnyCombinationGenerator) Generate(holeCards, communityCards []Card, rules *GameRules) [][]Card {
	pool := make([]Card, 0, len(holeCards)+len(communityCards))
	pool = append(pool, holeCards...)
	pool = append(pool, communityCards...)
	return combinations(pool, 5)
}

// ExactCombinationGenerator is a strategy that generates 5-card hands by taking
// a specific number of cards from the hole and the rest from the community.
// It implements the "exact" UseConstraint, which is the rule for Omaha.
type ExactCombinationGenerator struct{}

func (g *ExactCombinationGenerator) Generate(holeCards, communityCards []Card, rules *GameRules) [][]Card {
	numHoleCardsToUse := rules.HoleCards.UseCount
	numBoardCardsToUse := 5 - numHoleCardsToUse

	if len(holeCards) < numHoleCardsToUse || len(communityCards) < numBoardCardsToUse {
		return nil // Not enough cards to form a valid hand
	}

	holeCombos := combinations(holeCards, numHoleCardsToUse)
	boardCombos := combinations(communityCards, numBoardCardsToUse)

	if holeCombos == nil || boardCombos == nil {
		return nil
	}

	var all5CardCombos [][]Card
	for _, hc := range holeCombos {
		for _, bc := range boardCombos {
			// It's crucial to create a new slice for each hand to avoid slice memory sharing issues.
			// Appending directly to a shared slice that is also being appended to in a loop can lead to unexpected data overwrites.
			// By initializing a new slice `currentHand` for each combination, we ensure that each appended hand is a distinct entity.
			currentHand := make([]Card, 0, 5)
			currentHand = append(currentHand, hc...)
			currentHand = append(currentHand, bc...)
			all5CardCombos = append(all5CardCombos, currentHand)
		}
	}
	return all5CardCombos
}
