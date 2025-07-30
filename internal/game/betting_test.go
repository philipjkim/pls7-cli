package game

import (
	"fmt"
	"testing"
)

// SimpleActionProvider provides a fixed action for testing purposes.
type SimpleActionProvider struct {
	Action PlayerAction
}

func (m *SimpleActionProvider) GetAction(g *Game, p *Player) PlayerAction {
	switch m.Action.Type {
	case ActionFold, ActionBet, ActionRaise:
		return m.Action
	case ActionCheck, ActionCall:
		canCheck := p.CurrentBet == g.BetToCall
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionCall}
	}
	return m.Action
}

func newGameForBettingTests(playerNames []string, initialChips int) *Game {
	return NewGame(playerNames, initialChips, DifficultyMedium, true, false)
}

// TestBettingRound_PlayerMustCallAllIn tests a realistic multi-street all-in scenario.
func TestBettingRound_PlayerMustCallAllIn(t *testing.T) {
	// Scenario: 3 players, 3000 chips each. After pre-flop, pot is 3000.
	// On the flop, CPU 1 folds, CPU 2 bets all-in for 2000. Action is on YOU.
	// YOU must call the 2000 all-in. The round should then terminate.
	playerNames := []string{"YOU", "CPU 1", "CPU 2"}
	g := newGameForBettingTests(playerNames, 3000)

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
	playerAP := &SimpleActionProvider{Action: PlayerAction{Type: ActionCall}}
	cpuAP := &SimpleActionProvider{Action: PlayerAction{Type: ActionCheck}} // CPU players won't act further.
	g.ExecuteBettingLoop(playerAP, cpuAP, displayMiniGameState)

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
	g := newGameForBettingTests(playerNames, 10000)

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
	playerAP := &SimpleActionProvider{Action: PlayerAction{Type: ActionCheck}} // This should not be called.
	cpuAP := &SimpleActionProvider{Action: PlayerAction{Type: ActionCheck}}    // CPU players won't act further.
	g.ExecuteBettingLoop(playerAP, cpuAP, displayMiniGameState)

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

// TestBettingRound_PreFlopNoRaiseEndsCorrectly tests the specific scenario where the action
// folds/calls around to the Big Blind, who then checks, which should end the round.
func TestBettingRound_PreFlopNoRaiseEndsCorrectly(t *testing.T) {
	// Scenario: 6 players (4 active, 2 eliminated)
	// Players in g.Players: [YOU, CPU 1, CPU 2, CPU 3, CPU 4, CPU 5]
	// CPU 4 and CPU 5 were eliminated.

	// D: CPU 3, SB: YOU, BB: CPU 1
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	g := newGameForBettingTests(playerNames, 10000)

	// --- Setup Pre-Flop state ---
	g.DealerPos = 2  // CPU 2 was the dealer, and the dealer should be changed to CPU 3 after StartNewHand
	g.StartNewHand() // This will post blinds and deal cards
	g.ExecuteBettingLoop(
		&SimpleActionProvider{Action: PlayerAction{Type: ActionFold}},  // YOU will check/call
		&SimpleActionProvider{Action: PlayerAction{Type: ActionCheck}}, // CPUs will check/call
		displayMiniGameState,
	)

	// 1. CPU 2 calls.
	// 2. CPU 3 (Dealer) calls.
	// 3. YOU (SB) fold.
	// 4. CPU 1 (BB) checks.
	// The betting round should end here.
	fmt.Printf("%+v\n", g)
	if g.Pot != 3500 { // 500 (SB) + 1000 (BB) + 1000 (CPU 2) + 1000 (CPU 3)
		t.Errorf("Expected pot to be 3500, but got %d", g.Pot)
	}
}

type AggressorAndCallerActionProvider struct{}

func (a *AggressorAndCallerActionProvider) GetAction(g *Game, p *Player) PlayerAction {
	canCheck := p.CurrentBet == g.BetToCall
	if p.Name == "CPU 3" { // Aggressor
		if canCheck {
			return PlayerAction{Type: ActionBet, Amount: 1000} // Aggressor bets
		}
		return PlayerAction{Type: ActionRaise, Amount: 2000} // Aggressor raises
	} else {
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionCall}
	}
}

