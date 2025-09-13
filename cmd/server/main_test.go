package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"pls7-cli/internal/config"
	"pls7-cli/pkg/engine"
	"pls7-cli/pkg/network"
	"testing"
	"time"
)

// MockClient represents a mock client for testing
type MockClient struct {
	conn       net.Conn
	serverConn net.Conn
	received   []network.Message
	playerName string
	playerSeat int
}

// NewMockClient creates a new mock client connected to the server
func NewMockClient(serverAddr string) (*MockClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	return &MockClient{
		conn:     conn,
		received: make([]network.Message, 0),
	}, nil
}

// SendMessage sends a message to the server
func (mc *MockClient) SendMessage(msgType network.MessageType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := network.Message{
		Type: msgType,
		Data: data,
	}

	return json.NewEncoder(mc.conn).Encode(msg)
}

// ReceiveMessage receives a message from the server
func (mc *MockClient) ReceiveMessage() (*network.Message, error) {
	var msg network.Message
	err := json.NewDecoder(mc.conn).Decode(&msg)
	if err != nil {
		return nil, err
	}

	mc.received = append(mc.received, msg)
	return &msg, nil
}

// Identify sends an identification message to the server
func (mc *MockClient) Identify(name string) error {
	payload := network.IdentifyPayload{Name: name}
	err := mc.SendMessage(network.MsgIdentify, payload)
	if err != nil {
		return err
	}

	// Wait for welcome message
	msg, err := mc.ReceiveMessage()
	if err != nil {
		return err
	}

	if msg.Type != network.MsgWelcome {
		return fmt.Errorf("expected welcome message, got %s", msg.Type)
	}

	var welcomePayload network.WelcomePayload
	if err := json.Unmarshal(msg.Data, &welcomePayload); err != nil {
		return err
	}

	mc.playerName = welcomePayload.PlayerName
	mc.playerSeat = welcomePayload.PlayerSeat
	return nil
}

// SendAction sends a player action to the server
func (mc *MockClient) SendAction(action engine.PlayerAction) error {
	payload := network.PlayerActionPayload{Action: action}
	return mc.SendMessage(network.MsgPlayerAction, payload)
}

// Close closes the connection
func (mc *MockClient) Close() error {
	return mc.conn.Close()
}

// GetReceivedMessages returns all received messages
func (mc *MockClient) GetReceivedMessages() []network.Message {
	return mc.received
}

// TestServerBasicConnection tests basic server connection and client identification
func TestServerBasicConnection(t *testing.T) {
	// Change to project root directory for rules loading
	// This is needed because the test runs from cmd/server directory
	// but rules are in the project root
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Find project root (look for go.mod file)
	for {
		if _, err := os.Stat("go.mod"); err == nil {
			break
		}
		os.Chdir("..")
	}

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions("pls7")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}

	// Create server
	server := NewServer("8081", false, false, 2, 300000, 500, 1000, engine.DifficultyMedium, rules)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test client connection
	client, err := NewMockClient("localhost:8081")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	defer client.Close()

	// Test identification
	err = client.Identify("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to identify client: %v", err)
	}

	// Verify client was assigned correctly
	if client.playerName != "TestPlayer" {
		t.Errorf("Expected player name 'TestPlayer', got '%s'", client.playerName)
	}

	if client.playerSeat != 0 {
		t.Errorf("Expected player seat 0, got %d", client.playerSeat)
	}

	// Wait a bit for game to start
	time.Sleep(3 * time.Second)

	// Check if we received game state updates
	messages := client.GetReceivedMessages()
	gameStateReceived := false
	for _, msg := range messages {
		if msg.Type == network.MsgGameStateUpdate {
			gameStateReceived = true
			break
		}
	}

	if !gameStateReceived {
		t.Error("Expected to receive game state update")
	}

	// Stop server
	server.Stop()
}

