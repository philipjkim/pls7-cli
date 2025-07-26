package game

// CalculateBettingLimits returns the minimum and maximum allowed bet/raise amounts.
func (g *Game) CalculateBettingLimits() (min int, max int) {
	// Minimum bet/raise is typically the size of the big blind, or the previous bet size.
	// For simplicity, we'll use BigBlindAmt for now.
	min = BigBlindAmt

	// Pot-Limit calculation:
	// Max raise = Pot + (2 * last bet)
	// The amount a player can put in total = Pot + (2 * last bet) + their own previous call
	// So the final bet amount = Pot + last_bet + call_amount
	// A simpler way: Max bet is the size of the pot.
	// Let's use a simplified version for now: max bet is the size of the pot.
	max = g.Pot + g.BetToCall

	// A player cannot bet more than they have.
	player := g.Players[g.CurrentTurnPos]
	if max > player.Chips {
		max = player.Chips
	}

	return min, max
}
