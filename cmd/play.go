package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
	"pls7-cli/pkg/poker"
	"strings"
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

		// Main Game Loop (multi-hand)
		for {
			g.StartNewHand()

			// Single Hand Loop
			for g.Phase != game.PhaseHandOver {
				runInteractiveBettingRound(g)

				if g.CountActivePlayers() <= 1 {
					break
				}
				g.Advance()
			}

			// Conclude the hand
			if g.CountActivePlayers() > 1 {
				cli.DisplayGameState(g)
				showdownResults(g)
			} else {
				// Award pot to the last remaining player
				fmt.Println("Awarding pot to the last player...")
			}

			g.CleanupHand()

			if g.CountActivePlayers() <= 1 {
				fmt.Println("\n--- GAME OVER ---")
				break
			}

			fmt.Print("\nPress ENTER to start the next hand, or type 'quit' to exit > ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(input)) == "quit" {
				fmt.Println("Thanks for playing!")
				break
			}
		}
	},
}

// runInteractiveBettingRound has a robust loop to handle all betting scenarios.
func runInteractiveBettingRound(g *game.Game) {
	g.PrepareNewBettingRound()

	if g.CountActivePlayers() < 2 {
		return
	}

	numPlayers := len(g.Players)
	// The position of the player who needs to act last for the round to be complete.
	actionTargetPos := 0

	if g.Phase == game.PhasePreFlop {
		// In Pre-Flop, the Big Blind is the last to act initially.
		actionTargetPos = (g.DealerPos + 2) % numPlayers
	} else {
		// In Post-Flop, the Dealer is the last to act.
		actionTargetPos = g.DealerPos
	}

	for {
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
				// If a player bets or raises, the action must go all the way around again.
				// The new "last to act" is the player right before the aggressor.
				actionTargetPos = (g.CurrentTurnPos - 1 + numPlayers) % numPlayers
			}
		}

		// The round is over when the action has reached the target player.
		if g.CurrentTurnPos == actionTargetPos {
			break
		}

		g.CurrentTurnPos = (g.CurrentTurnPos + 1) % numPlayers
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
