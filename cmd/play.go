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
		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Println("\nStarting the game!")

		playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
		initialChips := game.BigBlindAmt * 300 // 300BB
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

// runInteractiveBettingRound has a more robust loop to handle betting correctly.
func runInteractiveBettingRound(g *game.Game) {
	g.PrepareNewBettingRound()

	if g.CountActivePlayers() < 2 {
		return
	}

	lastAggressorPos := -1
	if g.Phase == game.PhasePreFlop {
		lastAggressorPos = (g.DealerPos + 2) % len(g.Players) // BB is the initial aggressor
	}

	turnsTaken := 0
	for {
		if turnsTaken >= g.CountActivePlayers() {
			if g.Phase == game.PhasePreFlop && g.CurrentTurnPos == lastAggressorPos && g.BetToCall == game.BigBlindAmt {
				// Allow BB to act
			} else {
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
			if wasAggressive {
				lastAggressorPos = g.CurrentTurnPos
				turnsTaken = 1 // Reset counter after an aggressive action
			} else {
				turnsTaken++
			}
		}

		g.CurrentTurnPos = (g.CurrentTurnPos + 1) % len(g.Players)
	}
}

func showdownResults(g *game.Game) {
	fmt.Println("\n--- SHOWDOWN ---")

	// Get distribution results first to identify winners
	distributionResults := g.DistributePot()

	// Create a lookup map for winners
	winnerMap := make(map[string][]string) // map[playerName] -> ["High Winner", "Low Winner"]
	for _, result := range distributionResults {
		winType := ""
		if strings.HasPrefix(result.HandDesc, "High") || strings.HasPrefix(result.HandDesc, "takes") {
			winType = "High Winner"
		} else {
			winType = "Low Winner"
		}
		winnerMap[result.PlayerName] = append(winnerMap[result.PlayerName], winType)
	}

	// Show everyone's hands and winner status
	for _, player := range g.Players {
		if player.Status == game.PlayerStatusFolded {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards)

		// Build the full hand description
		handDesc := highHand.String()
		if lowHand != nil {
			var lowHandRanks []string
			for _, c := range lowHand.Cards {
				lowHandRanks = append(lowHandRanks, c.Rank.String())
			}
			if len(lowHandRanks) > 0 && lowHandRanks[0] == "A" {
				lowHandRanks = append(lowHandRanks[1:], lowHandRanks[0])
			}
			handDesc += fmt.Sprintf(" | Low: %s-High", strings.Join(lowHandRanks, "-"))
		}

		// Add winner status if they won
		winnerStatus := ""
		if statuses, ok := winnerMap[player.Name]; ok {
			winnerStatus = fmt.Sprintf(" (%s)", strings.Join(statuses, " & "))
		}

		fmt.Printf("- %-7s: %v -> %s%s\n", player.Name, player.Hand, handDesc, winnerStatus)
	}

	// Then, show the pot distribution results
	fmt.Println("\n--- POT DISTRIBUTION ---")
	for _, result := range distributionResults {
		fmt.Printf("%s won %d chips with %s\n", result.PlayerName, result.AmountWon, result.HandDesc)
	}
	fmt.Println("------------------------")
}

func init() {
	rootCmd.AddCommand(playCmd)
}
