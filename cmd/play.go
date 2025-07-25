package cmd

import (
	"fmt"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
	"pls7-cli/pkg/poker"
	"strings"

	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Starts a new game of PLS7",
	Long:  `Starts a new game of PLS7 with 1 player and 5 CPUs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// --- Welcome Message ---
		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Println("\nStarting the game!")

		// --- Step 5: Run the automated game loop ---

		playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
		initialChips := 10000

		// Create a new game instance
		g := game.NewGame(playerNames, initialChips)

		// Start the first hand
		g.StartNewHand()

		// Main game loop for a single hand
		for {
			cli.DisplayGameState(g)
			if g.Phase == game.PhaseShowdown {
				showdownResults(g)
			}

			isHandOver := g.Advance()
			if isHandOver {
				break
			}
		}

		fmt.Println("\nGame hand finished.")
	},
}

func showdownResults(g *game.Game) {
	fmt.Println("\nPlayers' Hands & Results:")
	for _, player := range g.Players {
		if player.Status == game.PlayerStatusFolded {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards)

		var resultStrings []string
		if highHand != nil {
			resultStrings = append(resultStrings, fmt.Sprintf("High: %s", highHand.String()))
		}
		if lowHand != nil {
			var lowHandRanks []string
			for _, c := range lowHand.Cards {
				lowHandRanks = append(lowHandRanks, c.Rank.String())
			}
			if len(lowHandRanks) > 0 && lowHandRanks[0] == "A" {
				lowHandRanks = append(lowHandRanks[1:], lowHandRanks[0])
			}
			resultStrings = append(resultStrings, fmt.Sprintf("Low: %s-High", strings.Join(lowHandRanks, "-")))
		}

		fmt.Printf("- %-7s: %v -> %s\n", player.Name, player.Hand, strings.Join(resultStrings, " | "))
	}
}

func init() {
	rootCmd.AddCommand(playCmd)
}
