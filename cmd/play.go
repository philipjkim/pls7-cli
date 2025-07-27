package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
	"pls7-cli/pkg/poker"
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
		initialChips := 200_000
		g := game.NewGame(playerNames, initialChips)

		g.StartNewHand()

		// Main game loop for a single hand
		for g.Phase != game.PhaseHandOver {
			runInteractiveBettingRound(g)

			if g.CountActivePlayers() <= 1 {
				fmt.Println("\nOnly one player left. Hand is over.")
				break
			}

			g.Advance()

			if g.Phase == game.PhaseShowdown {
				cli.DisplayGameState(g)
				showdownResults(g)
				g.Advance() // Move to HandOver
			}
		}

		fmt.Println("\nGame hand finished.")
	},
}

// runInteractiveBettingRound is the new orchestrator for a betting round.
func runInteractiveBettingRound(g *game.Game) {
	g.PrepareNewBettingRound()
	if g.CountActivePlayers() < 2 {
		return
	}
	numPlayers := len(g.Players)
	playersToAct := g.CountActivePlayers()
	actionCount := 0
	lastAggressorPos := -1
	if g.Phase == game.PhasePreFlop {
		lastAggressorPos = (g.DealerPos + 2) % numPlayers
	}
	for {
		if actionCount >= playersToAct {
			if g.CurrentTurnPos == lastAggressorPos {
				break
			}
			if lastAggressorPos == -1 {
				break
			}
		}
		player := g.Players[g.CurrentTurnPos]
		if player.Status == game.PlayerStatusPlaying {
			cli.DisplayGameState(g)
			var action game.PlayerAction
			if player.IsCPU {
				if player.CurrentBet < g.BetToCall {
					action = game.PlayerAction{Type: game.ActionCall}
				} else {
					action = game.PlayerAction{Type: game.ActionCheck}
				}
			} else {
				action = cli.PromptForAction(g)
			}
			wasAggressive := g.ProcessAction(player, action)
			actionCount++
			if wasAggressive {
				lastAggressorPos = g.CurrentTurnPos
				playersToAct = g.CountActivePlayers()
				actionCount = 1
			}
		}
		g.CurrentTurnPos = (g.CurrentTurnPos + 1) % numPlayers
	}
}

func showdownResults(g *game.Game) {
	fmt.Println("\n--- SHOWDOWN ---")
	// First, show everyone's hands
	for _, player := range g.Players {
		if player.Status == game.PlayerStatusFolded {
			continue
		}
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards)
		fmt.Printf("- %-7s: %v -> %s\n", player.Name, player.Hand, highHand.String())
	}

	// Then, show the pot distribution results
	fmt.Println("\n--- POT DISTRIBUTION ---")
	distributionResults := g.DistributePot()
	for _, result := range distributionResults {
		fmt.Printf("%s wins %d chips with %s\n", result.PlayerName, result.AmountWon, result.HandDesc)
	}
	fmt.Println("------------------------")
}

func init() {
	rootCmd.AddCommand(playCmd)
}
