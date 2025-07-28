package cli

import (
	"fmt"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"strings"
)

// DisplayGameState prints the current state of the game board and players.
func DisplayGameState(g *game.Game) {
	phaseName := strings.ToUpper(g.Phase.String())
	fmt.Printf("\n\n--- HAND #%d | PHASE: %s | POT: %s | BLINDS: %s/%s ---\n",
		g.HandCount, phaseName, util.FormatNumber(g.Pot), util.FormatNumber(game.SmallBlindAmt), util.FormatNumber(game.BigBlindAmt))

	var communityCardStrings []string
	for _, c := range g.CommunityCards {
		communityCardStrings = append(communityCardStrings, c.String())
	}
	fmt.Printf("Board: %s\n\n", strings.Join(communityCardStrings, " "))

	fmt.Println("Players:")
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

		handInfo := ""
		if !p.IsCPU {
			var handStrings []string
			for _, c := range p.Hand {
				handStrings = append(handStrings, c.String())
			}
			handInfo = fmt.Sprintf("| Hand: %s", strings.Join(handStrings, " "))
		}

		actionInfo := ""
		if p.Status != game.PlayerStatusEliminated {
			actionInfo = fmt.Sprintf(", Current Bet: %-6s", util.FormatNumber(p.CurrentBet))
			if p.LastActionDesc != "" && i != g.CurrentTurnPos {
				actionInfo += fmt.Sprintf(" - %s", p.LastActionDesc)
			}
		}

		line := fmt.Sprintf("%s%-7s: Chips: %-9s%s %s %s", indicator, p.Name, util.FormatNumber(p.Chips), actionInfo, status, handInfo)
		fmt.Println(strings.TrimSpace(line))
	}
	fmt.Println("-------------------------------------------------")
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
