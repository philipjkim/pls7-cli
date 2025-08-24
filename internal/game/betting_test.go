package game

import (
	"math/rand"
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"testing"
	"time"
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
		BettingLimit: "pot_limit",
	}
	return NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false, 0)
}

// newGameForBettingTestsWithRules creates a game with a specific rule abbreviation.
func newGameForBettingTestsWithRules(playerNames []string, initialChips int, ruleAbbr string) *Game {
	rules := &config.GameRules{
		Abbreviation: ruleAbbr,
	}
	switch ruleAbbr {
	case "NLH":
		rules.HoleCards = config.HoleCardRules{Count: 2}
		rules.LowHand = config.LowHandRules{Enabled: false}
		rules.BettingLimit = "no_limit"
	case "PLS7":
		rules.HoleCards = config.HoleCardRules{Count: 3}
		rules.LowHand = config.LowHandRules{Enabled: ruleAbbr == "PLS7", MaxRank: 7}
		rules.BettingLimit = "pot_limit"
	default: // PLS
		rules.HoleCards = config.HoleCardRules{Count: 3}
		rules.LowHand = config.LowHandRules{Enabled: false}
		rules.BettingLimit = "pot_limit"
	}
	return NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false, 0)
}

// all players have matched the bet, isBettingActionRequired should return false.
func TestIsBettingActionRequired_MatchedBets_False(t *testing.T) {
	g := newGameForBettingTestsWithRules([]string{"YOU", "CPU 1", "CPU 2"}, 10000, "NLH")
	g.StartNewHand()
	// Force a state where all active players have matched the bet
	g.BetToCall = BigBlindAmt
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.Status = PlayerStatusPlaying
			p.CurrentBet = BigBlindAmt
		}
	}
	if g.isBettingActionRequired() {
		t.Fatalf("expected no further betting required when all bets are matched")
	}
}

// when a player still needs to call, isBettingActionRequired should return true.
func TestIsBettingActionRequired_PlayerNeedsToCall_True(t *testing.T) {
	g := newGameForBettingTestsWithRules([]string{"YOU", "CPU 1", "CPU 2"}, 10000, "NLH")
	g.StartNewHand()
	g.BetToCall = BigBlindAmt
	// YOU still needs to call
	g.Players[0].Status = PlayerStatusPlaying
	g.Players[0].CurrentBet = SmallBlindAmt
	// Others have matched
	g.Players[1].Status = PlayerStatusPlaying
	g.Players[1].CurrentBet = BigBlindAmt
	g.Players[2].Status = PlayerStatusPlaying
	g.Players[2].CurrentBet = BigBlindAmt

	if !g.isBettingActionRequired() {
		t.Fatalf("expected betting to be required when a player must still call")
	}
}

// in pre-flop, when all players call and check, the betting loop should terminate.
func TestExecuteBettingLoop_PreFlop_AllCallCheck_Terminates_NLH(t *testing.T) {
	util.InitLogger(true)
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := newGameForBettingTestsWithRules(playerNames, 10000, "NLH")
	g.StartNewHand() // D: YOU, SB: CPU 1, BB: CPU 2, action starts at YOU

	actionProvider := &TestActionProvider{
		Actions: []PlayerAction{
			{Type: ActionCall},  // YOU
			{Type: ActionCall},  // CPU 1 (SB)
			{Type: ActionCheck}, // CPU 2 (BB)
		},
	}

	finished := make(chan struct{})
	go func() {
		g.ExecuteBettingLoop(actionProvider, func(g *Game) {})
		close(finished)
	}()

	select {
	case <-finished:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatalf("ExecuteBettingLoop did not terminate in expected time (potential infinite loop)")
	}

	if g.Pot != 3000 { // 1500 blinds + 1000 (YOU call) + 500 (SB completes)
		t.Fatalf("expected pot to be 3000 after all-call-check preflop, got %d", g.Pot)
	}
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

func TestBettingRound_MultiRaiseAndAllIn(t *testing.T) {
	util.InitLogger(true)
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	g := newGameForBettingTests(playerNames, 10000)
	g.Players[1].Chips = 3000 // CPU 1 is short-stacked
	g.StartNewHand()          // D: YOU, SB: CPU 1, BB: CPU 2, UTG: CPU 3

	actionProvider := &TestActionProvider{
		Actions: []PlayerAction{
			{Type: ActionRaise, Amount: 3000}, // CPU 3 (UTG) raises to 3000
			{Type: ActionFold},                // YOU (D) folds
			{Type: ActionCall},                // CPU 1 (SB) calls all-in
			{Type: ActionFold},                // CPU 2 (BB) folds
		},
	}

	g.ExecuteBettingLoop(actionProvider, func(g *Game) {})

	if g.Pot != 7000 {
		t.Errorf("Expected pot to be 7000, but got %d", g.Pot)
	}
	if g.Players[1].Status != PlayerStatusAllIn {
		t.Errorf("CPU 1 should be all-in")
	}
	if g.Players[0].Status != PlayerStatusFolded {
		t.Errorf("YOU should have folded")
	}
	if g.Players[2].Status != PlayerStatusFolded {
		t.Errorf("CPU 2 should have folded")
	}
}
