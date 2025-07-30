package cli

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"strings"
)

// DisplayGameState prints the current state of the game board and players.
func DisplayGameState(g *game.Game, isDevMode bool) {
	if !isDevMode {
		clearScreen()
	}

	var output string // Concat all output here and print at once not to be mixed with other logs

	phaseName := strings.ToUpper(g.Phase.String())
	output += fmt.Sprintf("\n\n--- HAND #%d | PHASE: %s | POT: %s | BLINDS: %s/%s ---\n",
		g.HandCount, phaseName, util.FormatNumber(g.Pot), util.FormatNumber(game.SmallBlindAmt), util.FormatNumber(game.BigBlindAmt))

	var communityCardStrings []string
	for _, c := range g.CommunityCards {
		communityCardStrings = append(communityCardStrings, c.String())
	}
	output += fmt.Sprintf("Board: %s\n\n", strings.Join(communityCardStrings, " "))

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
		if !p.IsCPU || isDevMode {
			var handStrings []string
			for _, c := range p.Hand {
				handStrings = append(handStrings, c.String())
			}
			handInfo = fmt.Sprintf("| Hand: %s", strings.Join(handStrings, " "))

			if g.Phase > game.PhasePreFlop {
				highRank, lowRank := poker.EvaluateHand(p.Hand, g.CommunityCards)
				rankInfo := fmt.Sprintf(" | High: %s", highRank.String())
				if lowRank != nil {
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
	}
	output += fmt.Sprintln("-------------------------------------------------")
	fmt.Print(output)
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
