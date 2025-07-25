package cmd

import (
	"fmt"
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
		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Println("\nStarting the game!")

		playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
		initialChips := 10000
		g := game.NewGame(playerNames, initialChips)

		g.StartNewHand()

		// Main game loop for a single hand
		for g.Phase != game.PhaseHandOver {
			// The betting round now handles its own internal loop of player actions
			g.RunBettingRound()

			// Stop the hand if only one player is left
			if g.CountActivePlayers() <= 1 {
				// In a real game, we'd award the pot here. For now, just end.
				break
			}

			g.Advance()

			if g.Phase == game.PhaseShowdown {
				showdownResults(g)
				g.Advance() // Move to HandOver
			}
		}

		fmt.Println("\nGame hand finished.")
	},
}

func showdownResults(g *game.Game) {
	fmt.Println("\n--- SHOWDOWN ---")
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
