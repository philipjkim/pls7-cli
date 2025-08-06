package poker

import (
	"fmt"
	"strings"
)

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
	return fmt.Sprintf("%s%s ", c.Rank.String(), c.Suit.String())
}

// CardsFromStrings is a helper function to make creating cards in tests easier.
// It takes a space-separated string of cards like "As Kd Tc" and converts it.
func CardsFromStrings(s string) []Card {
	if s == "" {
		return []Card{}
	}
	parts := strings.Split(s, " ")

	cards := make([]Card, len(parts))
	rankMap := map[rune]Rank{
		'2': Two, '3': Three, '4': Four, '5': Five, '6': Six, '7': Seven,
		'8': Eight, '9': Nine, 'T': Ten, 'J': Jack, 'Q': Queen, 'K': King, 'A': Ace,
	}
	suitMap := map[rune]Suit{
		's': Spade, 'h': Heart, 'd': Diamond, 'c': Club,
	}
	for i, part := range parts {
		rank := rankMap[rune(part[0])]
		suit := suitMap[rune(part[1])]
		cards[i] = Card{Rank: rank, Suit: suit}
	}
	return cards
}