// TestBettingRound_RoundEndsCorrectlyWhenLeftOfAggressorHasEliminated tests the scenario where
// the player to the left of the aggressor has been eliminated, which should end the betting
// round correctly without infinite loops or errors.
func TestBettingRound_RoundEndsCorrectlyWhenLeftOfAggressorHasEliminated(t *testing.T) {
	// Scenario: 4 players (3 active, 1 eliminated)
	// Players in g.Players: [YOU, CPU 1, CPU 2, CPU 3]
	// CPU 2 was eliminated.

	// D: CPU 3, SB: YOU, BB: CPU 1
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3"}
	g := newGameForBettingTests(playerNames, 10000)
	g.Players[2].Status = PlayerStatusEliminated // CPU 2 is eliminated
	g.Players[2].Chips = 0                       // CPU 2 has no chips

	// CPU 1 should only check/call, but CPU 3 should bet/raise.

	// --- Setup Pre-Flop state ---
	g.DealerPos = 2  // CPU 2 was the dealer, and the dealer should be changed to CPU 3 after StartNewHand
	g.StartNewHand() // This will post blinds and deal cards
	g.ExecuteBettingLoop(
		&SimpleActionProvider{Action: PlayerAction{Type: ActionFold}}, // YOU will check/call
		&AggressorAndCallerActionProvider{},                           // CPU 1 will check/call, CPU 3 will bet/raise
		displayMiniGameState,
	)

	// 1. CPU 3 raises to 2000.
	// 3. YOU (SB) fold.
	// 4. CPU 1 (BB) calls.
	// The betting round should end here.
	fmt.Printf("%+v\n", g)
	if g.Pot != 4500 { // 500 (YOU, SB) + 2000 (CPU 1, BB) + 2000 (CPU 3, D)
		t.Errorf("Expected pot to be 4500, but got %d", g.Pot)
	}
}

// TestBettingRound_HandlesAllFoldWithOneEliminatedPlayer tests the scenario where all players fold except one,
// and one player is eliminated. The betting round should end correctly without errors.
func TestBettingRound_HandlesAllFoldWithOneEliminatedPlayer(t *testing.T) {
	// Scenario: 6 players (5 active, 1 eliminated)
	// Players in g.Players: [YOU, CPU 1, CPU 2, CPU 3, CPU 4, CPU 5]
	// CPU 5 was eliminated.

	// D: CPU 3, SB: CPU 4, BB: YOU
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
	g := newGameForBettingTests(playerNames, 10000)
	g.Players[5].Status = PlayerStatusEliminated // CPU 5 is eliminated
	g.Players[5].Chips = 0                       // CPU 5 has no chips

	// --- Setup Pre-Flop state ---
	g.DealerPos = 2  // CPU 2 was the dealer, and the dealer should be changed to CPU 3 after StartNewHand
	g.StartNewHand() // This will post blinds and deal cards

	// All players fold except YOU.
	g.ExecuteBettingLoop(
		&SimpleActionProvider{Action: PlayerAction{Type: ActionCheck}}, // YOU will check
		&SimpleActionProvider{Action: PlayerAction{Type: ActionFold}},  // All active CPUs will fold
		displayMiniGameState,
	)

	// The betting round should end here with pot being 1500 (500 from YOU, 1000 from CPU 4) without infinite loops.
	fmt.Printf("%+v\n", g)
	if g.Pot != 1500 {
		t.Errorf("Expected pot to be 1500, but got %d", g.Pot)
	}
}

func displayMiniGameState(g *Game) {
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			fmt.Printf("%s's turn: Chips: %d, Current Bet: %d, Action: %v\n", p.Name, p.Chips, p.CurrentBet, p.LastActionDesc)
		}
	}
}

