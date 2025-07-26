package cli

import (
	"bufio"
	"fmt"
	"os"
	"pls7-cli/internal/game"
	"strconv"
	"strings"
)

// PromptForAction requests the player to choose an action during their turn.
func PromptForAction(g *game.Game) game.PlayerAction {
	player := g.Players[g.CurrentTurnPos]
	canCheck := player.CurrentBet == g.BetToCall

	var prompt strings.Builder
	prompt.WriteString("Choose your action: ")

	if canCheck {
		prompt.WriteString("chec(k), (b)et, (f)old > ")
	} else {
		prompt.WriteString(fmt.Sprintf("(c)all %d, (r)aise, (f)old > ", g.BetToCall-player.CurrentBet))
	}

	fmt.Print(prompt.String())
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// 입력을 파싱하여 해당하는 액션을 반환합니다.
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

	// 잘못된 액션 처리
	fmt.Println("Invalid action. Folding.")
	return game.PlayerAction{Type: game.ActionFold}
}

// promptForAmount는 베팅/레이즈 금액을 요청합니다.
func promptForAmount(g *game.Game, actionType game.ActionType) game.PlayerAction {
	minBet, maxBet := g.CalculateBettingLimits()
	// ActionType의 String() 메소드가 필요합니다. (다음 단계에서 추가 예정)
	actionName := "action"
	if actionType == game.ActionBet {
		actionName = "bet"
	} else if actionType == game.ActionRaise {
		actionName = "raise"
	}

	fmt.Printf("Enter amount to %s (minBet: %d, maxBet: %d): ", actionName, minBet, maxBet)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	amount, err := strconv.Atoi(strings.TrimSpace(input))

	if err != nil || amount < minBet || amount > maxBet {
		fmt.Println("Invalid amount. Folding.")
		return game.PlayerAction{Type: game.ActionFold}
	}

	return game.PlayerAction{Type: actionType, Amount: amount}
}
