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
		// --- Welcome Message ---
		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Println("\nStarting the game!")

		// --- Integrate evaluation results ---

		// 1. Create and shuffle the deck
		deck := poker.NewDeck()
		deck.Shuffle()

		// 2. Create players
		players := make([]*game.Player, 6)
		players[0] = &game.Player{Name: "YOU"}
		for i := 1; i < 6; i++ {
			players[i] = &game.Player{Name: fmt.Sprintf("CPU %d", i)}
		}

		// 3. Deal hands to players (3 cards each)
		for i := 0; i < 3; i++ {
			for _, p := range players {
				card, _ := deck.Deal()
				p.Hand = append(p.Hand, card)
			}
		}

		// 4. Deal community cards (5 cards)
		communityCards := make([]poker.Card, 5)
		for i := 0; i < 5; i++ {
			communityCards[i], _ = deck.Deal()
		}

		// 5. Display the results
		fmt.Println("\n--- Static Game State with Evaluation ---")

		var communityCardStrings []string
		for _, c := range communityCards {
			communityCardStrings = append(communityCardStrings, c.String())
		}
		fmt.Printf("Board: %s\n", strings.Join(communityCardStrings, " "))

		fmt.Println("\nPlayers' Hands & Results:")
		for _, player := range players {
			// Call the evaluation function
			highHand, lowHand := poker.EvaluateHand(player.Hand, communityCards)

			// Build the result string using the new String() method
			var resultStrings []string
			if highHand != nil {
				resultStrings = append(resultStrings, fmt.Sprintf("High: %s", highHand.String()))
			}
			if lowHand != nil {
				var lowHandRanks []string
				for _, c := range lowHand.Cards {
					lowHandRanks = append(lowHandRanks, c.Rank.String())
				}
				resultStrings = append(resultStrings, fmt.Sprintf("Low: %s-High", strings.Join(lowHandRanks, "-")))
			}

			fmt.Printf("- %-7s: %v -> %s\n", player.Name, player.Hand, strings.Join(resultStrings, " | "))
		}
		fmt.Println("\n-----------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}
