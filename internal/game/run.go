package game

import (
	"fmt"
	"pls7-cli/pkg/poker"

	"github.com/sirupsen/logrus"
)

// playerHoleCardsForDebug is YOU (human player) hole cards for debugging purposes.
var playerHoleCardsForDebug = map[string]map[string]string{
	"PLS7": {
		"3As":        "As Ah Ad", // For testing outs for Four of
		"AQT-suited": "As Qs Ts", // For testing outs for Flush, Straight, and Skip Straight
		"AAK":        "As Ah Ks", // For testing outs for Three of a Kind
		"A23-suited": "As 2s 3s", // For testing outs for Straight, Flush, and low hand scenarios
	},
	"PLS": {
		"3As":        "As Ah Ad", // For testing outs for Four of
		"AQT-suited": "As Qs Ts", // For testing outs for Flush, Straight, and Skip Straight
		"AAK":        "As Ah Ks", // For testing outs for Three of a Kind
		"AKQ-suited": "As Ks Qs", // For testing outs for Straight and Flush
	},
	"NLH": {
		"AA":        "As Ah", // For testing outs for Three of a Kind and Full House
		"KK":        "Ks Kh", // For testing outs for Three of a Kind and Full House
		"AK-suited": "As Ks", // For testing outs for Straight and Flush
		"KQ-suited": "Ks Qs", // For testing outs for Straight and Flush
	},
}

// ProcessAction updates the game state based on a player's action.
// It returns true if an aggressive action (bet, raise) was taken.
func (g *Game) ProcessAction(player *Player, action PlayerAction) (wasAggressive bool, event *ActionEvent) {
	g.ActionsTakenThisRound++ // Increment for any action
	event = &ActionEvent{PlayerName: player.Name, Action: action.Type}
	switch action.Type {
	case ActionFold:
		player.Status = PlayerStatusFolded
		player.LastActionDesc = "Fold"
	case ActionCheck:
		player.LastActionDesc = "Check"
	case ActionCall:
		amountToCall := g.BetToCall - player.CurrentBet
		event.Amount = amountToCall
		g.postBet(player, amountToCall)
		desc := fmt.Sprintf("Call %d", amountToCall)
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
	case ActionBet:
		g.ActionsTakenThisRound = 1 // This player is the new aggressor, reset the count
		event.Amount = action.Amount
		g.LastRaiseAmount = action.Amount
		g.postBet(player, action.Amount)
		g.BetToCall = player.CurrentBet
		desc := fmt.Sprintf("Bet %d", player.CurrentBet) // FIX: Use actual bet amount
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		g.Aggressor = player
		return true, event
	case ActionRaise:
		g.ActionsTakenThisRound = 1 // This player is the new aggressor, reset the count
		event.Amount = action.Amount
		amountToPost := action.Amount - player.CurrentBet
		previousBetToCall := g.BetToCall
		g.postBet(player, amountToPost)
		g.BetToCall = player.CurrentBet
		g.LastRaiseAmount = g.BetToCall - previousBetToCall
		desc := fmt.Sprintf("Raise to %d", player.CurrentBet) // FIX: Use actual bet amount
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		g.Aggressor = player
		return true, event
	}
	return false, event
}

