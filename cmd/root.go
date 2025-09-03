package cmd

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/engine"
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
	initialChips    int    // To hold the --initial-chips flag value
	smallBlind      int    // To hold the --small-blind flag value
	bigBlind        int    // To hold the --big-blind flag value
)

// CLIActionProvider implements the ActionProvider interface using the CLI.
type CLIActionProvider struct{}

func (p *CLIActionProvider) GetAction(g *engine.Game, _ *engine.Player, _ *rand.Rand) engine.PlayerAction {
	return cli.PromptForAction(g)
}

// CPUActionProvider implements the ActionProvider interface for CPU players.
type CPUActionProvider struct{}

func (p *CPUActionProvider) GetAction(g *engine.Game, pl *engine.Player, r *rand.Rand) engine.PlayerAction {
	return g.GetCPUAction(pl, r)
}

// CombinedActionProvider decides which provider to use based on player type.
type CombinedActionProvider struct{}

// GetAction method for CombinedActionProvider
func (p *CombinedActionProvider) GetAction(g *engine.Game, player *engine.Player, r *rand.Rand) engine.PlayerAction {
	if player.IsCPU {
		time.Sleep(g.CPUThinkTime())
		return g.GetCPUAction(player, r)
	}
	return cli.PromptForAction(g)
}

func runGame(_ *cobra.Command, _ []string) {
	util.InitLogger(devMode)

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions(ruleStr)
	if err != nil {
		logrus.Fatalf("Failed to load game rules: %v", err)
	}

	fmt.Printf("======== %s ========\n", rules.Name)

	playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}

	var difficulty engine.Difficulty
	switch difficultyStr {
	case "easy":
		difficulty = engine.DifficultyEasy
	case "medium":
		difficulty = engine.DifficultyMedium
	case "hard":
		difficulty = engine.DifficultyHard
	default:
		logrus.Warnf("Invalid difficulty '%s' specified. Defaulting to medium.", difficultyStr)
		difficulty = engine.DifficultyMedium
	}

	g := engine.NewGame(playerNames, initialChips, smallBlind, bigBlind, difficulty, rules, devMode, showOuts, blindUpInterval)

	actionProvider := &CombinedActionProvider{}

	// Main Game Loop (multi-hand)
	for {
		cli.DisplayGameState(g)

		blindEvent := g.StartNewHand()
		if blindEvent != nil {
			message := fmt.Sprintf("\n*** Blinds are now %s/%s ***\n", cli.FormatNumber(blindEvent.SmallBlind), cli.FormatNumber(blindEvent.BigBlind))
			fmt.Println(message)
		}

		// Single Hand Loop
		for g.Phase != engine.PhaseShowdown && g.Phase != engine.PhaseHandOver {
			if g.CountNonFoldedPlayers() <= 1 {
				break
			}
			g.PrepareNewBettingRound()

			// New Turn-by-turn Betting Loop
			for !g.IsBettingRoundOver() {
				player := g.CurrentPlayer()
				var action engine.PlayerAction

				if player.Status != engine.PlayerStatusPlaying {
					g.AdvanceTurn()
					continue
				}

				action = actionProvider.GetAction(g, player, g.Rand)

				_, event := g.ProcessAction(player, action)
				if event != nil {
					var eventMessage string
					switch event.Action {
					case engine.ActionFold:
						eventMessage = fmt.Sprintf("%s folds.", event.PlayerName)
					case engine.ActionCheck:
						eventMessage = fmt.Sprintf("%s checks.", event.PlayerName)
					case engine.ActionCall:
						eventMessage = fmt.Sprintf("%s calls %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					case engine.ActionBet:
						eventMessage = fmt.Sprintf("%s bets %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					case engine.ActionRaise:
						eventMessage = fmt.Sprintf("%s raises to %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					}
					if eventMessage != "" {
						fmt.Println(eventMessage)
					}
				}
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
			fmt.Println("--- POT AWARDED ---")
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

		if g.Players[0].Status == engine.PlayerStatusEliminated {
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
	rootCmd.Flags().IntVar(&initialChips, "initial-chips", 300000, "Initial chips for each player.")
	rootCmd.Flags().IntVar(&smallBlind, "small-blind", 500, "Small blind amount.")
	rootCmd.Flags().IntVar(&bigBlind, "big-blind", 1000, "Big blind amount.")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if initialChips <= 0 {
			return fmt.Errorf("initial-chips는 0보다 커야 합니다. 입력값: %d", initialChips)
		}
		if smallBlind <= 0 {
			return fmt.Errorf("small-blind는 0보다 커야 합니다. 입력값: %d", smallBlind)
		}
		if bigBlind <= 0 {
			return fmt.Errorf("big-blind는 0보다 커야 합니다. 입력값: %d", bigBlind)
		}
		if smallBlind >= bigBlind {
			return fmt.Errorf("small-blind(%d)는 big-blind(%d)보다 작아야 합니다", smallBlind, bigBlind)
		}
		return nil
	}
}
