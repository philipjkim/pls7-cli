package cmd

import (
	"bufio"
	"fmt"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/config"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/poker"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ruleStr       string // To hold the --rule flag value (load rules/{rule}.yml when the game starts)
	difficultyStr string // To hold the flag value
	devMode       bool   // To hold the --dev flag value
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

func runGame(cmd *cobra.Command, args []string) {
	util.InitLogger(devMode)

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions(ruleStr)
	if err != nil {
		logrus.Fatalf("Failed to load game rules: %v", err)
	}

	fmt.Printf("======== %s ========\n", rules.Name)
	fmt.Printf("Starting the game with %s difficulty!\n", difficultyStr)

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
	g := game.NewGame(playerNames, initialChips, difficulty, rules, devMode, showOuts)

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
			fmt.Println("--- POT DISTRIBUTION ---")
			results := g.AwardPotToLastPlayer()
			for _, result := range results {
				fmt.Printf(
					"%s wins %s chips with %s\n",
					result.PlayerName, util.FormatNumber(result.AmountWon), result.HandDesc,
				)
			}
			fmt.Println("------------------------")
		}

		g.CleanupHand()

		if g.Players[0].Status == game.PlayerStatusEliminated {
			fmt.Println("You have been eliminated. GAME OVER.")
			break
		}

		if g.CountRemainingPlayers() <= 1 {
			fmt.Println("--- GAME OVER ---")
			break
		}

		fmt.Print("Press ENTER to start the next hand, or type 'q' to exit > ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(input)) == "q" {
			fmt.Println("Thanks for playing!")
			break
		}
	}
}

func showdownResults(g *game.Game) {
	output := "\n--- SHOWDOWN ---\n"
	output += fmt.Sprintf("Community Cards: %s\n", g.CommunityCards)

	distributionResults := g.DistributePot()

	winnerMap := make(map[string][]string)
	for _, result := range distributionResults {
		winType := ""
		if strings.HasPrefix(result.HandDesc, "High") || strings.HasPrefix(result.HandDesc, "takes") {
			winType = "High Winner"
		} else if strings.HasPrefix(result.HandDesc, "Scoop") {
			winType = "High/Low Winner"
		} else {
			winType = "Low Winner"
		}
		winnerMap[result.PlayerName] = append(winnerMap[result.PlayerName], winType)

		// For debugging purposes, log the showdown results for "YOU"
		if result.PlayerName == "YOU" {
			logrus.Debugf(
				"showdownResults: YOU - handDesc=%v, winType=%v, winnerMapValue=%+v",
				result.HandDesc, winType, winnerMap[result.PlayerName],
			)
		}
	}
	logrus.Debugf(
		"showdownResults: winnerMap=%+v, distributionResults=%+v",
		winnerMap, distributionResults,
	)

	for _, player := range g.Players {
		if player.Status == game.PlayerStatusFolded || player.Status == game.PlayerStatusEliminated {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards, g.Rules)

		handDesc := highHand.String()
		if g.Rules.LowHand.Enabled && lowHand != nil {
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pls7",
	Short: "Starts a new game of PLS7",
	Long:  `Starts a new game of PLS7 with 1 player and 5 CPUs.`,
	Run:   runGame,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&ruleStr, "game rule", "r", "pls7", "Game rule to use (pls7, pls).")
	rootCmd.Flags().StringVarP(&difficultyStr, "difficulty", "d", "medium", "Set AI difficulty (easy, medium, hard)")
	rootCmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode for verbose logging.")
	rootCmd.Flags().BoolVar(&showOuts, "outs", false, "Shows outs for players if found (temporarily draws fixed good hole cards).")
}
