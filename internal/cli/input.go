package cli

import (
	"bufio"
	"fmt"
	"os"
	"pls7-cli/internal/game"
	"strings"
)

// PromptForAction asks the human player for their move.
func PromptForAction(g *game.Game) game.PlayerAction {
	player := g.Players[g.CurrentTurnPos]
	canCheck := player.CurrentBet == g.BetToCall

	// Build the prompt string based on legal moves
	var prompt strings.Builder
	prompt.WriteString("Choose your action: ")

	if canCheck {
		prompt.WriteString("chec(k), (b)et, (f)old > ")
	} else {
		prompt.WriteString(fmt.Sprintf("(c)all %d, (r)aise, (f)old > ", g.BetToCall-player.CurrentBet))
	}

	fmt.Print(prompt.String())

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Parse input and return the corresponding action
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
		// TODO: Bet and Raise will be implemented in a later step
	}

	// Default or invalid action
	fmt.Println("Invalid action. Folding.")
	return game.PlayerAction{Type: game.ActionFold}
}
