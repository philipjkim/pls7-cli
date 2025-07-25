package poker

// HandRank defines the ranking of a poker hand.
type HandRank int

/* The order is important, from lowest to highest rank. */
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

// EvaluateHand analyzes a full 8-card pool and determines the best high and low hands.
// This is the main function to be implemented.
// For now, it's a placeholder.
func EvaluateHand(holeCards []Card, communityCards []Card) (highResult *HandResult, lowResult *HandResult) {
	// TODO: Implement the entire evaluation logic here.
	// This will be the core task of Step 4.

	// Placeholder return for now
	return nil, nil
}
