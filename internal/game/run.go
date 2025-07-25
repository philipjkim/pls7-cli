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
func (g *Game) Advance() {
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
	case PhaseShowdown:
		g.Phase = PhaseHandOver
	}
}

// RunBettingRound executes a full betting round.
func (g *Game) RunBettingRound() {
	// For now, we'll just mock a simple bet-call scenario
	// In a real game, this would be a complex loop.
	// Let's simulate one player betting and others calling.

	if g.Phase == PhasePreFlop {
		// Mock Blinds for now
		g.Pot += 15
		fmt.Println("--- Pre-Flop Betting (mock) ---")
		g.BetToCall = 10 // Mock BB
	} else {
		fmt.Printf("--- %s Betting (mock) ---\n", g.Phase.String())
		g.BetToCall = 0 // No bet yet
	}

	// Reset player bets for the new round
	for _, p := range g.Players {
		p.CurrentBet = 0
	}

	// This is a simplified loop. A real one is more complex.
	// For now, we assume everyone just calls or checks.
	numPlayers := len(g.Players)
	startPos := (g.DealerPos + 1) % numPlayers
	if g.Phase == PhasePreFlop {
		startPos = (g.DealerPos + 3) % numPlayers // Action starts after BB
	}

	for i := 0; i < numPlayers; i++ {
		playerPos := (startPos + i) % numPlayers
		player := g.Players[playerPos]
		g.CurrentTurnPos = playerPos

		if player.Status == PlayerStatusPlaying {
			// Mock action: everyone just checks or calls
			if g.BetToCall > 0 {
				// Call
				callAmount := g.BetToCall - player.CurrentBet
				if player.Chips >= callAmount {
					player.Chips -= callAmount
					player.CurrentBet += callAmount
					g.Pot += callAmount
				}
			}
			// If BetToCall is 0, they "check".
		}
	}
}

// dealCommunityCards deals n cards to the board.
func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}
