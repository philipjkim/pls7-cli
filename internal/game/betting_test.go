package game

import (
	"fmt"
	"testing"
)

// MockActionProvider provides a fixed action for testing purposes.
type MockActionProvider struct {
	Action PlayerAction
}

func (m *MockActionProvider) GetAction(g *Game) PlayerAction {
	return m.Action
}

// TestBettingRound_PlayerMustCallAllIn tests a realistic multi-street all-in scenario.
func TestBettingRound_PlayerMustCallAllIn(t *testing.T) {
	// Scenario: 3 players, 3000 chips each. After pre-flop, pot is 3000.
	// On the flop, CPU 1 folds, CPU 2 bets all-in for 2000. Action is on YOU.
	// YOU must call the 2000 all-in. The round should then terminate.
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := NewGame(playerNames, 3000, DifficultyMedium)

	// --- Manually set the game state to be on the Flop, after pre-flop betting ---
	g.Phase = PhaseFlop
	g.Pot = 3000 // 1000 from each player pre-flop
	g.Players[0].Chips = 2000
	g.Players[1].Chips = 2000
	g.Players[2].Chips = 2000
	g.DealerPos = 0
	g.CurrentTurnPos = 1 // Action starts with CPU 1 after the dealer

	// --- Manually process flop actions leading up to YOU's turn ---
	// 1. CPU 1 Folds.
	g.ProcessAction(g.Players[1], PlayerAction{Type: ActionFold})
	g.CurrentTurnPos = 2

	// 2. CPU 2 bets all-in for their remaining 2000 chips.
	g.ProcessAction(g.Players[2], PlayerAction{Type: ActionBet, Amount: 2000})
	g.CurrentTurnPos = 0 // Action is now on YOU

	// --- This is the action we are testing ---
	// 3. YOU must now call the 2000 all-in.
	providerYOU := &MockActionProvider{Action: PlayerAction{Type: ActionCall}}
	g.ExecuteBettingLoop(providerYOU, displayMiniGameState)

	// --- Assertions ---
	// The betting loop should have terminated correctly.
	if g.Players[0].Chips != 0 {
		t.Errorf("Expected YOU to have 0 chips after calling all-in, but got %d", g.Players[0].Chips)
	}
	if g.Players[0].Status != PlayerStatusAllIn {
		t.Errorf("Expected YOU's status to be AllIn, but got %v", g.Players[0].Status)
	}
	if g.Pot != 7000 { // 3000 (pre-flop) + 2000 (CPU 2) + 2000 (YOU)
		t.Errorf("Expected final pot to be 7000, but got %d", g.Pot)
	}
}

// TestBettingRound_SkipsWhenNoFurtherActionPossible tests the scenario where a player's bet
// is called by opponents who go all-in with fewer chips.
func TestBettingRound_SkipsWhenNoFurtherActionPossible(t *testing.T) {
	// Scenario: 3 players. On the flop, YOU bet 5000.
	// CPU 1 calls all-in for 1000. CPU 2 calls all-in for 2000.
	// Since both opponents are all-in and have fewer chips, there is no more action.
	// The betting round should terminate immediately without asking YOU to act again.
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := NewGame(playerNames, 10000, DifficultyMedium)

	// --- Manually set the game state to be on the Flop ---
	g.Phase = PhaseFlop
	g.DealerPos = 2
	g.CurrentTurnPos = 0 // Action starts with YOU

	// --- Manually process flop actions ---
	// 1. YOU bet 5000.
	g.ProcessAction(g.Players[0], PlayerAction{Type: ActionBet, Amount: 5000})
	g.CurrentTurnPos = 1

	// 2. CPU 1 calls but only has 1000 chips, so goes all-in.
	g.Players[1].Chips = 1000
	g.ProcessAction(g.Players[1], PlayerAction{Type: ActionCall})
	g.CurrentTurnPos = 2

	// 3. CPU 2 calls but only has 2000 chips, so goes all-in.
	g.Players[2].Chips = 2000
	g.ProcessAction(g.Players[2], PlayerAction{Type: ActionCall})
	g.CurrentTurnPos = 0 // Action would return to YOU

	// --- This is the action we are testing ---
	// The betting loop should recognize that no further action is possible and return immediately.
	providerYOU := &MockActionProvider{Action: PlayerAction{Type: ActionCheck}} // This should not be called.
	g.ExecuteBettingLoop(providerYOU, displayMiniGameState)

	// --- Assertions ---
	// The main assertion is that the test completes without timing out.
	// We also check the final state.
	if g.Players[0].Chips != 5000 { // 10000 - 5000
		t.Errorf("Expected YOU to have 5000 chips, but got %d", g.Players[0].Chips)
	}
	if g.Pot != 8000 { // 5000 (YOU) + 1000 (CPU 1) + 2000 (CPU 2)
		t.Errorf("Expected final pot to be 8000, but got %d", g.Pot)
	}
}

func displayMiniGameState(g *Game) {
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			fmt.Printf("%s's turn: Chips: %d, Current Bet: %d, Action: %v\n", p.Name, p.Chips, p.CurrentBet, p.LastActionDesc)
		}
	}
}
