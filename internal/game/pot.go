package game

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"
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

	// Sort players by their total bet in hand to determine pot tiers
	sort.Slice(showdownPlayers, func(i, j int) bool {
		return showdownPlayers[i].TotalBetInHand < showdownPlayers[j].TotalBetInHand
	})

	var pots []PotTier

	remainingPot := g.Pot
	lastBet := 0

	logrus.Debugf("DistributePot: Initial Pot: %d, Showdown Players: %+v", g.Pot, showdownPlayers)

	for i, p := range showdownPlayers {
		betAmount := p.TotalBetInHand
		contribution := betAmount - lastBet

		logrus.Debugf("Player %s (TotalBetInHand: %d): betAmount: %d, lastBet: %d, contribution: %d", p.Name, p.TotalBetInHand, betAmount, lastBet, contribution)

		if contribution > 0 {
			// Create a new pot tier
			currentPotAmount := 0
			for j := i; j < len(showdownPlayers); j++ {
				currentPotAmount += contribution
			}

			pots = append(pots, PotTier{
				Amount:  currentPotAmount,
				Players: showdownPlayers[i:], // Players eligible for this pot
				MaxBet:  betAmount,
			})
			remainingPot -= currentPotAmount
			logrus.Debugf("  New PotTier created: Amount: %d, MaxBet: %d, Players: %v", currentPotAmount, betAmount, getPlayerNames(showdownPlayers[i:]))
		}
		lastBet = betAmount
		logrus.Debugf("  After processing player %s: pots: %+v, remainingPot: %d, lastBet: %d", p.Name, pots, remainingPot, lastBet)
	}

	// Distribute each pot
	for _, pot := range pots {
		logrus.Debugf("Distributing PotTier: Amount: %d, MaxBet: %d, Eligible Players: %v", pot.Amount, pot.MaxBet, getPlayerNames(pot.Players))
		highWinners, bestHighHand := findBestHighHand(pot.Players, g.CommunityCards, g.LowlessMode)
		lowWinners, bestLowHand := findBestLowHand(pot.Players, g.CommunityCards, g.LowlessMode)

		if !g.LowlessMode && len(lowWinners) > 0 {
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
				results = append(results, DistributionResult{
					PlayerName: winner.Name,
					AmountWon:  lowShare,
					HandDesc:   lowHandDesc,
				})
				logrus.Debugf("    %s wins %d from low pot", winner.Name, lowShare)
			}

			// Distribute High Pot
			highShare := highPot / len(highWinners)
			for _, winner := range highWinners {
				winner.Chips += highShare
				results = append(results, DistributionResult{
					PlayerName: winner.Name,
					AmountWon:  highShare,
					HandDesc:   fmt.Sprintf("High: %s", bestHighHand.String()),
				})
				logrus.Debugf("    %s wins %d from high pot", winner.Name, highShare)
			}
		} else {
			// High hand scoops the entire pot
			highShare := pot.Amount / len(highWinners)
			for _, winner := range highWinners {
				winner.Chips += highShare
				results = append(results, DistributionResult{
					PlayerName: winner.Name,
					AmountWon:  highShare,
					HandDesc:   fmt.Sprintf("High: %s (Scoop)", bestHighHand.String()),
				})
				logrus.Debugf("    %s scoops %d from pot", winner.Name, highShare)
			}
		}
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

func findBestHighHand(players []*Player, communityCards []poker.Card, isLowLess bool) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		highHand, _ := poker.EvaluateHand(p.Hand, communityCards, isLowLess)
		if bestHand == nil || compareHandResults(highHand, bestHand) == 1 {
			bestHand = highHand
			winners = []*Player{p}
		} else if compareHandResults(highHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

func findBestLowHand(players []*Player, communityCards []poker.Card, isLowLess bool) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		_, lowHand := poker.EvaluateHand(p.Hand, communityCards, isLowLess)
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
