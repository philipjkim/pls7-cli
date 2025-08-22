package game

// BettingLimitCalculator defines the interface for different betting limit strategies.
type BettingLimitCalculator interface {
	// CalculateBettingLimits returns the minimum and maximum allowed bet/raise amounts
	// based on the current game state. The amounts returned are the TOTAL size of the new bet.
	CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int)
}

// PotLimitCalculator implements the BettingLimitCalculator for pot-limit games.
type PotLimitCalculator struct{}

// CalculateBettingLimits for pot-limit games.
func (c *PotLimitCalculator) CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int) {
	player := g.Players[g.CurrentTurnPos]
	amountToCall := g.BetToCall - player.CurrentBet

	minRaiseTotal = calculateMinRaiseTotal(g, player)

	// Pot-Limit Raise calculation:
	// The max raise is the size of the pot after the player has called.
	// Pot size after calling = current pot + all bets on the table + the call amount.
	potAfterCall := g.Pot + amountToCall
	maxRaiseAmount := potAfterCall
	maxRaiseTotal = g.BetToCall + maxRaiseAmount

	// A player cannot bet more than they have.
	if maxRaiseTotal > player.Chips+player.CurrentBet {
		maxRaiseTotal = player.Chips + player.CurrentBet
	}

	// A player must have enough chips to make a minimum raise.
	// If they don't, their max raise is their entire stack.
	if minRaiseTotal > player.Chips+player.CurrentBet {
		minRaiseTotal = player.Chips + player.CurrentBet
		maxRaiseTotal = player.Chips + player.CurrentBet // Can't raise more than all-in
	}

	// Edge case: If a player's all-in is less than a min-raise, other players can only call or fold.
	// If the max raise is less than the min raise (due to stack size), it means the player is going all-in for less than a full raise.
	if maxRaiseTotal < minRaiseTotal {
		minRaiseTotal = maxRaiseTotal
	}

	return minRaiseTotal, maxRaiseTotal
}

// NoLimitCalculator implements the BettingLimitCalculator for no-limit games.
type NoLimitCalculator struct{}

// CalculateBettingLimits for no-limit games.
func (c *NoLimitCalculator) CalculateBettingLimits(g *Game) (minRaiseTotal int, maxRaiseTotal int) {
	player := g.Players[g.CurrentTurnPos]

	minRaiseTotal = calculateMinRaiseTotal(g, player)

	// Maximum Raise calculation (player's entire stack)
	maxRaiseTotal = player.Chips + player.CurrentBet

	// A player must have enough chips to make a minimum raise.
	// If they don't, their max raise is their entire stack.
	if minRaiseTotal > player.Chips+player.CurrentBet {
		minRaiseTotal = player.Chips + player.CurrentBet
	}

	return minRaiseTotal, maxRaiseTotal
}

// calculateMinRaiseTotal calculates the minimum raise total based on the game's state.
func calculateMinRaiseTotal(g *Game, player *Player) int {
	// Minimum Raise calculation:
	minRaiseIncrease := g.LastRaiseAmount
	if minRaiseIncrease == 0 { // If no previous raise, min raise is the size of the bet to call
		minRaiseIncrease = g.BetToCall
	}
	if g.BetToCall == 0 { // If no bet, min bet is the Big Blind
		minRaiseIncrease = BigBlindAmt
	}
	return g.BetToCall + minRaiseIncrease
}
