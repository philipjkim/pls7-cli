package cli

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"
)

// DisplayGameState prints the current state of the game board and players.
func DisplayGameState(g *game.Game) {
	if !g.DevMode {
		clearScreen()
	}

	var output string // Concat all output here and print at once not to be mixed with other logs

	phaseName := strings.ToUpper(g.Phase.String())
	output += fmt.Sprintf("\n\n--- %s (%s) | HAND #%d | PHASE: %s | POT: %s | BLINDS: %s/%s ---\n",
		g.Rules.Abbreviation, g.Difficulty, g.HandCount, phaseName,
		util.FormatNumber(g.Pot), util.FormatNumber(game.SmallBlindAmt), util.FormatNumber(game.BigBlindAmt),
	)

	var communityCardStrings []string
	for _, c := range g.CommunityCards {
		communityCardStrings = append(communityCardStrings, c.String())
	}
	output += fmt.Sprintf("Board: %s\n\n", strings.Join(communityCardStrings, " "))

	totalChips := g.Pot
	output += fmt.Sprintln("Players:")
	for i, p := range g.Players {
		// --- NEW: Skip eliminated players from the display ---
		if p.Status == game.PlayerStatusEliminated {
			continue
		}
		// --- END OF NEW PART ---

		indicator := "  "
		if i == g.DealerPos {
			indicator = "D "
		}
		if i == g.CurrentTurnPos {
			indicator = "> "
		}

		status := ""
		if p.Status == game.PlayerStatusFolded {
			status = "(Folded)"
		}
		if p.Status == game.PlayerStatusAllIn && p.CurrentBet == 0 {
			status = "(All In)"
		}

		handInfo := ""
		if !p.IsCPU || g.DevMode {
			var handStrings []string
			for _, c := range p.Hand {
				handStrings = append(handStrings, c.String())
			}
			handInfo = fmt.Sprintf("| Hand: %s", strings.Join(handStrings, " "))

			if g.Phase > game.PhasePreFlop {
				highRank, lowRank := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
				rankInfo := fmt.Sprintf(" | High: %s", highRank.String())
				if g.Rules.LowHand.Enabled && lowRank != nil {
					rankInfo += fmt.Sprintf(", Low: %s", lowRank.String())
				}
				handInfo += rankInfo
			}
		}

		actionInfo := ""
		if p.Status != game.PlayerStatusEliminated {
			actionInfo = fmt.Sprintf(", Current Bet: %-6s", util.FormatNumber(p.CurrentBet))
			if p.LastActionDesc != "" && i != g.CurrentTurnPos {
				actionInfo += fmt.Sprintf(" - %s", p.LastActionDesc)
			}
		}
		logrus.Debugf(
			"Player %s: Status: %v, Current Bet: %s, Last Action: %s, actionInfo: [%s]",
			p.Name, p.Status, util.FormatNumber(p.CurrentBet), p.LastActionDesc, actionInfo,
		)

		line := fmt.Sprintf("%s%-7s: Chips: %-9s%s %s %s", indicator, p.Name, util.FormatNumber(p.Chips), actionInfo, status, handInfo)
		output += fmt.Sprintln(strings.TrimSpace(line))

		// Display outs for the player in dev mode
		if g.CanShowOuts(p) {
			hasOuts, outsInfo := poker.CalculateOuts(p.Hand, g.CommunityCards, g.Rules)
			if hasOuts {
				sort.Slice(outsInfo.AllOuts, func(i, j int) bool {
					if outsInfo.AllOuts[i].Suit != outsInfo.AllOuts[j].Suit {
						return outsInfo.AllOuts[i].Suit < outsInfo.AllOuts[j].Suit
					}
					return outsInfo.AllOuts[i].Rank < outsInfo.AllOuts[j].Rank
				})
				output += formatOuts(outsInfo)

				amountToCall := g.BetToCall - p.CurrentBet
				output += formatEquities(g.Pot, amountToCall, len(outsInfo.AllOuts), g.Phase)
			}
		}

		// Calculate total chips for the game
		if p.Status != game.PlayerStatusEliminated {
			totalChips += p.Chips
		}
	}

	if totalChips != game.BigBlindAmt*300*len(g.Players) {
		logrus.Warnf(
			"Total chips mismatch: expected %s, got %s",
			util.FormatNumber(game.BigBlindAmt*300*len(g.Players)),
			util.FormatNumber(totalChips),
		)
	} else {
		logrus.Debugf(
			"Total chips match expected value: %s",
			util.FormatNumber(game.BigBlindAmt*300*len(g.Players)),
		)
	}

	output += fmt.Sprintln("-------------------------------------------------")
	fmt.Print(output)
}

// formatOuts formats the outs cards for display.
func formatOuts(outsInfo *poker.OutsInfo) string {
	result := "\tAll Outs: "
	outStrings := make([]string, 0, len(outsInfo.AllOuts))
	for _, c := range outsInfo.AllOuts {
		outStrings = append(outStrings, c.String())
	}
	result += strings.Join(outStrings, ", ")

	if outsInfo.OutsPerHandRank != nil {
		result += "\n\tOuts by Hand Rank:\n"
		for rank, outs := range outsInfo.OutsPerHandRank {
			if len(outs) == 0 {
				continue
			}
			result += fmt.Sprintf("\t\t%s: ", rank.String())
			outRankStrings := make([]string, 0, len(outs))
			for _, c := range outs {
				outRankStrings = append(outRankStrings, c.String())
			}
			result += strings.Join(outRankStrings, ", ") + "\n"
		}
	}
	return result
}

func formatEquities(pot, amountToCall, numOuts int, phase game.GamePhase) string {
	numCommunityCards := 0
	if phase == game.PhaseFlop {
		numCommunityCards = 3
	} else if phase == game.PhaseTurn {
		numCommunityCards = 4
	} else {
		// For Pre-Flop and River, we don't calculate outs or equities
		return ""
	}

	return fmt.Sprintf("\n\t- Break-even equity based on pot odds: %.2f\n\t- Equity: %.2f\n",
		poker.CalculateBreakEvenEquityBasedOnPotOdds(pot, amountToCall),
		poker.CalculateEquity(numCommunityCards, numOuts),
	)
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
