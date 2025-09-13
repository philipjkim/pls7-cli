// Package network defines the communication protocol and data structures for
// multiplayer functionality over TCP sockets.
package network

import (
	"encoding/json"
	"pls7-cli/pkg/engine"
	"pls7-cli/pkg/poker"
	"reflect"
	"testing"
)

func TestProtocolSerialization(t *testing.T) {
	testCases := []struct {
		name         string
		messageType  MessageType
		payload      interface{}
		payloadType  reflect.Type
		originalData interface{}
	}{
		{
			name:        "Welcome Message",
			messageType: MsgWelcome,
			payload: &WelcomePayload{
				PlayerName:     "YOU",
				PlayerSeat:     0,
				WelcomeMessage: "Welcome to the game!",
			},
			payloadType: reflect.TypeOf(&WelcomePayload{}),
		},
		{
			name:        "Game State Update Message",
			messageType: MsgGameStateUpdate,
			payload: &GameStateUpdatePayload{
				Game: &engine.Game{Pot: 1000, Phase: engine.PhaseFlop},
			},
			payloadType: reflect.TypeOf(&GameStateUpdatePayload{}),
		},
		{
			name:        "Game Event Message",
			messageType: MsgGameEvent,
			payload: &GameEventPayload{
				Message: "Player 2 folds",
				Event:   &engine.ActionEvent{PlayerName: "Player 2", Action: engine.ActionFold},
			},
			payloadType: reflect.TypeOf(&GameEventPayload{}),
		},
		{
			name:        "Your Turn Message",
			messageType: MsgYourTurn,
			payload:     &YourTurnPayload{},
			payloadType: reflect.TypeOf(&YourTurnPayload{}),
		},
		{
			name:        "Error Message",
			messageType: MsgError,
			payload:     &ErrorPayload{Message: "Invalid action"},
			payloadType: reflect.TypeOf(&ErrorPayload{}),
		},
		{
			name:        "Player Action Message",
			messageType: MsgPlayerAction,
			payload: &PlayerActionPayload{
				Action: engine.PlayerAction{Type: engine.ActionBet, Amount: 500},
			},
			payloadType: reflect.TypeOf(&PlayerActionPayload{}),
		},
		{
			name:        "Identify Message",
			messageType: MsgIdentify,
			payload:     &IdentifyPayload{Name: "John"},
			payloadType: reflect.TypeOf(&IdentifyPayload{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Marshal the specific payload into raw JSON data
			rawData, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// 2. Create the top-level message
			originalMessage := Message{
				Type: tc.messageType,
				Data: rawData,
			}

			// 3. Marshal the full message
			messageBytes, err := json.Marshal(originalMessage)
			if err != nil {
				t.Fatalf("Failed to marshal full message: %v", err)
			}

			// 4. Unmarshal the full message back
			var unmarshaledMessage Message
			if err := json.Unmarshal(messageBytes, &unmarshaledMessage); err != nil {
				t.Fatalf("Failed to unmarshal full message: %v", err)
			}

			// 5. Check if the type is correct
			if unmarshaledMessage.Type != tc.messageType {
				t.Errorf("Expected message type %s, got %s", tc.messageType, unmarshaledMessage.Type)
			}

			// 6. Unmarshal the raw data into the specific payload type
			newPayload := reflect.New(tc.payloadType.Elem()).Interface()
			if err := json.Unmarshal(unmarshaledMessage.Data, newPayload); err != nil {
				t.Fatalf("Failed to unmarshal payload data: %v", err)
			}

			// 7. Compare the original payload with the unmarshaled one
			if !reflect.DeepEqual(tc.payload, newPayload) {
				t.Errorf("Payloads do not match after serialization/deserialization.\nOriginal: %+v\nReceived: %+v", tc.payload, newPayload)
			}
		})
	}
}

// TestGameStateUpdatePayloadWithRealGame tests serialization with a more complex Game object.
func TestGameStateUpdatePayloadWithRealGame(t *testing.T) {
	// Setup a realistic game state by defining rules directly in the test
	// to avoid dependency on internal packages or the file system.
	rules := &poker.GameRules{
		Name:         "Pot-Limit Sampyeong 7-or-Better",
		Abbreviation: "PLS7",
		BettingLimit: "pot_limit",
		HoleCards: poker.HoleCardRules{
			Count:         3,
			UseConstraint: "any",
		},
		HandRankings: poker.HandRankingsRules{
			UseStandardRankings: false,
			CustomRankings: []poker.CustomHandRanking{
				{Name: "skip_straight_flush", InsertAfterRank: "royal_flush"},
				{Name: "skip_straight", InsertAfterRank: "flush"},
			},
		},
		LowHand: poker.LowHandRules{
			Enabled: true,
			MaxRank: 7,
		},
	}

	game := engine.NewGame([]string{"YOU", "CPU1", "CPU2"}, 10000, 100, 200, engine.DifficultyEasy, rules, true, false, 0)
	game.StartNewHand()
	game.CommunityCards = poker.CardsFromStrings("As Ks Qs")

	payload := &GameStateUpdatePayload{
		Game: game,
	}

	// Marshal and unmarshal
	rawData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	var newPayload GameStateUpdatePayload
	if err := json.Unmarshal(rawData, &newPayload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Basic comparison (DeepEqual is complex for the full Game object with pointers)
	if newPayload.Game.Pot != game.Pot {
		t.Errorf("Expected pot %d, got %d", game.Pot, newPayload.Game.Pot)
	}
	if newPayload.Game.Phase != game.Phase {
		t.Errorf(
			"Expected phase %s, got %s",
			game.Phase, newPayload.Game.Phase,
		)
	}
	if len(newPayload.Game.Players) != len(game.Players) {
		t.Fatalf("Expected %d players, got %d", len(game.Players), len(newPayload.Game.Players))
	}
	if newPayload.Game.Players[0].Name != game.Players[0].Name {
		t.Errorf("Expected player 0 name %s, got %s", game.Players[0].Name, newPayload.Game.Players[0].Name)
	}
}
