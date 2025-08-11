package game

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"sort"
	"time"
)

// byRank implements sort.Interface for []poker.Rank
type byRank []poker.Rank

func (a byRank) Len() int           { return len(a) }
func (a byRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRank) Less(i, j int) bool { return a[i] > a[j] } // Sort descending

// Pre-defined AI profiles
var aiProfiles = map[string]AIProfile{
	"Tight-Aggressive": {
		Name:               "Tight-Aggressive",
		PlayHandThreshold:  20,   // Plays only top 20% of hands
		RaiseHandThreshold: 25,   // Raises with top 15% of hands
		BluffingFrequency:  0.15, // Bluffs 15% of the time
		AggressionFactor:   0.7,  // High aggression
		MinRaiseMultiplier: 2.5,
		MaxRaiseMultiplier: 4.0,
	},
	"Loose-Aggressive": {
		Name:               "Loose-Aggressive",
		PlayHandThreshold:  10,   // Plays top 40% of hands
		RaiseHandThreshold: 20,   // Raises with top 25% of hands
		BluffingFrequency:  0.35, // Bluffs 35% of the time
		AggressionFactor:   0.9,  // Very high aggression
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.5,
	},
	"Tight-Passive": {
		Name:               "Tight-Passive",
		PlayHandThreshold:  22,   // Plays only top 18% of hands
		RaiseHandThreshold: 28,   // Raises only with premium hands
		BluffingFrequency:  0.05, // Rarely bluffs
		AggressionFactor:   0.3,  // Low aggression, prefers calling
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 2.5,
	},
	"Loose-Passive": {
		Name:               "Loose-Passive",
		PlayHandThreshold:  8,    // Plays a wide range of hands
		RaiseHandThreshold: 24,   // Rarely raises
		BluffingFrequency:  0.10, // Bluffs occasionally with draws
		AggressionFactor:   0.2,  // Very passive, calls a lot
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.0,
	},
}

// GetCPUAction now uses the player's profile to make decisions.
func (g *Game) GetCPUAction(player *Player, r *rand.Rand) PlayerAction {

	strength := g.handEvaluator(g, player)
	canCheck := player.CurrentBet == g.BetToCall

	time.Sleep(g.CPUThinkTime())

	// --- Pre-Flop Logic ---
	if g.Phase == PhasePreFlop {
		if strength < player.Profile.PlayHandThreshold {
			return PlayerAction{Type: ActionFold}
		}
		if strength >= player.Profile.RaiseHandThreshold {
			return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
		}
		return PlayerAction{Type: ActionCall}
	}

	// --- Post-Flop Logic ---

	// 1. Bluffing Logic
	isBluffing := r.Float64() < player.Profile.BluffingFrequency
	if isBluffing && strength < float64(poker.OnePair) {
		if canCheck {
			return PlayerAction{Type: ActionBet, Amount: g.Pot / 2}
		}
		return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
	}

	// 2. Value Betting/Raising Logic (based on hand strength)
	if strength >= float64(poker.TwoPair) { // Strong hands
		if r.Float64() < player.Profile.AggressionFactor {
			return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
		} else {
			return PlayerAction{Type: ActionCall} // Slow play
		}
	} else if strength >= float64(poker.OnePair) { // Decent hands
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionCall}
	} else { // Weak hands / draws
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		// Decide whether to fold or call based on pot odds (simplified)
		potOdds := float64(g.BetToCall) / float64(g.Pot+g.BetToCall)
		// Simplified equity - just bluffing frequency for now
		if potOdds < player.Profile.BluffingFrequency*0.5 { // Call if pot odds are good
			return PlayerAction{Type: ActionCall}
		}
		return PlayerAction{Type: ActionFold}
	}
}

// evaluateHandStrength calculates a score for a hand, used for AI decisions.
func evaluateHandStrength(g *Game, player *Player) float64 {
	// Post-Flop Evaluation (based on the current best hand rank)
	if g.Phase > PhasePreFlop {
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards, g.Rules)
		return float64(highHand.Rank)
	}

	// Pre-Flop Evaluation (based on hole cards potential - Chen Formula inspired)
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
	if len(hand) >= 3 {
		if hand[0].Rank == hand[1].Rank || hand[0].Rank == hand[2].Rank || hand[1].Rank == hand[2].Rank {
			pairRank := hand[0].Rank
			if hand[1].Rank == hand[2].Rank {
				pairRank = hand[1].Rank
			}
			score += 15 + float64(pairRank) // Major bonus for pairs
		}
	} else if len(hand) == 2 {
		if hand[0].Rank == hand[1].Rank {
			score += 15 + float64(hand[0].Rank)
		}
	}

	// 3. Suited bonus
	if len(hand) >= 3 {
		if hand[0].Suit == hand[1].Suit || hand[0].Suit == hand[2].Suit || hand[1].Suit == hand[2].Suit {
			score += 2
		}
	} else if len(hand) == 2 {
		if hand[0].Suit == hand[1].Suit {
			score += 2
		}
	}

	// 4. Connector bonus (gap calculation)
	if len(hand) >= 3 {
		ranks := []poker.Rank{hand[0].Rank, hand[1].Rank, hand[2].Rank}
		// Sort ranks in descending order for consistent gap calculation
		sort.Sort(byRank(ranks))

		// Check for connectors
		if ranks[0] == ranks[1]+1 && ranks[1] == ranks[2]+1 { // 3-card straight
			score += 5
		} else if (ranks[0] == ranks[1]+1) || (ranks[1] == ranks[2]+1) { // 2-card connector
			score += 2
		}

		// Bonus for cards being higher than T and close together
		if ranks[0] >= poker.Ten && (ranks[0]-ranks[2] < 5) {
			score += 1
		}
	} else if len(hand) == 2 {
		ranks := []poker.Rank{hand[0].Rank, hand[1].Rank}
		sort.Sort(byRank(ranks))
		if ranks[0] == ranks[1]+1 { // connectors
			score += 2
		}
		if ranks[0] >= poker.Ten && (int(ranks[0])-int(ranks[1]) < 5) {
			score += 1
		}
	}

	return score
}
