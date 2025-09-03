package engine

import (
	"fmt"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// DistributionResult is a data structure that holds the outcome of a pot
// distribution for a single player. It's used to communicate the results
// back to the UI or logger.
type DistributionResult struct {
	PlayerName string // The name of the player who won a share of the pot.
	AmountWon  int    // The total amount of chips won by the player.
	HandDesc   string // A description of the winning hand (e.g., "High: Flush", "Low: 8-7-6-5-4").
}

// PotTier represents a single pot (either the main pot or a side pot) that is
// created when one or more players are all-in. Each tier has a specific amount
// and a list of players who are eligible to win it.
type PotTier struct {
	Amount  int       // The total chip amount in this specific pot tier.
	Players []*Player // The slice of players who are eligible to win this pot tier.
	MaxBet  int       // The maximum bet amount that players in this tier have contributed.
}

// AwardPotToLastPlayer handles the simple scenario where all but one player have
// folded. The remaining player wins the entire pot without a showdown.
func (g *Game) AwardPotToLastPlayer() []DistributionResult {
	var winner *Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
			winner = p
			break
		}
	}

	if winner != nil {
		winner.Chips += g.Pot
		result := DistributionResult{
			PlayerName: winner.Name,
			AmountWon:  g.Pot,
			HandDesc:   "takes the pot as the last remaining player",
		}
		g.Pot = 0
		return []DistributionResult{result}
	}
	return []DistributionResult{}
}

// DistributePot is the core function for calculating and awarding the pot(s) at
// the end of a hand. It correctly handles complex scenarios including multiple
// side pots for all-in players and High-Low split pots.
//
// The process is as follows:
//  1. It identifies all players who contributed to the pot and are eligible for a showdown.
//  2. It creates "bet tiers" based on the unique amounts players have bet. For example,
//     if P1 bets 100, P2 bets 200, and P3 bets 200, the tiers are 100 and 200.
//  3. It iterates through these tiers to build one or more `PotTier` objects (side pots).
//     The first pot tier's amount is calculated from the lowest all-in amount, and only
//     players who bet at least that much are eligible. Subsequent tiers are built from
//     the remaining amounts.
//  4. It then distributes each `PotTier` individually. For each pot, it finds the best
//     high hand and, if applicable, the best low hand among the eligible players.
//  5. It splits the pot tier's amount among the high and low winners (or scoops to high
//     if no qualifying low). It handles ties by splitting the shares further.
//  6. Finally, it aggregates the results into a slice of DistributionResult for display.
func (g *Game) DistributePot() []DistributionResult {
	var results []DistributionResult
	showdownPlayers := g.getShowdownPlayers()

	if len(showdownPlayers) == 0 {
		return results
	}

	// Create a list of all players who contributed to the pot.
	var allContributors []*Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated && p.TotalBetInHand > 0 {
			allContributors = append(allContributors, p)
		}
	}

	// Create a set of unique bet amounts from all contributors to define the tiers.
	betTiers := make(map[int]bool)
	for _, p := range allContributors {
		betTiers[p.TotalBetInHand] = true
	}

	// Create a sorted list of the bet tiers (from smallest to largest bet).
	var sortedTiers []int
	for bet := range betTiers {
		sortedTiers = append(sortedTiers, bet)
	}
	sort.Ints(sortedTiers)

	var pots []PotTier
	lastBet := 0

	logrus.Debugf("DistributePot: Initial Pot: %d, All Contributors: %v, Bet Tiers: %v", g.Pot, getPlayerNames(allContributors), sortedTiers)

	// Build the main and side pots based on the bet tiers.
	for _, tierBet := range sortedTiers {
		contribution := tierBet - lastBet
		if contribution <= 0 {
			continue
		}

		// Count players who contributed at least this much.
		numPlayersInTier := 0
		for _, p := range allContributors {
			if p.TotalBetInHand >= tierBet {
				numPlayersInTier++
			}
		}
		tierAmount := contribution * numPlayersInTier

		// Find which of the showdown players are eligible for this tier.
		var eligiblePlayers []*Player
		for _, sp := range showdownPlayers {
			if sp.TotalBetInHand >= tierBet {
				eligiblePlayers = append(eligiblePlayers, sp)
			}
		}

		if tierAmount > 0 && len(eligiblePlayers) > 0 {
			pots = append(pots, PotTier{
				Amount:  tierAmount,
				Players: eligiblePlayers,
				MaxBet:  tierBet,
			})
			logrus.Debugf(
				"  New PotTier created: Amount: %d, MaxBet: %d, Players: %v",
				tierAmount, tierBet, getPlayerNames(eligiblePlayers),
			)
			if len(eligiblePlayers) == 1 {
				logrus.Warnf(
					"  Single player %s eligible for PotTier with amount %d", eligiblePlayers[0].Name, tierAmount,
				)
			}
		}
		lastBet = tierBet
	}

	winnerChipMap := make(map[string]int)
	winnerHandDescMap := make(map[string]string)

	// Distribute each pot tier, starting with the main pot.
	for _, pot := range pots {
		logrus.Debugf("Distributing PotTier: Amount: %d, MaxBet: %d, Eligible Players: %v", pot.Amount, pot.MaxBet, getPlayerNames(pot.Players))
		highWinners, bestHighHand := findBestHighHand(pot.Players, g)
		lowWinners, bestLowHand := findBestLowHand(pot.Players, g)
		logrus.Debugf(
			"DistributePot: High Winners: %v, Best High Hand: %s",
			getPlayerNames(highWinners), bestHighHand,
		)
		logrus.Debugf(
			"DistributePot: Low Winners: %v, Best Low Hand: %s",
			getPlayerNames(lowWinners), bestLowHand,
		)

		// Check for a Hi-Lo split if the game rules allow it and there's a qualifying low hand.
		if g.Rules.LowHand.Enabled && len(lowWinners) > 0 {
			// Split the pot between high and low winners.
			lowPot := pot.Amount / 2
			highPot := pot.Amount - lowPot

			logrus.Debugf("  Split Pot: lowPot: %d, highPot: %d", lowPot, highPot)

			// Distribute the low half of the pot.
			lowShare := lowPot / len(lowWinners)
			var lowHandRanks []string
			for _, c := range bestLowHand.Cards {
				lowHandRanks = append(lowHandRanks, c.Rank.String())
			}
			if len(lowHandRanks) > 0 && lowHandRanks[0] == poker.Ace.String() {
				lowHandRanks = append(lowHandRanks[1:], lowHandRanks[0])
			}
			lowHandDesc := fmt.Sprintf("Low: %s-High", strings.Join(lowHandRanks, "-"))

			for _, winner := range lowWinners {
				winner.Chips += lowShare
				winnerChipMap[winner.Name] += lowShare
				winnerHandDescMap[winner.Name] = lowHandDesc
				logrus.Debugf("    %s wins %d from low pot", winner.Name, lowShare)
			}

			// Distribute the high half of the pot.
			highShare := highPot / len(highWinners)
			highHandDesc := fmt.Sprintf("High: %s", bestHighHand.String())
			for _, winner := range highWinners {
				winner.Chips += highShare
				winnerChipMap[winner.Name] += highShare
				// If a player won both high and low, they "scoop" the pot.
				if desc, exists := winnerHandDescMap[winner.Name]; exists && strings.HasPrefix(desc, "Low") {
					winnerHandDescMap[winner.Name] = fmt.Sprintf("Scoop! %s, %s", highHandDesc, desc)
				} else {
					winnerHandDescMap[winner.Name] = highHandDesc
				}
				logrus.Debugf("    %s wins %d from high pot", winner.Name, highShare)
			}
		} else {
			// If no qualifying low hand, the high hand "scoops" the entire pot.
			highShare := pot.Amount / len(highWinners)
			highHandDesc := fmt.Sprintf("High: %s (Scoop)", bestHighHand.String())
			for _, winner := range highWinners {
				winner.Chips += highShare
				winnerChipMap[winner.Name] += highShare
				winnerHandDescMap[winner.Name] = highHandDesc
				logrus.Debugf("    %s scoops %d from pot", winner.Name, highShare)
			}
		}
	}

	// Aggregate the winnings into the final result list.
	for name, amount := range winnerChipMap {
		results = append(results, DistributionResult{
			PlayerName: name,
			AmountWon:  amount,
			HandDesc:   winnerHandDescMap[name],
		})
	}

	g.Pot = 0
	logrus.Debugf("DistributePot: Final results: %+v", results)
	return results
}

