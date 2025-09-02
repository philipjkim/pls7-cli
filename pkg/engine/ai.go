package engine

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"sort"
	"time"
)

// byRank is a helper type that implements the sort.Interface for a slice of
// poker.Rank, allowing them to be sorted. It sorts in descending order (Ace high).
type byRank []poker.Rank

func (a byRank) Len() int           { return len(a) }
func (a byRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRank) Less(i, j int) bool { return a[i] > a[j] } // Sort descending

// aiProfiles contains a set of predefined AI personalities that dictate how a CPU
// player behaves. Each profile has different thresholds for playing, raising,
// and bluffing, creating varied opponent styles.
var aiProfiles = map[string]AIProfile{
	"Tight-Aggressive": {
		Name:               "Tight-Aggressive",
		PlayHandThreshold:  20,   // Plays only the top 20% of starting hands.
		RaiseHandThreshold: 25,   // Raises with the top 15% of hands.
		BluffingFrequency:  0.15, // Bluffs occasionally.
		AggressionFactor:   0.7,  // Highly likely to bet or raise with strong hands.
		MinRaiseMultiplier: 2.5,
		MaxRaiseMultiplier: 4.0,
	},
	"Loose-Aggressive": {
		Name:               "Loose-Aggressive",
		PlayHandThreshold:  10,   // Plays a wide range of hands (top 40%).
		RaiseHandThreshold: 20,   // Raises often.
		BluffingFrequency:  0.35, // Bluffs frequently.
		AggressionFactor:   0.9,  // Very aggressive.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.5,
	},
	"Tight-Passive": {
		Name:               "Tight-Passive",
		PlayHandThreshold:  22,   // Very selective with starting hands.
		RaiseHandThreshold: 28,   // Rarely raises, only with premium hands.
		BluffingFrequency:  0.05, // Almost never bluffs.
		AggressionFactor:   0.3,  // Prefers to call rather than bet or raise.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 2.5,
	},
	"Loose-Passive": {
		Name:               "Loose-Passive",
		PlayHandThreshold:  8,    // Plays many hands (calling station).
		RaiseHandThreshold: 24,   // Rarely raises.
		BluffingFrequency:  0.10, // Bluffs infrequently.
		AggressionFactor:   0.2,  // Very passive, calls often, folds to aggression.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.0,
	},
}

// GetCPUAction determines the action for an AI-controlled player based on their
// assigned profile and the current game state. This method implements the
// ActionProvider interface for CPU players.
// The logic is divided into pre-flop and post-flop stages.
func (g *Game) GetCPUAction(player *Player, r *rand.Rand) PlayerAction {
	// First, evaluate the strength of the player's hand.
	strength := g.handEvaluator(g, player)
	canCheck := player.CurrentBet == g.BetToCall

	// Simulate thinking time for a more realistic game pace.
	time.Sleep(g.CPUThinkTime())

	// --- Pre-Flop Logic ---
	// Based on a simplified hand strength score.
	if g.Phase == PhasePreFlop {
		// Fold if hand strength is below the profile's play threshold.
		if strength < player.Profile.PlayHandThreshold {
			return PlayerAction{Type: ActionFold}
		}
		// Raise if hand strength is above the profile's raise threshold.
		if strength >= player.Profile.RaiseHandThreshold {
			return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
		}
		// Otherwise, just call.
		return PlayerAction{Type: ActionCall}
	}

	// --- Post-Flop Logic ---
	// Based on the actual rank of the 5-card hand.

	// 1. Bluffing Logic: Decide whether to bluff based on profile frequency.
	// A bluff is only attempted with a weak hand (less than OnePair).
	isBluffing := r.Float64() < player.Profile.BluffingFrequency
	if isBluffing && strength < float64(poker.OnePair) {
		if canCheck {
			// A "probe" bet when checked to.
			return PlayerAction{Type: ActionBet, Amount: g.Pot / 2}
		}
		// A bluff raise.
		return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
	}

	// 2. Value Betting/Raising Logic (based on hand strength).
	if strength >= float64(poker.TwoPair) { // Strong hands (Two Pair or better).
		// Decide whether to be aggressive or "slow play" (trap).
		if r.Float64() < player.Profile.AggressionFactor {
			return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
		} else {
			return PlayerAction{Type: ActionCall} // Slow play.
		}
	} else if strength >= float64(poker.OnePair) { // Decent, but vulnerable hands.
		// Prefer to see the next card cheaply.
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionCall}
	} else { // Weak hands / draws.
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		// Decide whether to fold or call based on a simplified version of pot odds.
		potOdds := float64(g.BetToCall) / float64(g.Pot+g.BetToCall)
		// A very rough estimation of equity.
		if potOdds < player.Profile.BluffingFrequency*0.5 { // Call if pot odds are favorable.
			return PlayerAction{Type: ActionCall}
		}
		return PlayerAction{Type: ActionFold}
	}
}

// evaluateHandStrength calculates a numerical score for a player's hand to guide
// AI decision-making. The evaluation method differs between pre-flop and post-flop.
//
// Post-flop, the score is simply the rank of the player's best 5-card hand.
//
// Pre-flop, it uses a custom scoring system to assess the potential of the hole
// cards, considering:
// - High card values (points for cards Ten and above).
// - A significant bonus for pairs.
// - A small bonus for suited cards.
// - A bonus for connected cards (cards in sequence).
func evaluateHandStrength(g *Game, player *Player) float64 {
	// Post-Flop: The strength is the actual rank of the hand.
	if g.Phase > PhasePreFlop {
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards, g.Rules)
		if highHand != nil {
			return float64(highHand.Rank)
		}
		return 0
	}

	// Pre-Flop: Evaluate potential based on hole cards using a custom heuristic.
	var score float64
	hand := player.Hand

	// 1. High card points for cards Ten or higher.
	rankPoints := map[poker.Rank]float64{
		poker.Ace: 10, poker.King: 8, poker.Queen: 7, poker.Jack: 6, poker.Ten: 5,
	}
	for _, c := range hand {
		score += rankPoints[c.Rank]
	}

	// 2. Add a large bonus for having a pair in the hole cards.
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

	// 3. Add a small bonus if any cards are suited.
	if len(hand) >= 3 {
		if hand[0].Suit == hand[1].Suit || hand[0].Suit == hand[2].Suit || hand[1].Suit == hand[2].Suit {
			score += 2
		}
	} else if len(hand) == 2 {
		if hand[0].Suit == hand[1].Suit {
			score += 2
		}
	}

	// 4. Add a bonus for card connectivity (potential to make a straight).
	if len(hand) >= 3 {
		ranks := []poker.Rank{hand[0].Rank, hand[1].Rank, hand[2].Rank}
		// Sort ranks in descending order for consistent gap calculation.
		sort.Sort(byRank(ranks))

		// Check for connectors.
		if ranks[0] == ranks[1]+1 && ranks[1] == ranks[2]+1 { // 3-card straight
			score += 5
		} else if (ranks[0] == ranks[1]+1) || (ranks[1] == ranks[2]+1) { // 2-card connector
			score += 2
		}

		// Bonus for cards being high and close together.
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
