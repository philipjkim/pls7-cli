package cli

import (
	"bufio"
	"fmt"
	"os"
	"pls7-cli/internal/game"
	"pls7-cli/internal/util"
	"strconv"
	"strings"
)

// PromptForAction requests the player to choose an action during their turn.
func PromptForAction(g *game.Game) game.PlayerAction {
	DisplayGameState(g)

	// for loop to keep prompting until a valid action is chosen
	for {
		player := g.Players[g.CurrentTurnPos]
		canCheck := player.CurrentBet == g.BetToCall
		amountToCall := g.BetToCall - player.CurrentBet

		var prompt strings.Builder
		prompt.WriteString("Choose your action: ")

		if canCheck {
			prompt.WriteString("chec(k), (b)et, (f)old > ")
		} else {
			// If amountToCall is negative, it means remaining players have bet all-in with less than the current bet.
			// So the player does not need to act anything, call.
			if amountToCall < 0 {
				return game.PlayerAction{Type: game.ActionCall}
			}

			prompt.WriteString(fmt.Sprintf("(c)all %s, ", util.FormatNumber(amountToCall)))
			// Only show raise option if the player has enough chips to make a valid raise.
			minRaise, _ := g.CalculateBettingLimits()
			if player.Chips > amountToCall && player.CurrentBet+player.Chips >= minRaise {
				prompt.WriteString("(r)aise, ")
			}
			prompt.WriteString("(f)old > ")
		}

		fmt.Print(prompt.String())
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "f":
			return game.PlayerAction{Type: game.ActionFold}
		case "k":
			if canCheck {
				return game.PlayerAction{Type: game.ActionCheck}
			}
		case "c":
			if !canCheck {
				return game.PlayerAction{Type: game.ActionCall}
			}
		case "b":
			if canCheck {
				return promptForAmount(g, game.ActionBet)
			}
		case "r":
			if !canCheck {
				return promptForAmount(g, game.ActionRaise)
			}
		}

		fmt.Println("Invalid action.")
	}
}

// promptForAmount requests the betting/raising amount.
func promptForAmount(g *game.Game, actionType game.ActionType) game.PlayerAction {
	for {
		minBet, maxBet := g.CalculateBettingLimits()
		actionName := "bet"
		if actionType == game.ActionRaise {
			actionName = "raise to"
		}

		fmt.Printf(
			"Enter amount to %s (min: %s, max: %s): ",
			actionName, util.FormatNumber(minBet), util.FormatNumber(maxBet),
		)

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		amount, err := strconv.Atoi(strings.TrimSpace(input))

		if err != nil || amount < minBet || amount > maxBet {
			fmt.Println("Invalid amount. Please try again.")
		} else {
			return game.PlayerAction{Type: actionType, Amount: amount}
		}
	}
}
