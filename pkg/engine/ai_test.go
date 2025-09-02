package engine

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"testing"
)

func TestEvaluateHandStrength(t *testing.T) {
	testCases := []struct {
		name              string
		phase             GamePhase
		holeCardsStr      string
		communityCardsStr string
		expectedScore     float64
	}{
		{name: "Pre-Flop - High Pair (Aces)", phase: PhasePreFlop, holeCardsStr: "As Ac 2d", expectedScore: 49},
		{name: "Pre-Flop - Low Pair", phase: PhasePreFlop, holeCardsStr: "2s 2c Ad", expectedScore: 27},
		{name: "Pre-Flop - Suited Connectors", phase: PhasePreFlop, holeCardsStr: "8s 7s 2d", expectedScore: 4},
		{name: "Pre-Flop - Premium Suited High Cards", phase: PhasePreFlop, holeCardsStr: "As Ks Qs", expectedScore: 33},
		{name: "Post-Flop - Full House", phase: PhaseTurn, holeCardsStr: "As Ac Qd", communityCardsStr: "Ah Kc Kh 3d 4s", expectedScore: float64(poker.FullHouse)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:          tc.phase,
				CommunityCards: poker.CardsFromStrings(tc.communityCardsStr),
				Rules:          &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			player := &Player{Hand: poker.CardsFromStrings(tc.holeCardsStr)}
			score := evaluateHandStrength(g, player)
			if score != tc.expectedScore {
				t.Errorf("Expected score %.2f, but got %.2f", tc.expectedScore, score)
			}
		})
	}
}

func TestCPUActionProfileBased(t *testing.T) {
	lagProfile := aiProfiles["Loose-Aggressive"]
	tpProfile := aiProfiles["Tight-Passive"]

	testCases := []struct {
		name           string
		seed           int64
		profile        *AIProfile
		phase          GamePhase
		handStrength   float64
		canCheck       bool
		expectedAction ActionType
	}{
		{name: "LAG AI - Bluffs with weak hand", seed: 2, profile: &lagProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionBet},
		{name: "LAG AI - No Bluff on high random", seed: 12345, profile: &lagProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionCheck},
		{name: "TP AI - No Bluff even with low random", seed: 1, profile: &tpProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionCheck},
		{name: "Pre-Flop - Folds below threshold", seed: 1, profile: &tpProfile, phase: PhasePreFlop, handStrength: 21, canCheck: false, expectedAction: ActionFold},
		{name: "Pre-Flop - Raises above threshold", seed: 1, profile: &tpProfile, phase: PhasePreFlop, handStrength: 29, canCheck: false, expectedAction: ActionRaise},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:     tc.phase,
				Pot:       100,
				BetToCall: 0,
				Rules:     &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			if !tc.canCheck {
				g.BetToCall = 10
			}
			player := &Player{Profile: tc.profile}

			g.handEvaluator = func(g *Game, p *Player) float64 { return tc.handStrength }

			r := rand.New(rand.NewSource(tc.seed))
			action := g.GetCPUAction(player, r)

			if action.Type != tc.expectedAction {
				t.Errorf("Expected action %v, but got %v", tc.expectedAction, action.Type)
			}
		})
	}
}
