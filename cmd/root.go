package cmd

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/config"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ruleStr         string // To hold the --rule flag value (load rules/{rule}.yml when the game starts)
	difficultyStr   string // To hold the flag value
	devMode         bool   // To hold the --dev flag value
	showOuts        bool   // To hold the --outs flag value (this does not work if devMode is true, as it will always show outs in dev mode)
	blindUpInterval int    // To hold the --blind-up flag value
)

// CLIActionProvider implements the ActionProvider interface using the CLI.
type CLIActionProvider struct{}

func (p *CLIActionProvider) GetAction(g *game.Game, _ *game.Player, r *rand.Rand) game.PlayerAction {
	return cli.PromptForAction(g)
}

// CPUActionProvider implements the ActionProvider interface for CPU players.
type CPUActionProvider struct{}

func (p *CPUActionProvider) GetAction(g *game.Game, pl *game.Player, r *rand.Rand) game.PlayerAction {
	return g.GetCPUAction(pl, r)
}

// CombinedActionProvider decides which provider to use based on player type.
type CombinedActionProvider struct{}

// GetAction method for CombinedActionProvider
func (p *CombinedActionProvider) GetAction(g *game.Game, player *game.Player, r *rand.Rand) game.PlayerAction {
	if player.IsCPU {
		time.Sleep(g.CPUThinkTime())
		return g.GetCPUAction(player, r)
	}
	return cli.PromptForAction(g)
}

func runGame(cmd *cobra.Command, args []string) {
	util.InitLogger(devMode)

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions(ruleStr)
	if err != nil {
		logrus.Fatalf("Failed to load game rules: %v", err)
	}

	fmt.Printf("======== %s ========\n", rules.Name)

	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}
	initialChips := game.BigBlindAmt * 300

	var difficulty game.Difficulty
	switch difficultyStr {
	case "easy":
		difficulty = game.DifficultyEasy
	case "medium":
		difficulty = game.DifficultyMedium
	case "hard":
		difficulty = game.DifficultyHard
	default:
		logrus.Warnf("Invalid difficulty '%s' specified. Defaulting to medium.", difficultyStr)
		difficulty = game.DifficultyMedium
	}

	g := game.NewGame(playerNames, initialChips, difficulty, rules, devMode, showOuts, blindUpInterval)

	actionProvider := &CombinedActionProvider{}

	// Main Game Loop (multi-hand)
	for {
		cli.DisplayGameState(g)

		blindMessage := g.StartNewHand()
		if blindMessage != "" {
			fmt.Println(blindMessage)
		}

		// Single Hand Loop
		for g.Phase != game.PhaseShowdown && g.Phase != game.PhaseHandOver {
			if g.CountNonFoldedPlayers() <= 1 {
				break
			}
			g.PrepareNewBettingRound()

			// New Turn-by-turn Betting Loop
			for !g.IsBettingRoundOver() {
				player := g.CurrentPlayer()
				var action game.PlayerAction

				if player.Status != game.PlayerStatusPlaying {
					g.AdvanceTurn()
					continue
				}

				action = actionProvider.GetAction(g, player, g.Rand)

				_, eventMessage := g.ProcessAction(player, action)
				if eventMessage != "" {
					fmt.Println(eventMessage)
				}
				time.Sleep(300 * time.Millisecond) // Delay after action
				g.AdvanceTurn()
			}
			g.Advance()
		}

		// Conclude the hand
		if g.CountNonFoldedPlayers() > 1 {
			showdownMessages := cli.FormatShowdownResults(g)
			for _, msg := range showdownMessages {
				fmt.Println(msg)
			}
		} else {
			results := g.AwardPotToLastPlayer()
			fmt.Println("--- POT DISTRIBUTION ---")
			for _, result := range results {
				fmt.Printf(
					"%s wins %s chips with %s\n",
					result.PlayerName, cli.FormatNumber(result.AmountWon), result.HandDesc,
				)
			}
			fmt.Println("------------------------")
		}

		cleanupMessages := g.CleanupHand()
		for _, msg := range cleanupMessages {
			fmt.Println(msg)
		}

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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pls7",
	Short: "Starts a new game of Poker",
	Long:  `Starts a new game of Poker (PLS7, PLS, NLH) with 1 player and 5 CPUs.`, // Corrected escaping for backticks and quotes within the string literal. The original string was fine.
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
	rootCmd.Flags().StringVarP(&ruleStr, "rule", "r", "pls7", "Game rule to use (pls7, pls, nlh).")
	rootCmd.Flags().StringVarP(&difficultyStr, "difficulty", "d", "medium", "Set AI difficulty (easy, medium, hard)")
	rootCmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode for verbose logging.")
	rootCmd.Flags().BoolVar(&showOuts, "outs", false, "Shows outs for players if found (temporarily draws fixed good hole cards).")
	rootCmd.Flags().IntVar(&blindUpInterval, "blind-up", 2, "Sets the number of rounds for blind up. 0 means no blind up.")
}
