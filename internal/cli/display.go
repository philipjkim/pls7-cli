package cli

import (
	"fmt"
	"pls7-cli/internal/game"
)

// DisplayStaticGameState prints a static, hardcoded game state.
func DisplayStaticGameState(g *game.Game) {
	fmt.Println("\n--- Static Game State Simulation ---")

	fmt.Println("\nCommunity Cards:")
	fmt.Printf("Board: %v\n", g.CommunityCards)

	fmt.Println("\nPlayers' Hands:")
	for _, player := range g.Players {
		fmt.Printf("- %s: %v\n", player.Name, player.Hand)
	}
	fmt.Println("\n------------------------------------")
}
