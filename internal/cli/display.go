package cli

import (
	"fmt"
	"pls7-cli/internal/game"
	"strings"
	"time"
)

// DisplayGameState prints the current state of the game board and players.
func DisplayGameState(g *game.Game) {
	clearScreen() // Clears the console for a fresh display

	phaseName := strings.ToUpper(g.Phase.String())
	fmt.Printf("--- HAND #%d | PHASE: %s | POT: %s | BLINDS: %s/%s ---\n",
		g.HandCount, phaseName, FormatNumber(g.Pot), FormatNumber(game.SmallBlindAmt), FormatNumber(game.BigBlindAmt))

	var communityCardStrings []string
	for _, c := range g.CommunityCards {
		communityCardStrings = append(communityCardStrings, c.String())
	}
	fmt.Printf("Board: %s\n\n", strings.Join(communityCardStrings, " "))

	fmt.Println("Players:")
	for i, p := range g.Players {
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
		} else if p.Status == game.PlayerStatusEliminated {
			status = "(Eliminated)"
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
			actionInfo = fmt.Sprintf(", Current Bet: %-6s", FormatNumber(p.CurrentBet))
			// Show last action unless it's the current player's turn
			if p.LastActionDesc != "" && i != g.CurrentTurnPos {
				actionInfo += fmt.Sprintf(" - %s", p.LastActionDesc)
			}
		}

		line := fmt.Sprintf("%s%-7s: Chips: %-9s%s %s %s", indicator, p.Name, FormatNumber(p.Chips), actionInfo, status, handInfo)
		fmt.Println(strings.TrimSpace(line))
	}
	fmt.Println("-------------------------------------------------")
	time.Sleep(1 * time.Second) // Pause for 2 seconds to let the player see the state
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
