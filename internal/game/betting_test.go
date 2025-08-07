package game

import (
	"math/rand"
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"testing"
)

// TestActionProvider provides a predefined sequence of actions for testing.
type TestActionProvider struct {
	Actions []PlayerAction
	index   int
}

func (p *TestActionProvider) GetAction(g *Game, player *Player, r *rand.Rand) PlayerAction {
	if p.index < len(p.Actions) {
		action := p.Actions[p.index]
		p.index++
		return action
	}
	return PlayerAction{Type: ActionFold} // Default to fold
}

func newGameForBettingTests(playerNames []string, initialChips int) *Game {
	rules := &config.GameRules{
		Abbreviation: "PLS",
		HoleCards:    config.HoleCardRules{Count: 3},
		LowHand:      config.LowHandRules{Enabled: false},
	}
	return NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false)
}

func TestBettingRound_AllInAndCall(t *testing.T) {
	util.InitLogger(true)
	playerNames := []string{"YOU", "CPU 1"}
	g := newGameForBettingTests(playerNames, 10000)
	g.StartNewHand() // YOU is SB (500), CPU 1 is BB (1000)

	// YOU (SB) raises all-in, CPU 1 (BB) calls.
	actionProvider := &TestActionProvider{
		Actions: []PlayerAction{
			{Type: ActionRaise, Amount: 10000}, // YOU raises all-in
			{Type: ActionCall},                 // CPU 1 calls
		},
	}

	g.ExecuteBettingLoop(actionProvider, func(g *Game) {})

	if g.Players[0].Status != PlayerStatusAllIn || g.Players[1].Status != PlayerStatusAllIn {
		t.Errorf("Both players should be all-in")
	}
	if g.Pot != 20000 {
		t.Errorf("Expected final pot to be 20000, but got %d", g.Pot)
	}
}

func TestBettingRound_PreFlopCheckEndsRound(t *testing.T) {
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := newGameForBettingTests(playerNames, 10000)
	g.StartNewHand() // D: YOU, SB: CPU 1, BB: CPU 2

	actionProvider := &TestActionProvider{
		Actions: []PlayerAction{
			{Type: ActionCall},  // YOU (D)
			{Type: ActionCall},  // CPU 1 (SB)
			{Type: ActionCheck}, // CPU 2 (BB)
		},
	}
	g.ExecuteBettingLoop(actionProvider, func(g *Game) {})

	if g.Pot != 3000 {
		t.Errorf("Expected pot to be 3000, but got %d", g.Pot)
	}
}
