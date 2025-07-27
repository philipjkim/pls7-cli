package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// ProcessAction updates the game state based on a player's action.
// It returns true if an aggressive action (bet, raise) was taken.
func (g *Game) ProcessAction(player *Player, action PlayerAction) (wasAggressive bool) {
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
	case ActionBet:
		g.postBet(player, action.Amount)
		g.BetToCall = player.CurrentBet
		fmt.Printf("%s bets %d.\n", player.Name, action.Amount)
		return true // Aggressive action
	case ActionRaise:
		// In a Raise action, Amount means the total final bet size.
		amountToPost := action.Amount - player.CurrentBet
		g.postBet(player, amountToPost)
		g.BetToCall = player.CurrentBet
		fmt.Printf("%s raises to %d.\n", player.Name, action.Amount)
		return true // Aggressive action
	}
	return false
}

// CleanupHand checks for eliminated players and prepares for the next hand.
func (g *Game) CleanupHand() {
	fmt.Println("\n--- End of Hand ---")
	for _, p := range g.Players {
		if p.Chips == 0 && p.Status != PlayerStatusEliminated {
			p.Status = PlayerStatusEliminated
			fmt.Printf("%s has been eliminated!\n", p.Name)
		}
	}
}

// CountRemainingPlayers counts players who have not been eliminated from the entire game.
// This is used to check for a game-over condition (e.g., only one player is left).
func (g *Game) CountRemainingPlayers() int {
	count := 0
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			count++
		}
	}
	return count
}

// CountPlayersInHand counts players who have not folded in the current hand.
// This is used to determine if a betting round should continue or if a hand should end early.
func (g *Game) CountPlayersInHand() int {
	count := 0
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded {
			count++
		}
	}
	return count
}

// CountNonFoldedPlayers counts players who have not folded in the current hand.
// This includes players who are all-in and will see the showdown.
func (g *Game) CountNonFoldedPlayers() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying || p.Status == PlayerStatusAllIn {
			count++
		}
	}
	return count
}

// CountPlayersAbleToAct counts players who can still take betting actions.
// This excludes players who are all-in or have folded.
func (g *Game) CountPlayersAbleToAct() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			count++
		}
	}
	return count
}

// StartNewHand now skips eliminated players.
func (g *Game) StartNewHand() {
	g.HandCount++
	g.Phase = PhasePreFlop
	g.Deck = poker.NewDeck()
	g.Deck.Shuffle()
	g.CommunityCards = []poker.Card{}
	g.Pot = 0

	// Move dealer button to the next active player
	g.DealerPos = g.FindNextActivePlayer(g.DealerPos)

	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.Hand = []poker.Card{}
			p.CurrentBet = 0
			p.Status = PlayerStatusPlaying
		}
	}

	// Post blinds from active players
	sbPos := g.FindNextActivePlayer(g.DealerPos)
	bbPos := g.FindNextActivePlayer(sbPos)
	g.postBet(g.Players[sbPos], SmallBlindAmt)
	g.postBet(g.Players[bbPos], BigBlindAmt)

	g.BetToCall = BigBlindAmt
	g.CurrentTurnPos = g.FindNextActivePlayer(bbPos)

	// Deal cards to active players
	for i := 0; i < 3; i++ {
		for pos, p := range g.Players {
			if p.Status == PlayerStatusPlaying {
				card, _ := g.Deck.Deal()
				g.Players[pos].Hand = append(g.Players[pos].Hand, card)
			}
		}
	}
}

// FindNextActivePlayer finds the index of the next player who is not eliminated.
func (g *Game) FindNextActivePlayer(startPos int) int {
	pos := (startPos + 1) % len(g.Players)
	for {
		if g.Players[pos].Status != PlayerStatusEliminated {
			return pos
		}
		pos = (pos + 1) % len(g.Players)
	}
}

// postBet is a helper to handle a player's bet.
func (g *Game) postBet(player *Player, amount int) {
	if player.Chips < amount {
		amount = player.Chips
	}
	player.Chips -= amount
	player.CurrentBet += amount
	g.Pot += amount
	if player.Chips == 0 {
		player.Status = PlayerStatusAllIn
	}
}

// PrepareNewBettingRound resets player bets for the new round.
func (g *Game) PrepareNewBettingRound() {
	if g.Phase == PhasePreFlop {
		return // Pre-flop bets (blinds) are already posted.
	}
	g.BetToCall = 0
	for _, p := range g.Players {
		p.CurrentBet = 0
	}
	g.CurrentTurnPos = (g.DealerPos + 1) % len(g.Players)
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

// dealCommunityCards deals n cards to the board.
func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}