// getShowdownPlayers returns a slice of players who are still active in the
// hand and thus eligible to participate in the showdown.
func (g *Game) getShowdownPlayers() []*Player {
	var active []*Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
			active = append(active, p)
		}
	}
	return active
}

// findBestHighHand iterates through a list of players and determines who has the
// best high hand according to the game's rules. It returns the winning player(s)
// (in case of a tie) and the best hand result.
func findBestHighHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		highHand, _ := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if highHand == nil {
			continue
		}
		if bestHand == nil || compareHandResults(highHand, bestHand) == 1 {
			bestHand = highHand
			winners = []*Player{p}
		} else if compareHandResults(highHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

// findBestLowHand iterates through a list of players and determines who has the
// best qualifying low hand. It returns the winning player(s) and the best low hand.
// If no player has a qualifying low hand, it returns nil.
func findBestLowHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		_, lowHand := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if lowHand == nil {
			continue
		}
		// For low hands, a lower result is better.
		if bestHand == nil || compareHandResults(lowHand, bestHand) == -1 {
			bestHand = lowHand
			winners = []*Player{p}
		} else if compareHandResults(lowHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

// compareHandResults compares two hand results to determine which is stronger.
// It first compares by HandRank, then by HighValues for tie-breaking.
// Returns 1 if h1 > h2, -1 if h1 < h2, 0 if h1 == h2.
func compareHandResults(h1, h2 *poker.HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	// Ranks are the same, compare kickers.
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0 // Hands are identical.
}

// getPlayerNames is a helper function for logging, returning a slice of player names.
func getPlayerNames(players []*Player) []string {
	names := make([]string, len(players))
	for i, p := range players {
		names[i] = p.Name
	}
	return names
}
