package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// DistributionResult stores the results of the pot distribution for printing.
type DistributionResult struct {
	PlayerName string
	AmountWon  int
	HandDesc   string
}

// PotTier represents a single pot (main or side pot) to be distributed.
type PotTier struct {
	Amount  int
	Players []*Player
	MaxBet  int
}

// AwardPotToLastPlayer finds the single remaining player and gives them the pot.
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

// DistributePot calculates and distributes the pot to the winner(s).
func (g *Game) DistributePot() []DistributionResult {
	results := []DistributionResult{}
	showdownPlayers := g.getShowdownPlayers()

	if len(showdownPlayers) == 0 {
		return results
	}

	// Create a list of all players who contributed to the pot
	allContributors := []*Player{}
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated && p.TotalBetInHand > 0 {
			allContributors = append(allContributors, p)
		}
	}

	// Create a set of unique bet amounts from all contributors
	betTiers := make(map[int]bool)
	for _, p := range allContributors {
		betTiers[p.TotalBetInHand] = true
	}

	// Create a sorted list of the bet tiers
	sortedTiers := []int{}
	for bet := range betTiers {
		sortedTiers = append(sortedTiers, bet)
	}
	sort.Ints(sortedTiers)

	var pots []PotTier
	lastBet := 0

	logrus.Debugf("DistributePot: Initial Pot: %d, All Contributors: %v, Bet Tiers: %v", g.Pot, getPlayerNames(allContributors), sortedTiers)

	for _, tierBet := range sortedTiers {
		contribution := tierBet - lastBet
		if contribution <= 0 {
			continue
		}

		// Count players who are part of this tier
		numPlayersInTier := 0
		for _, p := range allContributors {
			if p.TotalBetInHand >= tierBet {
				numPlayersInTier++
			}
		}
		tierAmount := contribution * numPlayersInTier

		// Find showdown players eligible for this tier
		eligiblePlayers := []*Player{}
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

	// Distribute each pot
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

		if g.Rules.LowHand.Enabled && len(lowWinners) > 0 {
			// Split pot for high and low
			lowPot := pot.Amount / 2
			highPot := pot.Amount - lowPot

			logrus.Debugf("  Split Pot: lowPot: %d, highPot: %d", lowPot, highPot)

			// Distribute Low Pot
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

			// Distribute High Pot
			highShare := highPot / len(highWinners)
			highHandDesc := fmt.Sprintf("High: %s", bestHighHand.String())
			for _, winner := range highWinners {
				winner.Chips += highShare
				winnerChipMap[winner.Name] += highShare
				// If the player also won low, append to their description
				if desc, exists := winnerHandDescMap[winner.Name]; exists {
					winnerHandDescMap[winner.Name] = fmt.Sprintf("Scoop: %s, %s", desc, highHandDesc)
				} else {
					winnerHandDescMap[winner.Name] = highHandDesc
				}
				logrus.Debugf("    %s wins %d from high pot", winner.Name, highShare)
			}
		} else {
			// High hand scoops the entire pot
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

// getShowdownPlayers filters for players who are eligible for the showdown.
func (g *Game) getShowdownPlayers() []*Player {
	active := []*Player{}
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
			active = append(active, p)
		}
	}
	return active
}

func findBestHighHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		highHand, _ := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if bestHand == nil || compareHandResults(highHand, bestHand) == 1 {
			bestHand = highHand
			winners = []*Player{p}
		} else if compareHandResults(highHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

func findBestLowHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		_, lowHand := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if lowHand == nil {
			continue
		}
		if bestHand == nil || compareHandResults(lowHand, bestHand) == -1 {
			bestHand = lowHand
			winners = []*Player{p}
		} else if compareHandResults(lowHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

func compareHandResults(h1, h2 *poker.HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0
}

// getPlayerNames is a helper for logging player names.
func getPlayerNames(players []*Player) []string {
	names := make([]string, len(players))
	for i, p := range players {
		names[i] = p.Name
	}
	return names
}
