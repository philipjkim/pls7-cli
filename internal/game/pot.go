package game

import (
	"fmt"
	"pls7-cli/pkg/poker"
)

// DistributionResult stores the results of the pot distribution for printing.
type DistributionResult struct {
	PlayerName string
	AmountWon  int
	HandDesc   string
}

// DistributePot calculates and distributes the pot to the winner(s).
// It returns a slice of results for display purposes.
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
	lowWinners, _ := findBestLowHand(showdownPlayers, g.CommunityCards)

	if len(lowWinners) > 0 {
		// Hi-Lo Split
		lowPot := g.Pot / 2
		highPot := g.Pot - lowPot // Handle odd chips

		// Distribute Low Pot
		lowShare := lowPot / len(lowWinners)
		for _, winner := range lowWinners {
			winner.Chips += lowShare
			results = append(results, DistributionResult{
				PlayerName: winner.Name,
				AmountWon:  lowShare,
				HandDesc:   "Low",
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

func (g *Game) getShowdownPlayers() []*Player {
	active := []*Player{}
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded {
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
		if bestHand == nil || compareHandResults(lowHand, bestHand) == -1 { // For low, smaller is better
			bestHand = lowHand
			winners = []*Player{p}
		} else if compareHandResults(lowHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

// compareHandResults compares two hands.
// Returns 1 if h1 > h2, -1 if h1 < h2, 0 if tie.
func compareHandResults(h1, h2 *poker.HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	// Ranks are the same, compare high values (kickers)
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0 // Tie
}
