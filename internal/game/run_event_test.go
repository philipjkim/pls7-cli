package game

import (
	"reflect"
	"testing"
)

func TestProcessAction_ReturnsActionEvent(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000)
	player := g.Players[0]

	// Test Call Action
	g.BetToCall = 1000
	player.CurrentBet = 0
	_, event := g.ProcessAction(player, PlayerAction{Type: ActionCall})
	expectedEvent := &ActionEvent{PlayerName: "YOU", Action: ActionCall, Amount: 1000}
	if !reflect.DeepEqual(event, expectedEvent) {
		t.Errorf("For Call, expected event %+v, got %+v", expectedEvent, event)
	}

	// Test Raise Action
	g.BetToCall = 1000
	player.CurrentBet = 1000
	_, event = g.ProcessAction(player, PlayerAction{Type: ActionRaise, Amount: 3000})
	expectedEvent = &ActionEvent{PlayerName: "YOU", Action: ActionRaise, Amount: 3000}
	if !reflect.DeepEqual(event, expectedEvent) {
		t.Errorf("For Raise, expected event %+v, got %+v", expectedEvent, event)
	}
}

func TestStartNewHand_ReturnsBlindEvent(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 100, 200)
	g.HandCount = 2 // This is the second hand, so the next hand will be the third
	g.BlindUpInterval = 2

	event := g.StartNewHand() // HandCount becomes 3, (3-1)%2 == 0, so blind up
	expectedEvent := &BlindEvent{SmallBlind: 200, BigBlind: 400}

	if !reflect.DeepEqual(event, expectedEvent) {
		t.Errorf("Expected blind event %+v, got %+v", expectedEvent, event)
	}
}
