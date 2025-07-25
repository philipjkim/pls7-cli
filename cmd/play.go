package cmd

import (
	"fmt"
	"pls7-cli/pkg/poker"

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

		// --- Step 2 Goal: Test Card & Deck implementation ---
		fmt.Println("\n--- Deck Test ---")
		deck := poker.NewDeck()
		deck.Shuffle()
		fmt.Println("Deck shuffled.")

		fmt.Println("\nDealing 5 cards:")
		for i := 0; i < 5; i++ {
			card, err := deck.Deal()
			if err != nil {
				fmt.Println("Error dealing card:", err)
				break
			}
			fmt.Printf("Dealt card: %s\n", card)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	// Here you will define your flags and configuration settings.
	// Example: playCmd.Flags().IntP("players", "p", 6, "Number of players to participate")
}