// CleanupHand checks for eliminated players and prepares for the next hand.
func (g *Game) CleanupHand() []string {
	var events []string
	events = append(events, "\n--- End of Hand ---")
	for _, p := range g.Players {
		if p.Chips == 0 && p.Status != PlayerStatusEliminated {
			p.Status = PlayerStatusEliminated
			events = append(events, fmt.Sprintf("%s has been eliminated!", p.Name))
		}
	}

	// Quit the game if only one player remains, noting who won the game.
	if g.CountRemainingPlayers() <= 1 {
		for _, p := range g.Players {
			if p.Status != PlayerStatusEliminated {
				events = append(events, fmt.Sprintf("%s wins the game!", p.Name))
				break
			}
		}
	}
	return events
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
func (g *Game) StartNewHand() (event *BlindEvent) {
	g.HandCount++

	// Update blinds based on the interval
	if g.BlindUpInterval > 0 && g.HandCount > 1 && (g.HandCount-1)%g.BlindUpInterval == 0 {
		g.SmallBlind *= 2
		g.BigBlind *= 2
		event = &BlindEvent{SmallBlind: g.SmallBlind, BigBlind: g.BigBlind}
	}

	g.Phase = PhasePreFlop
	g.Deck = poker.NewDeck()
	g.Deck.Shuffle(g.Rand)
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
	g.postBet(g.Players[sbPos], g.SmallBlind)
	g.postBet(g.Players[bbPos], g.BigBlind)

	g.BetToCall = g.BigBlind
	g.CurrentTurnPos = g.FindNextActivePlayer(bbPos)

	ruleAbbr := g.Rules.Abbreviation
	if g.DevMode || g.ShowsOuts {
		you := g.Players[0]
		if you.Status == PlayerStatusPlaying {
			// Edit the following line to set your hole cards for debugging purposes.
			switch ruleAbbr {
			case "PLS7", "PLS":
				playerHoleCards := poker.CardsFromStrings(playerHoleCardsForDebug[ruleAbbr]["3As"])
				firstCard, _ := g.Deck.DealForDebug(playerHoleCards[0])
				secondCard, _ := g.Deck.DealForDebug(playerHoleCards[1])
				thirdCard, _ := g.Deck.DealForDebug(playerHoleCards[2])
				you.Hand = []poker.Card{firstCard, secondCard, thirdCard}
			case "NLH":
				playerHoleCards := poker.CardsFromStrings(playerHoleCardsForDebug[ruleAbbr]["AA"])
				firstCard, _ := g.Deck.DealForDebug(playerHoleCards[0])
				secondCard, _ := g.Deck.DealForDebug(playerHoleCards[1])
				you.Hand = []poker.Card{firstCard, secondCard}
			default: // TODO: handle error case
				logrus.Warnf("Unsupported rule abbreviation: %s", ruleAbbr)
				playerHoleCards := poker.CardsFromStrings(playerHoleCardsForDebug[ruleAbbr]["3As"])
				firstCard, _ := g.Deck.DealForDebug(playerHoleCards[0])
				secondCard, _ := g.Deck.DealForDebug(playerHoleCards[1])
				thirdCard, _ := g.Deck.DealForDebug(playerHoleCards[2])
				you.Hand = []poker.Card{firstCard, secondCard, thirdCard}
			}
		}
		for i := 1; i < len(g.Players); i++ {
			for j := 0; j < g.Rules.HoleCards.Count; j++ {
				if g.Players[i].Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[i].Hand = append(g.Players[i].Hand, card)
				}
			}
		}
	} else {
		for i := 0; i < g.Rules.HoleCards.Count; i++ {
			for pos, p := range g.Players {
				if p.Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[pos].Hand = append(g.Players[pos].Hand, card)
				}
			}
		}
	}

	return event
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
	// If there is at least one player who can act and still needs to match the bet, betting is required.
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying && p.CurrentBet < g.BetToCall {
			return true
		}
	}
	// Otherwise, no further betting action is required for this round.
	return false
}

// PrepareNewBettingRound resets player bets and determines the starting player for a new round.
func (g *Game) PrepareNewBettingRound() {
	g.Aggressor = nil
	g.ActionsTakenThisRound = 0 // Reset the counter

	if g.Phase == PhasePreFlop {
		// Blinds are already posted, no need to reset bets.
		// Action starts after BB, and the BB is the action closer.
		bbPos := g.FindNextActivePlayer(g.FindNextActivePlayer(g.DealerPos))
		g.ActionCloserPos = bbPos
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
	g.ActionCloserPos = g.FindPreviousActivePlayer(g.CurrentTurnPos)
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

// IsBettingRoundOver checks if the betting round should end.
func (g *Game) IsBettingRoundOver() bool {
	// Condition 1: Not enough players to continue betting.
	if g.CountNonFoldedPlayers() <= 1 {
		return true
	}

	// Condition 2: All players who can act have taken at least one action.
	// Note: A player raising means others have to act again, which is handled by checking if bets are matched.
	if g.ActionsTakenThisRound < g.CountPlayersAbleToAct() {
		return false
	}

	// Condition 3: All bets are matched.
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			if p.CurrentBet < g.BetToCall {
				return false // A player still needs to call a bet.
			}
		}
	}

	// If we reach here, the round is over.
	return true
}

// CurrentPlayer returns the player whose turn it is.
func (g *Game) CurrentPlayer() *Player {
	return g.Players[g.CurrentTurnPos]
}

// AdvanceTurn moves the turn to the next active player.
func (g *Game) AdvanceTurn() {
	g.CurrentTurnPos = g.FindNextActivePlayer(g.CurrentTurnPos)
}