// TestActionCloserPosForPreFlop tests the action closer position for pre-flop betting rounds.
func TestActionCloserPosForPreFlop_WorksCorrectlyWithEliminatedPlayers(t *testing.T) {
	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
	g := newGameForBettingTests(playerNames, 10000)
	g.Players[5].Status = PlayerStatusEliminated // CPU 5 is eliminated
	g.Players[5].Chips = 0                       // CPU 5 has no chips
	g.DealerPos = 3
	expected := 0 // YOU is the action closer in pre-flop with CPU 5 eliminated
	actual := actionCloserPosForPreFlop(g)
	if expected != actual {
		t.Errorf("Expected action closer position to be %d, but got %d", expected, actual)
	}
}

// TestCalculateBettingLimits tests the pot-limit calculations in various scenarios.
func TestCalculateBettingLimits(t *testing.T) {
	testCases := []struct {
		name             string
		pot              int
		betToCall        int
		playerChips      int
		playerCurrentBet int
		lastRaiseAmount  int
		expectedMinRaise int
		expectedMaxRaise int
	}{
		{
			name:             "Pre-flop, first to act after blinds",
			pot:              1500, // SB(500) + BB(1000)
			betToCall:        1000, // Must call BB
			playerChips:      10000,
			playerCurrentBet: 0,
			lastRaiseAmount:  1000, // BB is the last raise
			expectedMinRaise: 2000, // Raise to 2 * BB
			expectedMaxRaise: 3500, // Pot(1500) + Call(1000) = 2500. Raise by 2500 to a total of 3500
		},
		{
			name:             "Post-flop, first to act",
			pot:              3000,
			betToCall:        0, // No bet yet
			playerChips:      10000,
			playerCurrentBet: 0,
			lastRaiseAmount:  0,
			expectedMinRaise: 1000, // Min bet is BB
			expectedMaxRaise: 3000, // Bet the pot
		},
		{
			name:             "Post-flop, facing a bet",
			pot:              4000, // Pot was 3000, someone bet 1000
			betToCall:        1000,
			playerChips:      8000,
			playerCurrentBet: 0,
			lastRaiseAmount:  1000,
			expectedMinRaise: 2000, // Raise to 2 * Bet
			expectedMaxRaise: 6000, // Pot(4000) + Call(1000) = 5000. Raise by 5000 to a total of 6000
		},
		{
			name:             "Player is all-in, max raise is their stack",
			pot:              4000,
			betToCall:        1000,
			playerChips:      3000, // Not enough to make the max pot raise
			playerCurrentBet: 0,
			lastRaiseAmount:  1000,
			expectedMinRaise: 2000,
			expectedMaxRaise: 3000, // Limited by stack
		},
		{
			name:             "Player doesn't have enough for min raise",
			pot:              4000,
			betToCall:        1000,
			playerChips:      1500, // Not enough to make the min raise of 2000
			playerCurrentBet: 0,
			lastRaiseAmount:  1000,
			expectedMinRaise: 1500, // Limited by stack
			expectedMaxRaise: 1500,
		},
		{
			name:             "Re-raising scenario",
			pot:              10000, // Complex pot
			betToCall:        3000,  // The Original bet was 1000, someone raised to 3000
			playerChips:      20000,
			playerCurrentBet: 1000,  // Player already called the initial 1000
			lastRaiseAmount:  2000,  // The Last raise was 2000 (3000-1000)
			expectedMinRaise: 5000,  // Min raise is the size of the last raise (2000), so 3000 + 2000 = 5000
			expectedMaxRaise: 15000, // Pot(10000) + Call(2000) = 12000. Raise by 12,000 to a total of 15,000
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup game state for the test
			g := &Game{
				Pot:             tc.pot,
				BetToCall:       tc.betToCall,
				LastRaiseAmount: tc.lastRaiseAmount,
				Players: []*Player{
					{Name: "YOU", Chips: tc.playerChips, CurrentBet: tc.playerCurrentBet},
				},
				CurrentTurnPos: 0,
			}

			minRaise, maxRaise := g.CalculateBettingLimits()

			if minRaise != tc.expectedMinRaise {
				t.Errorf("Expected min raise to be %d, but got %d", tc.expectedMinRaise, minRaise)
			}
			if maxRaise != tc.expectedMaxRaise {
				t.Errorf("Expected max raise to be %d, but got %d", tc.expectedMaxRaise, maxRaise)
			}
		})
	}
}

