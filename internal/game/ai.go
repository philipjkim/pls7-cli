package game

import "pls7-cli/pkg/poker"

// GetCPUAction determines the action for a CPU player based on "Medium" difficulty logic.
func (g *Game) GetCPUAction(player *Player) PlayerAction {
	strength := g.evaluateHandStrength(player)
	canCheck := player.CurrentBet == g.BetToCall

	// Post-Flop Logic
	if g.Phase > PhasePreFlop {
		if strength >= float64(poker.FullHouse) { // Very strong hand
			return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 2} // Simple 2x raise for now
		}
		if strength >= float64(poker.TwoPair) { // Decent hand
			if canCheck {
				return PlayerAction{Type: ActionBet, Amount: g.Pot / 2} // Bet half pot
			}
			return PlayerAction{Type: ActionCall}
		}
		// Weak hand
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionFold}
	}

	// Pre-Flop Logic
	if strength > 25 { // Premium hands
		return PlayerAction{Type: ActionRaise, Amount: g.BetToCall * 3} // 3x raise
	}
	if strength > 15 { // Playable hands
		return PlayerAction{Type: ActionCall}
	}

	// Fold weak pre-flop hands if there's a bet to call
	if !canCheck {
		return PlayerAction{Type: ActionFold}
	}
	// Otherwise, check for free
	return PlayerAction{Type: ActionCheck}
}

// evaluateHandStrength calculates a numerical score for a player's hand.
func (g *Game) evaluateHandStrength(player *Player) float64 {
	// Post-Flop Evaluation (based on current best hand)
	if g.Phase > PhasePreFlop {
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards)
		// We can directly use the HandRank as a score.
		return float64(highHand.Rank)
	}

	// Pre-Flop Evaluation (based on hole cards potential)
	var score float64
	hand := player.Hand

	// 1. High card points
	rankPoints := map[poker.Rank]float64{
		poker.Ace:   10,
		poker.King:  8,
		poker.Queen: 7,
		poker.Jack:  6,
		poker.Ten:   5,
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
		score += 15 + float64(pairRank) // Higher pairs are better
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
