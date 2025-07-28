package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
	"strings"
)

// DistributionResult stores the results of the pot distribution for printing.
type DistributionResult struct {
	PlayerName string
	AmountWon  int
	HandDesc   string
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

	if len(showdownPlayers) == 1 {
		winner := showdownPlayers[0]
		winner.Chips += g.Pot
		results = append(results, DistributionResult{
			PlayerName: winner.Name,
			AmountWon:  g.Pot,
			HandDesc:   "takes the pot",
		})
		g.Pot = 0
		return results
	}

	highWinners, bestHighHand := findBestHighHand(showdownPlayers, g.CommunityCards)
	lowWinners, bestLowHand := findBestLowHand(showdownPlayers, g.CommunityCards)

	if len(lowWinners) > 0 {
		lowPot := g.Pot / 2
		highPot := g.Pot - lowPot

		// Distribute Low Pot
		lowShare := lowPot / len(lowWinners)
		var lowHandRanks []string
		for _, c := range bestLowHand.Cards {
			lowHandRanks = append(lowHandRanks, c.Rank.String())
		}
		if len(lowHandRanks) > 0 && lowHandRanks[0] == "A" {
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
		}

	} else {
		// High hand scoops the entire pot
		highShare := g.Pot / len(highWinners)
		for _, winner := range highWinners {
			winner.Chips += highShare
			results = append(results, DistributionResult{
				PlayerName: winner.Name,
				AmountWon:  highShare,
				HandDesc:   fmt.Sprintf("High: %s (Scoop)", bestHighHand.String()),
			})
		}
	}

	g.Pot = 0
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

func findBestHighHand(players []*Player, communityCards []poker.Card) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		highHand, _ := poker.EvaluateHand(p.Hand, communityCards)
		if bestHand == nil || compareHandResults(highHand, bestHand) == 1 {
			bestHand = highHand
			winners = []*Player{p}
		} else if compareHandResults(highHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

func findBestLowHand(players []*Player, communityCards []poker.Card) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		_, lowHand := poker.EvaluateHand(p.Hand, communityCards)
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
