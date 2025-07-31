package game

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
)

// ProcessAction updates the game state based on a player's action.
// It returns true if an aggressive action (bet, raise) was taken.
func (g *Game) ProcessAction(player *Player, action PlayerAction) (wasAggressive bool) {
	switch action.Type {
	case ActionFold:
		player.Status = PlayerStatusFolded
		player.LastActionDesc = "Fold"
		fmt.Printf("%s folds.\n", player.Name)
	case ActionCheck:
		player.LastActionDesc = "Check"
		fmt.Printf("%s checks.\n", player.Name)
	case ActionCall:
		amountToCall := g.BetToCall - player.CurrentBet
		g.postBet(player, amountToCall)
		desc := fmt.Sprintf("Call %s", util.FormatNumber(amountToCall))
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		fmt.Printf("%s calls %s.\n", player.Name, util.FormatNumber(amountToCall))
	case ActionBet:
		g.LastRaiseAmount = action.Amount
		g.postBet(player, action.Amount)
		g.BetToCall = player.CurrentBet
		desc := fmt.Sprintf("Bet %s", util.FormatNumber(player.CurrentBet)) // FIX: Use actual bet amount
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		fmt.Printf("%s bets %s.\n", player.Name, util.FormatNumber(player.CurrentBet)) // FIX: Use actual bet amount
		return true
	case ActionRaise:
		amountToPost := action.Amount - player.CurrentBet
		g.LastRaiseAmount = amountToPost
		g.postBet(player, amountToPost)
		g.BetToCall = player.CurrentBet
		desc := fmt.Sprintf("Raise to %s", util.FormatNumber(player.CurrentBet)) // FIX: Use actual bet amount
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		fmt.Printf("%s raises to %s.\n", player.Name, util.FormatNumber(player.CurrentBet)) // FIX: Use actual bet amount
		return true
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

	// Quit the game if only one player remains, noting who won the game.
	if g.CountRemainingPlayers() <= 1 {
		for _, p := range g.Players {
			if p.Status != PlayerStatusEliminated {
				fmt.Printf("%s wins the game!\n", p.Name)
				break
			}
		}
		return
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

// StartNewHand now resets the LastActionDesc field.
func (g *Game) StartNewHand() {
	g.HandCount++
	g.Phase = PhasePreFlop
	g.Deck = poker.NewDeck()
	g.Deck.Shuffle()
	g.CommunityCards = []poker.Card{}
	g.Pot = 0
	g.LastRaiseAmount = 0

	g.DealerPos = g.FindNextActivePlayer(g.DealerPos)

	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.Hand = []poker.Card{}
			p.CurrentBet = 0
			p.TotalBetInHand = 0 // Reset total bet in hand
			p.Status = PlayerStatusPlaying
			p.LastActionDesc = "" // Reset action description
		}
	}

	sbPos := g.FindNextActivePlayer(g.DealerPos)
	bbPos := g.FindNextActivePlayer(sbPos)
	g.postBet(g.Players[sbPos], SmallBlindAmt)
	g.postBet(g.Players[bbPos], BigBlindAmt)

	g.BetToCall = BigBlindAmt
	g.CurrentTurnPos = g.FindNextActivePlayer(bbPos)

	if g.DevMode {
		// Deal [As, Qs, Ts] to the first player in dev mode to test Skip Straight with high Ace.
		you := g.Players[0]
		if you.Status == PlayerStatusPlaying {
			firstCard, _ := g.Deck.DealForDebug(poker.Card{Rank: poker.Ace, Suit: poker.Spade})
			secondCard, _ := g.Deck.DealForDebug(poker.Card{Rank: poker.Queen, Suit: poker.Spade})
			thirdCard, _ := g.Deck.DealForDebug(poker.Card{Rank: poker.Ten, Suit: poker.Spade})
			you.Hand = []poker.Card{firstCard, secondCard, thirdCard}
		}
		for i := 1; i < len(g.Players); i++ {
			for j := 0; j < 3; j++ {
				if g.Players[i].Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[i].Hand = append(g.Players[i].Hand, card)
				}
			}
		}
	} else {
		for i := 0; i < 3; i++ {
			for pos, p := range g.Players {
				if p.Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[pos].Hand = append(g.Players[pos].Hand, card)
				}
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
	player.TotalBetInHand += amount // Update total bet in hand
	g.Pot += amount
	if player.Chips == 0 {
		player.Status = PlayerStatusAllIn
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

// dealCommunityCards deals n cards to the board.
func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}

// isBettingActionRequired checks if there is any pending bet that needs to be called.
// The round can be skipped if all non-folded players have the same amount bet.
func (g *Game) isBettingActionRequired() bool {
	// If less than two players can even act (have chips and haven't folded), no betting can occur.
	if g.CountPlayersAbleToAct() < 2 {
		// However, we must check if the single active player needs to call a previous all-in.
		for _, p := range g.Players {
			if p.Status == PlayerStatusPlaying && p.CurrentBet < g.BetToCall {
				return true // This player must act.
			}
		}
		return false
	}
	return true
}

// PrepareNewBettingRound resets player bets and determines the starting player for a new round.
func (g *Game) PrepareNewBettingRound() {
	if g.Phase == PhasePreFlop {
		// Blinds are already posted, no need to reset bets.
		return
	}
	// For post-flop rounds, reset bets and start with the player after the dealer.
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.CurrentBet = 0
			p.LastActionDesc = ""
		}
	}
	g.BetToCall = 0
	g.LastRaiseAmount = 0
	g.CurrentTurnPos = g.FindNextActivePlayer(g.DealerPos)
}

// FindPreviousActivePlayer finds the index of the previous player who is not eliminated.
func (g *Game) FindPreviousActivePlayer(startPos int) int {
	// TODO: Handle case where all players are eliminated.
	pos := (startPos - 1 + len(g.Players)) % len(g.Players)
	for {
		if g.Players[pos].Status != PlayerStatusEliminated {
			return pos
		}
		pos = (pos - 1 + len(g.Players)) % len(g.Players)
	}
}

// ExecuteBettingLoop runs the core betting logic for a round.
// It assumes the round has already been prepared.
func (g *Game) ExecuteBettingLoop(
	playerActionProvider ActionProvider,
	cpuActionProvider ActionProvider,
	displayCurrentStatus func(g *Game),
) {
	// If only one player remains in the hand, award them the pot immediately.
	if g.CountNonFoldedPlayers() == 1 {
		// Find the last remaining player
		var lastPlayer *Player
		for _, p := range g.Players {
			if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
				lastPlayer = p
				break
			}
		}
		if lastPlayer != nil {
			logrus.Debugf("Only one player (%s) remains in the hand. Awarding pot.", lastPlayer.Name)
			// Award the pot to this player. This will also reset g.Pot to 0.
			g.AwardPotToLastPlayer()
		}
		return // End the betting loop
	}

	if g.CountPlayersAbleToAct() < 2 {
		// If only one player can act, check if they need to call a previous all-in.
		player := g.Players[g.CurrentTurnPos]
		if player.Status == PlayerStatusPlaying && player.CurrentBet < g.BetToCall {
			// This single player must act.
		} else {
			return // Otherwise, no betting is possible, so skip the round.
		}
	}

	actionCloserPos := 0
	if g.Phase == PhasePreFlop {
		actionCloserPos = actionCloserPosForPreFlop(g) // BB acts last
	} else {
		actionCloserPos = g.DealerPos // Dealer acts last
	}

	for {
		player := g.Players[g.CurrentTurnPos]

		if player.Status == PlayerStatusPlaying {
			displayCurrentStatus(g) // Display the current game state

			var action PlayerAction
			if player.IsCPU {
				action = cpuActionProvider.GetAction(g, player)
			} else {
				action = playerActionProvider.GetAction(g, player)
			}

			wasAggressive := g.ProcessAction(player, action)
			logrus.Debugf(
				"%s's action: %v, wasAggressive: %v, currentActionCloser: %v\n",
				player.Name, action, wasAggressive, g.Players[actionCloserPos].Name,
			)
			if wasAggressive {
				previousActionCloserPos := actionCloserPos
				actionCloserPos = g.FindPreviousActivePlayer(g.CurrentTurnPos)
				logrus.Debugf(
					"action closer changed from %v to %v\n",
					g.Players[previousActionCloserPos].Name,
					g.Players[actionCloserPos].Name,
				)
			}
		}

		if g.CurrentTurnPos == actionCloserPos {
			break
		}

		g.CurrentTurnPos = g.FindNextActivePlayer(g.CurrentTurnPos)
	}
}

// actionCloserPosForPreFlop returns the position of the action closer in Pre-Flop phase.
func actionCloserPosForPreFlop(g *Game) int {
	// In Pre-Flop, the action closer is the Big Blind.
	ac := (g.DealerPos + 2) % len(g.Players)
	for {
		if g.Players[ac].Status != PlayerStatusEliminated {
			return ac
		}
		ac = (ac + 1) % len(g.Players) // Skip eliminated players
	}
}
