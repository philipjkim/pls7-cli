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
	fmt.Printf("--- HAND #%d | PHASE: %s | POT: %d ---\n", 1, phaseName, g.Pot) // Hand # is static for now

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

		fmt.Printf("%s%-7s: Chips: %-5d %s\n", indicator, p.Name, p.Chips, status)
	}
	fmt.Println("-------------------------------------------------")
	time.Sleep(2 * time.Second) // Pause for cinematic effect
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