// TestServerMultipleClients tests multiple clients connecting
func TestServerMultipleClients(t *testing.T) {
	// Change to project root directory for rules loading
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Find project root (look for go.mod file)
	for {
		if _, err := os.Stat("go.mod"); err == nil {
			break
		}
		os.Chdir("..")
	}

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions("pls7")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}

	// Create server
	server := NewServer("8082", false, false, 2, 300000, 500, 1000, engine.DifficultyMedium, rules)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create multiple clients
	clients := make([]*MockClient, 3)
	for i := 0; i < 3; i++ {
		client, err := NewMockClient("localhost:8082")
		if err != nil {
			t.Fatalf("Failed to create mock client %d: %v", i, err)
		}
		defer client.Close()

		err = client.Identify(fmt.Sprintf("Player%d", i))
		if err != nil {
			t.Fatalf("Failed to identify client %d: %v", i, err)
		}

		clients[i] = client
	}

	// Wait for game to start
	time.Sleep(3 * time.Second)

	// Verify all clients received game state updates
	for i, client := range clients {
		messages := client.GetReceivedMessages()
		gameStateReceived := false
		for _, msg := range messages {
			if msg.Type == network.MsgGameStateUpdate {
				gameStateReceived = true
				break
			}
		}

		if !gameStateReceived {
			t.Errorf("Client %d did not receive game state update", i)
		}
	}

	// Stop server
	server.Stop()
}

// TestServerPlayerAction tests player action handling
func TestServerPlayerAction(t *testing.T) {
	// Change to project root directory for rules loading
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Find project root (look for go.mod file)
	for {
		if _, err := os.Stat("go.mod"); err == nil {
			break
		}
		os.Chdir("..")
	}

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions("pls7")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}

	// Create server
	server := NewServer("8083", false, false, 2, 300000, 500, 1000, engine.DifficultyMedium, rules)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client, err := NewMockClient("localhost:8083")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	defer client.Close()

	// Identify client
	err = client.Identify("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to identify client: %v", err)
	}

	// Wait for game to start
	time.Sleep(3 * time.Second)

	// Send a player action (fold)
	action := engine.PlayerAction{Type: engine.ActionFold, Amount: 0}
	err = client.SendAction(action)
	if err != nil {
		t.Fatalf("Failed to send player action: %v", err)
	}

	// Wait a bit for the action to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if we received game events
	messages := client.GetReceivedMessages()
	gameEventReceived := false
	for _, msg := range messages {
		if msg.Type == network.MsgGameEvent {
			gameEventReceived = true
			break
		}
	}

	if !gameEventReceived {
		t.Error("Expected to receive game event after player action")
	}

	// Stop server
	server.Stop()
}

// TestServerBroadcast tests that game state updates are broadcast to all clients
func TestServerBroadcast(t *testing.T) {
	// Load game rules
	rules, err := config.LoadGameRulesFromOptions("pls7")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}

	// Create server
	server := NewServer("8084", false, false, 2, 300000, 500, 1000, engine.DifficultyMedium, rules)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create two clients
	client1, err := NewMockClient("localhost:8084")
	if err != nil {
		t.Fatalf("Failed to create mock client 1: %v", err)
	}
	defer client1.Close()

	client2, err := NewMockClient("localhost:8084")
	if err != nil {
		t.Fatalf("Failed to create mock client 2: %v", err)
	}
	defer client2.Close()

	// Identify both clients
	err = client1.Identify("Player1")
	if err != nil {
		t.Fatalf("Failed to identify client 1: %v", err)
	}

	err = client2.Identify("Player2")
	if err != nil {
		t.Fatalf("Failed to identify client 2: %v", err)
	}

	// Wait for game to start
	time.Sleep(3 * time.Second)

	// Send action from client 1
	action := engine.PlayerAction{Type: engine.ActionFold, Amount: 0}
	err = client1.SendAction(action)
	if err != nil {
		t.Fatalf("Failed to send action from client 1: %v", err)
	}

	// Wait for broadcast
	time.Sleep(200 * time.Millisecond)

	// Check that both clients received the game event
	messages1 := client1.GetReceivedMessages()
	messages2 := client2.GetReceivedMessages()

	eventReceived1 := false
	eventReceived2 := false

	for _, msg := range messages1 {
		if msg.Type == network.MsgGameEvent {
			eventReceived1 = true
			break
		}
	}

	for _, msg := range messages2 {
		if msg.Type == network.MsgGameEvent {
			eventReceived2 = true
			break
		}
	}

	if !eventReceived1 {
		t.Error("Client 1 did not receive game event")
	}

	if !eventReceived2 {
		t.Error("Client 2 did not receive game event (broadcast failed)")
	}

	// Stop server
	server.Stop()
}
