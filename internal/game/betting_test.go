package game

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"testing"
)

// TestActionProvider provides a predefined sequence of actions for testing.
type TestActionProvider struct {
	Actions []PlayerAction
	index   int
}

func (p *TestActionProvider) GetAction(_ *Game, _ *Player, _ *rand.Rand) PlayerAction {
	if p.index < len(p.Actions) {
		action := p.Actions[p.index]
		p.index++
		return action
	}
	return PlayerAction{Type: ActionFold} // Default to fold
}

func newGameForBettingTests(playerNames []string, initialChips int) *Game {
	rules := &poker.GameRules{
		Abbreviation: "PLS",
		HoleCards:    poker.HoleCardRules{Count: 3},
		LowHand:      poker.LowHandRules{Enabled: false},
		BettingLimit: "pot_limit",
	}
	return NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false, 0)
}

// newGameForBettingTestsWithRules creates a game with a specific rule abbreviation.
func newGameForBettingTestsWithRules(playerNames []string, initialChips int, ruleAbbr string) *Game {
	rules := &poker.GameRules{
		Abbreviation: ruleAbbr,
	}
	switch ruleAbbr {
	case "NLH":
		rules.HoleCards = poker.HoleCardRules{Count: 2}
		rules.LowHand = poker.LowHandRules{Enabled: false}
		rules.BettingLimit = "no_limit"
	case "PLS7":
		rules.HoleCards = poker.HoleCardRules{Count: 3}
		rules.LowHand = poker.LowHandRules{Enabled: ruleAbbr == "PLS7", MaxRank: 7}
		rules.BettingLimit = "pot_limit"
	default: // PLS
		rules.HoleCards = poker.HoleCardRules{Count: 3}
		rules.LowHand = poker.LowHandRules{Enabled: false}
		rules.BettingLimit = "pot_limit"
	}
	return NewGame(playerNames, initialChips, DifficultyMedium, rules, true, false, 0)
}

// all players have matched the bet, isBettingActionRequired should return false.
func TestIsBettingActionRequired_MatchedBets_False(t *testing.T) {
	g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1", "CPU2"}, 10000, "NLH")
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
	g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1", "CPU2"}, 10000, "NLH")
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

func TestIsBettingRoundOver(t *testing.T) {
	t.Run("Round not over - bets not matched", func(t *testing.T) {
		g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1"}, 10000, "NLH")
		g.Players[0].CurrentBet = 100
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		if g.IsBettingRoundOver() {
			t.Error("Expected betting round to NOT be over")
		}
	})

	t.Run("Round over - all bets matched", func(t *testing.T) {
		g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1"}, 10000, "NLH")
		g.Players[0].CurrentBet = 200
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over")
		}
	})

	t.Run("Round over - one player left", func(t *testing.T) {
		g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1"}, 10000, "NLH")
		g.Players[0].Status = PlayerStatusPlaying
		g.Players[1].Status = PlayerStatusFolded
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over when only one player remains")
		}
	})

	t.Run("Round over - all-in player cannot act on a raise", func(t *testing.T) {
		g := newGameForBettingTestsWithRules([]string{"YOU", "CPU1"}, 10000, "NLH")
		g.Players[0].Status = PlayerStatusAllIn
		g.Players[0].CurrentBet = 100
		g.Players[1].Status = PlayerStatusPlaying
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over when a player is all-in and cannot call a raise")
		}
	})
}
