package poker

import (
	"fmt"
	"math/rand"
	"time"
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
func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
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
