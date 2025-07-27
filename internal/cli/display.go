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
	fmt.Printf("--- HAND #%d | PHASE: %s | POT: %d | BLINDS: %d/%d ---\n",
		g.HandCount, phaseName, g.Pot, game.SmallBlindAmt, game.BigBlindAmt)

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
		}

		handInfo := ""
		if !p.IsCPU {
			var handStrings []string
			for _, c := range p.Hand {
				handStrings = append(handStrings, c.String())
			}
			handInfo = fmt.Sprintf("| Hand: %s", strings.Join(handStrings, " "))
		}

		line := fmt.Sprintf("%s%-7s: Chips: %-7d %s %s", indicator, p.Name, p.Chips, status, handInfo)
		fmt.Println(strings.TrimSpace(line))
	}
	fmt.Println("-------------------------------------------------")
	time.Sleep(1 * time.Second) // Pause for 2 seconds to let the player see the state
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
