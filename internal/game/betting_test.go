package game

import (
	"testing"
)

// MockActionProvider provides a fixed action for testing purposes.
type MockActionProvider struct {
	Action PlayerAction
}

func (m *MockActionProvider) GetAction(g *Game) PlayerAction {
	return m.Action
}

func TestRunInteractiveBettingRound_AllInStall(t *testing.T) {
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := NewGame(playerNames, 10000, DifficultyMedium)

	// Setup: YOU can act, CPU 1 and 2 are all-in.
	g.Players[0].Status = PlayerStatusPlaying
	g.Players[1].Status = PlayerStatusAllIn
	g.Players[2].Status = PlayerStatusAllIn
	g.Phase = PhaseRiver
	g.CurrentTurnPos = 0 // It's YOU's turn.

	// The action provider will make YOU check.
	provider := &MockActionProvider{Action: PlayerAction{Type: ActionCheck}}

	// Action: Run the (currently buggy) betting round.
	g.RunInteractiveBettingRound(provider)

	// After YOU acts, turn should move to CPU 1.
	t.Errorf("Expected CurrentTurnPos to be 1 after YOU acts, but got %d", g.CurrentTurnPos)
}
