package engine

import (
	"math/rand"
	"testing"
)

// MockActionRecorder records which players were asked to provide an action.
type MockActionRecorder struct {
	ActionsRequestedFrom map[string]int
}

func (m *MockActionRecorder) GetAction(g *Game, p *Player, r *rand.Rand) PlayerAction {
	if m.ActionsRequestedFrom == nil {
		m.ActionsRequestedFrom = make(map[string]int)
	}
	m.ActionsRequestedFrom[p.Name]++
	// For this test, the player will just check.
	return PlayerAction{Type: ActionCheck}
}

// TestHeadsUpPostFlopBettingRound reproduces the bug where a player's turn is skipped
// in a heads-up situation post-flop.
func TestHeadsUpPostFlopBettingRound(t *testing.T) {
	// Setup: 3 players, one folds pre-flop, leaving a heads-up match.
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000)
	g.StartNewHand() // Start the hand to set up blinds, etc.

	// Manually set state to be post-flop, heads-up
	g.Players[0].Status = PlayerStatusPlaying // YOU
	g.Players[1].Status = PlayerStatusPlaying // CPU1
	g.Players[2].Status = PlayerStatusFolded  // CPU2 folded

	g.Phase = PhaseFlop // Move to flop

	// This is the betting loop logic from cmd/root.go
	g.PrepareNewBettingRound()
	mockProvider := &MockActionRecorder{}

	// Loop until the betting round is considered over by the game logic.
	// We add a safeguard to prevent infinite loops in case of logic errors.
	for i := 0; i < 10; i++ { // Safeguard
		if g.IsBettingRoundOver() {
			break
		}
		player := g.CurrentPlayer()
		if player.Status != PlayerStatusPlaying {
			g.AdvanceTurn()
			continue
		}
		// Get the action, which also records that the player was asked.
		action := mockProvider.GetAction(g, player, g.Rand)
		g.ProcessAction(player, action)
		g.AdvanceTurn()
	}

	// Assertion: Both active players should have been asked for an action.
	// With the bug, only one player will be in the map.
	if len(mockProvider.ActionsRequestedFrom) != 2 {
		t.Errorf("Expected 2 players to have acted, but got %d. Players who acted: %v",
			len(mockProvider.ActionsRequestedFrom), mockProvider.ActionsRequestedFrom)
	}

	if _, ok := mockProvider.ActionsRequestedFrom["YOU"]; !ok {
		t.Error("Expected player 'YOU' to have acted, but they did not.")
	}
	if _, ok := mockProvider.ActionsRequestedFrom["CPU1"]; !ok {
		t.Error("Expected player 'CPU1' to have acted, but they did not.")
	}
}
