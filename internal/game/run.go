package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// StartNewHand prepares the game for a new hand, including posting blinds.
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

	// Post blinds
	sbPos := (g.DealerPos + 1) % len(g.Players)
	bbPos := (g.DealerPos + 2) % len(g.Players)
	g.postBet(g.Players[sbPos], SmallBlindAmt)
	g.postBet(g.Players[bbPos], BigBlindAmt)

	g.BetToCall = BigBlindAmt
	g.CurrentTurnPos = (bbPos + 1) % len(g.Players) // Action starts after BB

	// Deal 3 cards to each player
	for i := 0; i < 3; i++ {
		for _, p := range g.Players {
			card, _ := g.Deck.Deal()
			p.Hand = append(p.Hand, card)
		}
	}
}

// postBet is a helper to handle a player's bet.
func (g *Game) postBet(player *Player, amount int) {
	if player.Chips < amount {
		amount = player.Chips // Player goes all-in
	}
	player.Chips -= amount
	player.CurrentBet += amount
	g.Pot += amount
}

// Advance moves the game to the next phase.
func (g *Game) Advance() {
	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop
		g.dealCommunityCards(3) // FIX: Deal Flop cards
	case PhaseFlop:
		g.Phase = PhaseTurn
		g.dealCommunityCards(1) // FIX: Deal Turn card
	case PhaseTurn:
		g.Phase = PhaseRiver
		g.dealCommunityCards(1) // FIX: Deal River card
	case PhaseRiver:
		g.Phase = PhaseShowdown
	case PhaseShowdown:
		g.Phase = PhaseHandOver
	}
}

// RunBettingRound executes a full betting round where everyone calls.
func (g *Game) RunBettingRound() {
	fmt.Printf("\n--- %s Betting ---\n", g.Phase.String())

	// For post-flop rounds, reset bets and start with the player after the dealer.
	if g.Phase != PhasePreFlop {
		g.BetToCall = 0
		for _, p := range g.Players {
			p.CurrentBet = 0
		}
		g.CurrentTurnPos = (g.DealerPos + 1) % len(g.Players)
	}

	numPlayers := len(g.Players)
	startPos := g.CurrentTurnPos

	for i := 0; i < numPlayers; i++ {
		playerPos := (startPos + i) % numPlayers
		player := g.Players[playerPos]
		g.CurrentTurnPos = playerPos

		if player.Status == PlayerStatusPlaying {
			amountToCall := g.BetToCall - player.CurrentBet
			if amountToCall > 0 {
				g.postBet(player, amountToCall)
			}
			// If amountToCall is 0, the player "checks".
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
