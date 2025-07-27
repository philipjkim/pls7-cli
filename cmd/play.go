package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"strings"
)

var difficultyStr string // To hold the flag value

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Starts a new game of PLS7",
	Long:  `Starts a new game of PLS7 with 1 player and 5 CPUs.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Printf("\nStarting the game with %s difficulty!\n", difficultyStr)

		var difficulty game.Difficulty
		switch strings.ToLower(difficultyStr) {
		case "easy":
			difficulty = game.DifficultyEasy
		case "hard":
			difficulty = game.DifficultyHard
		default:
			difficulty = game.DifficultyMedium
		}

		playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
		initialChips := game.BigBlindAmt * 300 // 300BB
		g := game.NewGame(playerNames, initialChips, difficulty)

		// Main Game Loop (multi-hand)
		for {
			g.StartNewHand()

			// Single Hand Loop
			for {
				if g.CountPlayersInHand() <= 1 {
					break
				}

				if g.CountPlayersAbleToAct() < 2 {
					for g.Phase != game.PhaseShowdown {
						g.Advance()
					}
				}

				switch g.Phase {
				case game.PhasePreFlop, game.PhaseFlop, game.PhaseTurn, game.PhaseRiver:
					runInteractiveBettingRound(g)
					g.Advance()
				case game.PhaseShowdown, game.PhaseHandOver:
					break
				}

				if g.Phase == game.PhaseShowdown || g.Phase == game.PhaseHandOver {
					break
				}
			}

			// Conclude the hand
			if g.CountPlayersInHand() > 1 {
				cli.DisplayGameState(g)
				showdownResults(g)
			} else {
				fmt.Println("\n--- POT DISTRIBUTION ---")
				results := g.AwardPotToLastPlayer()
				for _, result := range results {
					fmt.Printf("%s wins %s chips with %s\n", result.PlayerName, util.FormatNumber(result.AmountWon), result.HandDesc)
				}
				fmt.Println("------------------------")
			}

			g.CleanupHand()

			if g.Players[0].Status == game.PlayerStatusEliminated {
				fmt.Println("\nYou have been eliminated. GAME OVER.")
				break
			}

			if g.CountRemainingPlayers() <= 1 {
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

	if g.CountPlayersInHand() < 2 {
		return
	}

	numPlayers := len(g.Players)
	actionCloserPos := 0

	if g.Phase == game.PhasePreFlop {
		actionCloserPos = (g.DealerPos + 2) % numPlayers
	} else {
		actionCloserPos = g.DealerPos
	}

	for {
		player := g.Players[g.CurrentTurnPos]

		if player.Status == game.PlayerStatusPlaying {
			cli.DisplayGameState(g)

			var action game.PlayerAction
			if player.IsCPU {
				action = g.GetCPUAction(player)
			} else {
				action = cli.PromptForAction(g)
			}

			wasAggressive := g.ProcessAction(player, action)
			if wasAggressive {
				actionCloserPos = (g.CurrentTurnPos - 1 + numPlayers) % numPlayers
			}
		}

		if g.CurrentTurnPos == actionCloserPos {
			break
		}

		g.CurrentTurnPos = g.FindNextActivePlayer(g.CurrentTurnPos)
	}
}

func showdownResults(g *game.Game) {
	fmt.Println("\n--- SHOWDOWN ---")

	distributionResults := g.DistributePot()

	winnerMap := make(map[string][]string)
	for _, result := range distributionResults {
		winType := ""
		if strings.HasPrefix(result.HandDesc, "High") || strings.HasPrefix(result.HandDesc, "takes") {
			winType = "High Winner"
		} else {
			winType = "Low Winner"
		}
		winnerMap[result.PlayerName] = append(winnerMap[result.PlayerName], winType)
	}

	for _, player := range g.Players {
		// FIX: Do not show eliminated or folded players in the showdown result list.
		if player.Status == game.PlayerStatusFolded || player.Status == game.PlayerStatusEliminated {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards)

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

		winnerStatus := ""
		if statuses, ok := winnerMap[player.Name]; ok {
			winnerStatus = fmt.Sprintf(" (%s)", strings.Join(statuses, " & "))
		}

		fmt.Printf("- %-7s: %v -> %s%s\n", player.Name, player.Hand, handDesc, winnerStatus)
	}

	fmt.Println("\n--- POT DISTRIBUTION ---")
	for _, result := range distributionResults {
		fmt.Printf("%s wins %s chips with %s\n", result.PlayerName, util.FormatNumber(result.AmountWon), result.HandDesc)
	}
	fmt.Println("------------------------")
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().StringVarP(&difficultyStr, "difficulty", "d", "medium", "Set AI difficulty (easy, medium, hard)")
}
