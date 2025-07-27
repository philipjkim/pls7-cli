package game

// GetCPUAction determines the action for a CPU player.
// For now, it uses a very simple "always call/check" logic.
func (g *Game) GetCPUAction(player *Player) PlayerAction {
	canCheck := player.CurrentBet == g.BetToCall

	if canCheck {
		return PlayerAction{Type: ActionCheck}
	}

	// If a player cannot check, must call (or fold, but our simple AI never folds yet).
	return PlayerAction{Type: ActionCall}
}
