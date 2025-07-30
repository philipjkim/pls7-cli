package poker

import "fmt"

// Suit represents the suit of a card.
type Suit int

const (
	Spade Suit = iota
	Heart
	Diamond
	Club
)

// String makes Suit implement the Stringer interface for easy printing.
func (s Suit) String() string {
	return []string{"♠️️", "♥️️", "♦️", "♣️️"}[s]
}

// Rank represents the rank of a card.
type Rank int

const (
	Two Rank = iota + 2 // Start from 2 to match poker rules
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace // Value is 14
)

// String makes Rank implement the Stringer interface for easy printing.
func (r Rank) String() string {
	if r >= Two && r <= Ten {
		return fmt.Sprintf("%d", r)
	}
	return map[Rank]string{
		Jack:  "J",
		Queen: "Q",
		King:  "K",
		Ace:   "A",
	}[r]
}

// Card represents a single playing card.
type Card struct {
	Suit Suit
	Rank Rank
}

// String makes Card implement the Stringer interface.
func (c Card) String() string {
	return fmt.Sprintf("[ %s%s ]", c.Rank.String(), c.Suit.String())
}
