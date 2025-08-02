package cmd

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"strings"
)

var (
	difficultyStr string // To hold the flag value
	devMode       bool   // To hold the --dev flag value
	lowlessMode   bool   // To hold the --lowless flag value
	showOuts      bool   // To hold the --outs flag value (this does not work if devMode is true, as it will always show outs in dev mode)
)

// CLIActionProvider implements the ActionProvider interface using the CLI.
type CLIActionProvider struct{}

func (p *CLIActionProvider) GetAction(g *game.Game, _ *game.Player) game.PlayerAction {
	return cli.PromptForAction(g)
}

// CPUActionProvider implements the ActionProvider interface for CPU players.
type CPUActionProvider struct{}

func (p *CPUActionProvider) GetAction(g *game.Game, pl *game.Player) game.PlayerAction {
	return g.GetCPUAction(pl)
}

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Starts a new game of PLS7",
	Long:  `Starts a new game of PLS7 with 1 player and 5 CPUs.`,
	Run: func(cmd *cobra.Command, args []string) {
		util.InitLogger(devMode)

		fmt.Println("==================================================")
		fmt.Println("     PLS7 (Pot Limit Sampyong - 7 or better)")
		fmt.Println("==================================================")
		fmt.Printf("\nStarting the game with %s difficulty!\n", difficultyStr)

		playerActionProvider := &CLIActionProvider{}
		cpuActionProvider := &CPUActionProvider{}

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
		g := game.NewGame(playerNames, initialChips, difficulty, devMode, lowlessMode, showOuts)

		// Main Game Loop (multi-hand)
		for {
			g.StartNewHand()

			// Single Hand Loop
			for {
				if g.CountNonFoldedPlayers() <= 1 {
					break
				}

				if g.CountPlayersAbleToAct() < 2 {
					for g.Phase != game.PhaseShowdown {
						g.Advance()
					}
				}

				switch g.Phase {
				case game.PhasePreFlop, game.PhaseFlop, game.PhaseTurn, game.PhaseRiver:
					g.PrepareNewBettingRound()
					g.ExecuteBettingLoop(playerActionProvider, cpuActionProvider, cli.DisplayGameState)
					g.Advance()
				case game.PhaseShowdown, game.PhaseHandOver:
					break
				}

				if g.Phase == game.PhaseShowdown || g.Phase == game.PhaseHandOver {
					break
				}
			}

			// Conclude the hand
			if g.CountNonFoldedPlayers() > 1 {
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

			fmt.Print("\nPress ENTER to start the next hand, or type 'q' to exit > ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(input)) == "q" {
				fmt.Println("Thanks for playing!")
				break
			}
		}
	},
}

func showdownResults(g *game.Game) {
	output := "\n--- SHOWDOWN ---\n"

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
		if player.Status == game.PlayerStatusFolded || player.Status == game.PlayerStatusEliminated {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards, g.LowlessMode)

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

		output += fmt.Sprintf("- %-7s: %v -> %s%s\n", player.Name, player.Hand, handDesc, winnerStatus)
	}

	logrus.Debugf("distributionResults: %+v", distributionResults)

	output += fmt.Sprintln("\n--- POT DISTRIBUTION ---")
	for _, result := range distributionResults {
		output += fmt.Sprintf(
			"%s wins %s chips with %s\n",
			result.PlayerName, util.FormatNumber(result.AmountWon), result.HandDesc,
		)
	}
	output += fmt.Sprintln("------------------------")
	fmt.Println(output)
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().StringVarP(&difficultyStr, "difficulty", "d", "medium", "Set AI difficulty (easy, medium, hard)")
	playCmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode for verbose logging.")
	playCmd.Flags().BoolVar(&lowlessMode, "lowless", false, "Enable lowless mode (play with high hand only).")
	playCmd.Flags().BoolVar(&showOuts, "outs", false, "Shows outs for players if found.")
}
