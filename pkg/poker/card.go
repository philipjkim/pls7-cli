// Package poker provides the core data structures and rules for playing poker.
// It includes types for cards, decks, hand evaluation, and game rules, forming
// the foundational building blocks for a poker game engine.
package poker

import (
	"fmt"
	"strings"
)

// Suit represents the suit of a playing card (Spade, Heart, Diamond, Club).
type Suit int

// Suit constants define the four suits in a standard deck of cards.
const (
	Spade   Suit = iota // Spade represents the spade suit (♠️).
	Heart               // Heart represents the heart suit (♥️).
	Diamond             // Diamond represents the diamond suit (♦️).
	Club                // Club represents the club suit (♣️).
)

// String returns the emoji representation of the suit. It implements the fmt.Stringer
// interface, allowing for easy printing.
func (s Suit) String() string {
	return []string{"♠️️", "♥️️", "♦️", "♣️️"}[s]
}

// Rank represents the rank of a playing card, from Two (2) to Ace (14).
type Rank int

// Rank constants define the thirteen ranks in a standard deck.
// The values are assigned starting from 2 to align with their poker value.
const (
	Two   Rank = iota + 2 // Two represents the rank 2.
	Three                 // Three represents the rank 3.
	Four                  // Four represents the rank 4.
	Five                  // Five represents the rank 5.
	Six                   // Six represents the rank 6.
	Seven                 // Seven represents the rank 7.
	Eight                 // Eight represents the rank 8.
	Nine                  // Nine represents the rank 9.
	Ten                   // Ten represents the rank 10.
	Jack                  // Jack represents the rank 11.
	Queen                 // Queen represents the rank 12.
	King                  // King represents the rank 13.
	Ace                   // Ace represents the rank 14, the highest rank.
)

// String returns the string representation of the rank (e.g., "A", "K", "10").
// It implements the fmt.Stringer interface for easy printing.
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

// Card represents a single playing card with a specific Suit and Rank.
type Card struct {
	Suit Suit // The suit of the card (e.g., Spade, Heart).
	Rank Rank // The rank of the card (e.g., Ace, King).
}

// String returns the string representation of a card, combining its rank and suit
// (e.g., "As ", "Kd "). It implements the fmt.Stringer interface.
func (c Card) String() string {
	return fmt.Sprintf("%s%s ", c.Rank.String(), c.Suit.String())
}

// CardsFromStrings is a utility function for creating a slice of cards from a
// space-separated string. It is primarily used for testing and setting up
// specific game scenarios.
//
// The string format for each card is a two-character string:
// The first character represents the rank ('A', 'K', 'Q', 'J', 'T', '9'-'2').
// The second character represents the suit ('s', 'h', 'd', 'c').
// Example: "As Kd Tc" creates a slice with the Ace of Spades, King of Diamonds,
// and Ten of Clubs.
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
