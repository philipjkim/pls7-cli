package poker

import (
	"math/rand"
	"testing"
	"time"
)

// TestDeck_DealForDebug tests the DealForDebug method, which is used to deal specific cards for testing purposes.
func TestDeck_DealForDebug(t *testing.T) {
	wantedCards := []Card{
		{Suit: Spade, Rank: Ace},
		{Suit: Spade, Rank: King},
		{Suit: Spade, Rank: Queen},
	}

	deck := NewDeck()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	deck.Shuffle(r)
	if len(deck.Cards) < 52 {
		t.Fatalf("Expected deck to have 52 cards, but got %d", len(deck.Cards))
	}
	for _, card := range wantedCards {
		expectedDeckSize := len(deck.Cards) - 1
		dealtCard, err := deck.DealForDebug(card)
		if err != nil {
			t.Errorf("Failed to deal card %s: %v", card, err)
		}
		if dealtCard != card {
			t.Errorf("Expected dealt card to be %s, but got %s", card, dealtCard)
		}
		if len(deck.Cards) != expectedDeckSize {
			t.Errorf(
				"Expected deck size to be %d after dealing %s, but got %d",
				expectedDeckSize, card, len(deck.Cards),
			)
		}
	}
}
