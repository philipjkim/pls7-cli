// Package network defines the communication protocol and data structures for
// multiplayer functionality over TCP sockets.
package network

import (
	"encoding/json"
	"pls7-cli/pkg/engine"
)

// MessageType is a string alias for defining different kinds of messages
// sent between the server and clients. This provides a structured way to
// identify the purpose of each message.
type MessageType string

// Constants for all message types used in the protocol.
const (
	// --- Server -> Client Message Types ---

	// MsgWelcome is sent by the server to a client upon successful connection.
	// It provides the client with their assigned player details.
	MsgWelcome MessageType = "welcome"

	// MsgGameStateUpdate is sent by the server to all clients whenever the
	// public game state changes (e.g., after a player acts, at the start of a
	// new phase). It contains the full game state required for rendering.
	MsgGameStateUpdate MessageType = "game_state_update"

	// MsgGameEvent is sent by the server to all clients to announce a specific,
	// discrete event, such as a player folding or a blind increase.
	MsgGameEvent MessageType = "game_event"

	// MsgYourTurn is sent by the server to a single client to prompt them
	// to take their action.
	MsgYourTurn MessageType = "your_turn"

	// MsgError is sent by the server to a client to report an error, such
	// as an invalid action attempt.
	MsgError MessageType = "error"

	// --- Client -> Server Message Types ---

	// MsgPlayerAction is sent by a client to the server to communicate the
	// action they have chosen to take (e.g., fold, bet, raise).
	MsgPlayerAction MessageType = "player_action"

	// MsgIdentify is sent by a client to the server immediately after connecting
	// to provide a desired player name or identifier.
	MsgIdentify MessageType = "identify"
)

// Message is the fundamental data structure for all communication. It wraps every
// message with a specific type, and the actual data is contained as raw JSON.
// This allows the receiver to parse the generic message first, check the type,
// and then unmarshal the data into the correct specific struct.
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// --- Payload Structs for Specific Message Types ---

// WelcomePayload is the data structure for the MsgWelcome message.
type WelcomePayload struct {
	PlayerName     string `json:"player_name"`
	PlayerSeat     int    `json:"player_seat"`
	WelcomeMessage string `json:"welcome_message"`
}

// GameStateUpdatePayload is the data structure for the MsgGameStateUpdate message.
// It embeds the main Game struct from the engine.
type GameStateUpdatePayload struct {
	Game *engine.Game `json:"game"`
}

// GameEventPayload is the data structure for the MsgGameEvent message.
// It can be used for various announcements.
type GameEventPayload struct {
	Message string `json:"message"`
	// Optionally, include the original ActionEvent for more structured data
	Event *engine.ActionEvent `json:"event,omitempty"`
}

// YourTurnPayload is the data structure for the MsgYourTurn message.
// It might include valid actions to help the client UI.
type YourTurnPayload struct {
	// Currently empty, but could be extended with valid actions, bet ranges, etc.
}

// ErrorPayload is the data structure for the MsgError message.
type ErrorPayload struct {
	Message string `json:"message"`
}

// PlayerActionPayload is the data structure for the MsgPlayerAction message.
type PlayerActionPayload struct {
	Action engine.PlayerAction `json:"action"`
}

// IdentifyPayload is the data structure for the MsgIdentify message.
type IdentifyPayload struct {
	Name string `json:"name"`
}
