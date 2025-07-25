package game

import (
	"fmt"
	"pls7-cli/internal/cli"
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

	sbPos := (g.DealerPos + 1) % len(g.Players)
	bbPos := (g.DealerPos + 2) % len(g.Players)
	g.postBet(g.Players[sbPos], SmallBlindAmt)
	g.postBet(g.Players[bbPos], BigBlindAmt)

	g.BetToCall = BigBlindAmt
	g.CurrentTurnPos = (bbPos + 1) % len(g.Players)

	for i := 0; i < 3; i++ {
		for _, p := range g.Players {
			card, _ := g.Deck.Deal()
			p.Hand = append(p.Hand, card)
		}
	}
}

// ProcessAction updates the game state based on a player's action.
func (g *Game) ProcessAction(player *Player, action PlayerAction) {
	switch action.Type {
	case ActionFold:
		player.Status = PlayerStatusFolded
		fmt.Printf("%s folds.\n", player.Name)
	case ActionCheck:
		fmt.Printf("%s checks.\n", player.Name)
	case ActionCall:
		amountToCall := g.BetToCall - player.CurrentBet
		g.postBet(player, amountToCall)
		fmt.Printf("%s calls %d.\n", player.Name, amountToCall)
		// TODO: Bet and Raise logic will be added here.
	}
}

// RunBettingRound now handles the interactive loop.
func (g *Game) RunBettingRound() {
	// This logic will become more complex to handle re-raises.
	// For now, it's a single pass.
	if g.Phase != PhasePreFlop {
		g.BetToCall = 0
		for _, p := range g.Players {
			p.CurrentBet = 0
		}
		g.CurrentTurnPos = (g.DealerPos + 1) % len(g.Players)
	}

	numPlayers := len(g.Players)
	playersInRound := g.CountActivePlayers()

	for i := 0; i < playersInRound; i++ {
		// This loop needs to be smarter to handle betting correctly.
		// We will fix this in the next step.
		player := g.Players[g.CurrentTurnPos]

		if player.Status == PlayerStatusPlaying {
			cli.DisplayGameState(g) // Display state before each action

			var action PlayerAction
			if player.IsCPU {
				// Mock CPU action: always check or call
				if player.CurrentBet < g.BetToCall {
					action = PlayerAction{Type: ActionCall}
				} else {
					action = PlayerAction{Type: ActionCheck}
				}
			} else {
				// Get action from human player
				action = cli.PromptForAction(g)
			}
			g.ProcessAction(player, action)
		}
		g.CurrentTurnPos = (g.CurrentTurnPos + 1) % numPlayers
	}
}

// CountActivePlayers returns the number of players still in the hand.
func (g *Game) CountActivePlayers() int {
	count := 0
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded {
			count++
		}
	}
	return count
}

// (Other functions like postBet, Advance, dealCommunityCards remain the same)
func (g *Game) postBet(player *Player, amount int) {
	if player.Chips < amount {
		amount = player.Chips
	}
	player.Chips -= amount
	player.CurrentBet += amount
	g.Pot += amount
}

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

func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}