// TestBettingLoop_AllInScenarios tests various all-in scenarios to ensure the betting loop and pot are handled correctly.
func TestBettingLoop_AllInScenarios(t *testing.T) {
	// Sub-test 1: One player goes all-in and is called by another.
	t.Run("Single All-In and Call", func(t *testing.T) {
		playerNames := []string{"YOU", "CPU 1"}
		g := newGameForBettingTests(playerNames, 10000)
		g.StartNewHand() // YOU is SB, CPU 1 is BB

		// YOU goes all-in
		playerActionProvider := &SimpleActionProvider{Action: PlayerAction{Type: ActionRaise, Amount: 10000}}
		// CPU 1 calls
		cpuActionProvider := &SimpleActionProvider{Action: PlayerAction{Type: ActionCall}}

		g.ExecuteBettingLoop(playerActionProvider, cpuActionProvider, displayMiniGameState)

		if g.Players[0].Status != PlayerStatusAllIn {
			t.Errorf("Expected YOU to be all-in, but status is %v", g.Players[0].Status)
		}
		if g.Players[1].Status != PlayerStatusAllIn {
			t.Errorf("Expected CPU 1 to be all-in, but status is %v", g.Players[1].Status)
		}
		if g.Pot != 20000 {
			t.Errorf("Expected final pot to be 20000, but got %d", g.Pot)
		}
	})

	// Sub-test 2: Multiple all-ins creating a main pot and a side pot.
	t.Run("Multiple All-Ins with Side Pot", func(t *testing.T) {
		playerNames := []string{"ShortStack", "MidStack", "BigStack"}
		g := newGameForBettingTests(playerNames, 0)
		g.Players[0].Chips = 2000  // ShortStack
		g.Players[1].Chips = 5000  // MidStack
		g.Players[2].Chips = 10000 // BigStack
		g.StartNewHand()           // ShortStack is SB, MidStack is BB

		// Action: BigStack raises to 10,000 (all-in), ShortStack calls, MidStack calls.
		actionProviders := map[string]ActionProvider{
			"BigStack":   &SimpleActionProvider{Action: PlayerAction{Type: ActionRaise, Amount: 10000}},
			"ShortStack": &SimpleActionProvider{Action: PlayerAction{Type: ActionCall}},
			"MidStack":   &SimpleActionProvider{Action: PlayerAction{Type: ActionCall}},
		}
		provider := &MultiActionProvider{Providers: actionProviders}

		g.ExecuteBettingLoop(provider, provider, displayMiniGameState)

		// This is a simplified check. A full side-pot implementation is needed in pot.go
		if g.Pot != 17000 { // 2000*3 (main) + 3000*2 (side) = 12,000 is wrong. Correct is 2000+5000+10,000
			t.Errorf("Expected final pot to be 17000, but got %d", g.Pot)
		}
	})
}

// MultiActionProvider is a helper for tests with multiple players taking different actions.
type MultiActionProvider struct {
	Providers map[string]ActionProvider
}

func (m *MultiActionProvider) GetAction(g *Game, p *Player) PlayerAction {
	if provider, ok := m.Providers[p.Name]; ok {
		return provider.GetAction(g, p)
	}
	// Default action if no specific provider is found for the player
	return PlayerAction{Type: ActionFold}
}
