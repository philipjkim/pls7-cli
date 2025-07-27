package game

// CalculateBettingLimits returns the minimum and maximum allowed raise amounts.
// Note: The amounts returned are the TOTAL size of the new bet.
func (g *Game) CalculateBettingLimits() (minRaiseTotal int, maxRaiseTotal int) {
	player := g.Players[g.CurrentTurnPos]
	amountToCall := g.BetToCall - player.CurrentBet

	// Minimum Raise calculation:
	// A raise must be at least as large as the previous bet or raise.
	// For simplicity, we'll define a min raise as doubling the current bet to call.
	minRaiseTotal = g.BetToCall * 2
	if g.BetToCall == 0 { // If no bet, min bet is the Big Blind
		minRaiseTotal = BigBlindAmt
	}

	// Pot-Limit Raise calculation:
	// The max raise is the size of the pot after the player has called.
	// Pot size after calling = current pot + all bets on the table + the call amount.
	// Our g.Pot already includes all bets, so:
	potAfterCall := g.Pot + amountToCall
	maxRaiseAmount := potAfterCall
	maxRaiseTotal = g.BetToCall + maxRaiseAmount

	// A player cannot bet more than they have.
	if maxRaiseTotal > player.Chips+player.CurrentBet {
		maxRaiseTotal = player.Chips + player.CurrentBet
	}

	// A player must have enough chips to make a minimum raise.
	if minRaiseTotal > player.Chips+player.CurrentBet {
		minRaiseTotal = player.Chips + player.CurrentBet
	}

	return minRaiseTotal, maxRaiseTotal
}
