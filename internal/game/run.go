package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// StartNewHand prepares the game for a new hand and sets it to PreFlop.
func (g *Game) StartNewHand() {
	g.Phase = PhasePreFlop
	g.Deck = poker.NewDeck()
	g.Deck.Shuffle()
	g.CommunityCards = []poker.Card{}
	g.Pot = 0
	g.DealerPos = (g.DealerPos + 1) % len(g.Players)

	for _, p := range g.Players {
		p.Hand = []poker.Card{}
		p.CurrentBet = 0
		p.Status = PlayerStatusPlaying
	}

	// Deal 3 cards to each player
	for i := 0; i < 3; i++ {
		for _, p := range g.Players {
			card, _ := g.Deck.Deal()
			p.Hand = append(p.Hand, card)
		}
	}
}

// Advance moves the game to the next phase.
// It returns true if the hand is over.
func (g *Game) Advance() bool {
	g.runBettingRound() // Mock betting round

	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop
		g.dealCommunityCards(3)
	case PhaseFlop:
		g.Phase = PhaseTurn
		g.dealCommunityCards(1)
	case PhaseTurn:
		g.Phase = PhaseRiver
		g.dealCommunityCards(1)
	case PhaseRiver:
		g.Phase = PhaseShowdown
		g.showdown()
	case PhaseShowdown:
		g.Phase = PhaseHandOver
		return true // Hand is now over
	}
	return false
}

// dealCommunityCards deals n cards to the board.
func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}

// runBettingRound is a placeholder for now.
func (g *Game) runBettingRound() {
	fmt.Println("--- Betting Round (mock) ---")
}

// showdown handles the end of the hand.
func (g *Game) showdown() {
	fmt.Println("\n--- SHOWDOWN ---")
	// This will be expanded to show results.
}
