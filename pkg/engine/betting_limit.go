package engine

// BettingLimitCalculator defines an interface for calculating valid bet and raise
// sizes based on a specific betting structure (e.g., Pot-Limit, No-Limit).
// This allows the game engine to handle different poker variants by plugging in
// the appropriate calculator.
type BettingLimitCalculator interface {
	// CalculateBettingLimits determines the minimum and maximum valid raise amounts
	// for the current player. The returned values represent the total amount the
	// player's bet would be after the raise, not just the raise increment.
	//
	// For example, if the current bet is 50 and the minimum raise increment is 50,
	// the minRaiseTotal returned would be 100.
	CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int)
}

// PotLimitCalculator implements the BettingLimitCalculator for Pot-Limit games.
type PotLimitCalculator struct{}

// CalculateBettingLimits calculates the valid raise range for a Pot-Limit game.
// In Pot-Limit, the maximum raise amount is the size of the total pot after the
// player has made their call.
func (c *PotLimitCalculator) CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int) {
	player := g.Players[g.CurrentTurnPos]
	amountToCall := g.BetToCall - player.CurrentBet

	minRaiseTotal = g.minRaiseAmount()

	// Pot-Limit Raise calculation: The maximum raise is the size of the pot
	// after the player has notionally made their call.
	// The implementation calculates this as: current pot + amount to call.
	// Note: A more standard calculation would also include all other bets on the table.
	potAfterCall := g.Pot + amountToCall
	maxRaiseAmount := potAfterCall
	maxRaiseTotal = g.BetToCall + maxRaiseAmount

	// A player cannot bet more chips than they have.
	if maxRaiseTotal > player.Chips+player.CurrentBet {
		maxRaiseTotal = player.Chips + player.CurrentBet
	}

	// If a player's all-in is less than a legal minimum raise, they can still go all-in.
	if minRaiseTotal > player.Chips+player.CurrentBet {
		minRaiseTotal = player.Chips + player.CurrentBet
		maxRaiseTotal = player.Chips + player.CurrentBet // The max raise is also the all-in amount.
	}

	// If the max raise is less than the min raise (due to a short stack all-in),
	// then the valid raise amount is clamped to the max raise.
	if maxRaiseTotal < minRaiseTotal {
		minRaiseTotal = maxRaiseTotal
	}

	return minRaiseTotal, maxRaiseTotal
}

// NoLimitCalculator implements the BettingLimitCalculator for No-Limit games.
type NoLimitCalculator struct{}

// CalculateBettingLimits calculates the valid raise range for a No-Limit game.
// In No-Limit, the maximum raise is simply the player's entire chip stack (all-in).
func (c *NoLimitCalculator) CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int) {
	player := g.Players[g.CurrentTurnPos]

	minRaiseTotal = g.minRaiseAmount()

	// The maximum raise in No-Limit is the player's entire stack.
	maxRaiseTotal = player.Chips + player.CurrentBet

	// If a player's all-in is less than a legal minimum raise, they can still go all-in.
	if minRaiseTotal > player.Chips+player.CurrentBet {
		minRaiseTotal = player.Chips + player.CurrentBet
	}

	return minRaiseTotal, maxRaiseTotal
}
