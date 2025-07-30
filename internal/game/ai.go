package game

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"time"
)

// GetCPUAction is a dispatcher that calls the appropriate AI logic based on difficulty.
func (g *Game) GetCPUAction(player *Player) PlayerAction {
	// For the actual game, we create a random source here.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch g.Difficulty {
	case DifficultyEasy:
		return g.getEasyAction(player, r)
	case DifficultyMedium:
		return g.getMediumAction(player) // Medium AI doesn't use randomness
	case DifficultyHard:
		return g.getHardAction(player, r)
	}
	return g.getMediumAction(player)
}

// getEasyAction implements the "Easy" AI: unpredictable and random.
func (g *Game) getEasyAction(player *Player, r *rand.Rand) PlayerAction {
	canCheck := player.CurrentBet == g.BetToCall

	time.Sleep(CPUThinkTime)

	if !canCheck {
		switch g.Phase {
		case PhasePreFlop:
			if r.Float64() < 0.8 {
				return PlayerAction{Type: ActionCall}
			}
			return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2}
		case PhaseFlop, PhaseTurn:
			prob := r.Float64()
			if prob < 0.3 {
				return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2}
			}
			if prob < 0.7 {
				return PlayerAction{Type: ActionCall}
			}
			return PlayerAction{Type: ActionFold}
		case PhaseRiver:
			return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2}
		default:
			panic("unhandled default case")
		}
	}
	return PlayerAction{Type: ActionCheck}
}

// getMediumAction implements the "Medium" AI: honest and rule-based.
func (g *Game) getMediumAction(player *Player) PlayerAction {
	strength := g.handEvaluator(g, player) // Use the function field
	canCheck := player.CurrentBet == g.BetToCall

	time.Sleep(CPUThinkTime)

	// Post-Flop Logic
	if g.Phase > PhasePreFlop {
		if strength >= float64(poker.FullHouse) {
			time.Sleep(CPUThinkTime)
			return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2}
		}
		if strength >= float64(poker.TwoPair) {
			if canCheck {
				time.Sleep(CPUThinkTime)
				return PlayerAction{Type: ActionBet, Amount: g.Pot / 2}
			}
			return PlayerAction{Type: ActionCall}
		}
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionFold}
	}

	// Pre-Flop Logic
	if strength > 25 {
		time.Sleep(CPUThinkTime)
		return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 3}
	}
	if strength > 15 {
		return PlayerAction{Type: ActionCall}
	}
	if !canCheck {
		return PlayerAction{Type: ActionFold}
	}
	return PlayerAction{Type: ActionCheck}
}

// getHardAction implements the "Hard" AI: strategic with bluffing.
func (g *Game) getHardAction(player *Player, r *rand.Rand) PlayerAction {
	strength := g.handEvaluator(g, player) // Use the function field
	canCheck := player.CurrentBet == g.BetToCall

	time.Sleep(CPUThinkTime)

	// 20% chance to bluff with a weak hand post-flop
	if g.Phase > PhasePreFlop && strength < float64(poker.OnePair) && r.Float64() < 0.20 {
		time.Sleep(CPUThinkTime)
		if canCheck {
			return PlayerAction{Type: ActionBet, Amount: g.Pot / 2}
		}
		return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2}
	}

	return g.getMediumAction(player)
}

// evaluateHandStrength is now a standalone function, not a method, so it can be assigned to the handEvaluator field.
func evaluateHandStrength(g *Game, player *Player) float64 {
	// Post-Flop Evaluation (based on the current best hand)
	if g.Phase > PhasePreFlop {
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards)
		return float64(highHand.Rank)
	}

	// Pre-Flop Evaluation (based on hole cards potential)
	var score float64
	hand := player.Hand

	// 1. High card points
	rankPoints := map[poker.Rank]float64{
		poker.Ace: 10, poker.King: 8, poker.Queen: 7, poker.Jack: 6, poker.Ten: 5,
	}
	for _, c := range hand {
		score += rankPoints[c.Rank]
	}

	// 2. Pair bonus
	if hand[0].Rank == hand[1].Rank || hand[0].Rank == hand[2].Rank || hand[1].Rank == hand[2].Rank {
		pairRank := hand[0].Rank
		if hand[1].Rank == hand[2].Rank {
			pairRank = hand[1].Rank
		}
		score += 15 + float64(pairRank)
	}

	// 3. Suited bonus
	if hand[0].Suit == hand[1].Suit || hand[0].Suit == hand[2].Suit || hand[1].Suit == hand[2].Suit {
		score += 4
	}

	// 4. Connector bonus
	ranks := []poker.Rank{hand[0].Rank, hand[1].Rank, hand[2].Rank}
	isConnector := (ranks[0] == ranks[1]+1) || (ranks[0] == ranks[2]+1) || (ranks[1] == ranks[2]+1)
	if isConnector {
		score += 5
	}

	return score
}
