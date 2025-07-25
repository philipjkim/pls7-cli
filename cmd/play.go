package cmd

import (
	"fmt"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
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

		// --- Step 3 Goal: Simulate a static game state ---

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

		// 5. Create the game state object
		staticGame := &game.Game{
			Players:        players,
			CommunityCards: communityCards,
		}

		// 6. Display the static game state
		cli.DisplayStaticGameState(staticGame)
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	// Here you will define your flags and configuration settings.
	// Example: playCmd.Flags().IntP("players", "p", 6, "Number of players to participate")
}
