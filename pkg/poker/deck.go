package poker

import (
	"fmt"
	"math/rand"
)

// Deck represents a deck of cards.
type Deck struct {
	Cards []Card
}

// NewDeck creates a standard 52-card deck.
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for suit := Spade; suit <= Club; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return &Deck{Cards: cards}
}

// Shuffle randomizes the order of cards in the deck.
func (d *Deck) Shuffle(r *rand.Rand) {
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Deal removes and returns the top card from the deck.
// Returns an error if the deck is empty.
func (d *Deck) Deal() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, fmt.Errorf("deck is empty")
	}
	card := d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]
	return card, nil
}

// DealForDebug removes and returns the specific card from the deck.
// This is used for testing purposes to ensure specific cards can be dealt.
func (d *Deck) DealForDebug(card Card) (Card, error) {
	for i, c := range d.Cards {
		if c == card {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			return c, nil
		}
	}
	return Card{}, fmt.Errorf("card %s not found in deck", card)
}
