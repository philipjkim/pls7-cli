package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/engine"
	"pls7-cli/pkg/network"
	"pls7-cli/pkg/poker"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Server represents the multiplayer poker server
type Server struct {
	game         *engine.Game
	clients      map[net.Conn]*Client
	clientsMux   sync.RWMutex
	listener     net.Listener
	port         string
	devMode      bool
	showOuts     bool
	blindUp      int
	initialChips int
	smallBlind   int
	bigBlind     int
	difficulty   engine.Difficulty
	rules        *poker.GameRules
	done         chan bool // Channel to signal server shutdown
}

// Client represents a connected client
type Client struct {
	conn       net.Conn
	playerName string
	playerSeat int
	server     *Server
}

// NewServer creates a new server instance
func NewServer(port string, devMode, showOuts bool, blindUp, initialChips, smallBlind, bigBlind int, difficulty engine.Difficulty, rules *poker.GameRules) *Server {
	return &Server{
		clients:      make(map[net.Conn]*Client),
		port:         port,
		devMode:      devMode,
		showOuts:     showOuts,
		blindUp:      blindUp,
		initialChips: initialChips,
		smallBlind:   smallBlind,
		bigBlind:     bigBlind,
		difficulty:   difficulty,
		rules:        rules,
		done:         make(chan bool),
	}
}

// Start starts the server and begins accepting connections
func (s *Server) Start() error {
	util.InitLogger(s.devMode)

	var err error
	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to start server on port %s: %v", s.port, err)
	}

	logrus.Infof("Server started on port %s", s.port)
	logrus.Infof("Waiting for clients to connect...")

	// Accept connections in a loop with shutdown signal
	for {
		select {
		case <-s.done:
			logrus.Info("Server shutdown requested")
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if server is shutting down
				select {
				case <-s.done:
					return nil
				default:
					logrus.Errorf("Failed to accept connection: %v", err)
					continue
				}
			}

			// Handle each connection in a separate goroutine
			go s.handleClient(conn)
		}
	}
}

// Stop stops the server
func (s *Server) Stop() error {
	close(s.done) // Signal shutdown

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleClient handles a single client connection
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		conn:   conn,
		server: s,
	}

	logrus.Infof("New client connected from %s", conn.RemoteAddr().String())

	// Add client to the server's client list
	s.clientsMux.Lock()
	s.clients[conn] = client
	s.clientsMux.Unlock()

	// Remove client when connection closes
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients, conn)
		s.clientsMux.Unlock()
		logrus.Infof("Client %s disconnected", conn.RemoteAddr().String())
	}()

	// Set connection timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Handle incoming messages from this client
	decoder := json.NewDecoder(conn)
	for {
		var msg network.Message
		if err := decoder.Decode(&msg); err != nil {
			// Check if it's a timeout or connection closed error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logrus.Warnf("Client %s timed out", conn.RemoteAddr().String())
			} else {
				logrus.Errorf("Failed to decode message from client %s: %v", conn.RemoteAddr().String(), err)
			}
			break
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		if err := s.handleMessage(client, msg); err != nil {
			logrus.Errorf("Failed to handle message from client %s: %v", conn.RemoteAddr().String(), err)
			s.sendError(client, err.Error())
		}
	}
}

// handleMessage processes incoming messages from clients
func (s *Server) handleMessage(client *Client, msg network.Message) error {
	switch msg.Type {
	case network.MsgIdentify:
		var payload network.IdentifyPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal identify payload: %v", err)
		}
		return s.handleIdentify(client, payload)

	case network.MsgPlayerAction:
		var payload network.PlayerActionPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal player action payload: %v", err)
		}
		return s.handlePlayerAction(client, payload)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleIdentify processes client identification
func (s *Server) handleIdentify(client *Client, payload network.IdentifyPayload) error {
	client.playerName = payload.Name

	// Assign a seat to the client
	client.playerSeat = len(s.clients) - 1 // -1 because we're counting after adding the client

	// Send welcome message
	welcomePayload := network.WelcomePayload{
		PlayerName:     client.playerName,
		PlayerSeat:     client.playerSeat,
		WelcomeMessage: fmt.Sprintf("Welcome %s! You are seated at position %d", client.playerName, client.playerSeat),
	}

	return s.sendMessage(client, network.MsgWelcome, welcomePayload)
}

// handlePlayerAction processes player actions
func (s *Server) handlePlayerAction(client *Client, payload network.PlayerActionPayload) error {
	if s.game == nil {
		return fmt.Errorf("game not started yet")
	}

	// Find the player in the game
	player := s.findPlayerByClient(client)
	if player == nil {
		return fmt.Errorf("player not found in game")
	}

	// Process the action
	_, event := s.game.ProcessAction(player, payload.Action)

	// Broadcast game state update to all clients
	s.broadcastGameStateUpdate()

	// If there was an event, broadcast it
	if event != nil {
		s.broadcastGameEvent(event)
	}

	return nil
}

// findPlayerByClient finds the game player associated with a client
func (s *Server) findPlayerByClient(client *Client) *engine.Player {
	if s.game == nil || client.playerSeat >= len(s.game.Players) {
		return nil
	}
	return s.game.Players[client.playerSeat]
}

// sendMessage sends a message to a specific client
func (s *Server) sendMessage(client *Client, msgType network.MessageType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	msg := network.Message{
		Type: msgType,
		Data: data,
	}

	return json.NewEncoder(client.conn).Encode(msg)
}

// sendError sends an error message to a specific client
func (s *Server) sendError(client *Client, message string) error {
	payload := network.ErrorPayload{Message: message}
	return s.sendMessage(client, network.MsgError, payload)
}

// broadcastGameStateUpdate sends the current game state to all connected clients
func (s *Server) broadcastGameStateUpdate() {
	if s.game == nil {
		return
	}

	payload := network.GameStateUpdatePayload{Game: s.game}

	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	for _, client := range s.clients {
		if err := s.sendMessage(client, network.MsgGameStateUpdate, payload); err != nil {
			logrus.Errorf("Failed to send game state update to client %s: %v", client.conn.RemoteAddr().String(), err)
		}
	}
}

// broadcastGameEvent sends a game event to all connected clients
func (s *Server) broadcastGameEvent(event *engine.ActionEvent) {
	payload := network.GameEventPayload{
		Message: fmt.Sprintf("%s %s", event.PlayerName, event.Action),
		Event:   event,
	}

	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	for _, client := range s.clients {
		if err := s.sendMessage(client, network.MsgGameEvent, payload); err != nil {
			logrus.Errorf("Failed to send game event to client %s: %v", client.conn.RemoteAddr().String(), err)
		}
	}
}

// startGame initializes and starts a new game
func (s *Server) startGame() error {
	// Create player names based on connected clients
	playerNames := make([]string, 6) // Max 6 players

	s.clientsMux.RLock()
	clientCount := len(s.clients)
	clientList := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clientList = append(clientList, client)
	}
	s.clientsMux.RUnlock()

	// Assign human players first
	for i, client := range clientList {
		if i < 6 {
			playerNames[i] = client.playerName
		}
	}

	// Fill remaining slots with CPU players
	for i := clientCount; i < 6; i++ {
		playerNames[i] = fmt.Sprintf("CPU %d", i+1)
	}

	// Create the game
	s.game = engine.NewGame(playerNames, s.initialChips, s.smallBlind, s.bigBlind, s.difficulty, s.rules, s.devMode, s.showOuts, s.blindUp)

	logrus.Info("Game started")

	// Broadcast initial game state
	s.broadcastGameStateUpdate()

	return nil
}

func main() {
	// Default configuration
	port := "8080"
	devMode := false
	showOuts := false
	blindUp := 2
	initialChips := 300000
	smallBlind := 500
	bigBlind := 1000
	difficulty := engine.DifficultyMedium
	ruleStr := "pls7"

	// Parse command line arguments (simple flag parsing for now)
	if len(os.Args) > 1 {
		for i, arg := range os.Args[1:] {
			switch arg {
			case "--port", "-p":
				if i+1 < len(os.Args) {
					port = os.Args[i+2]
				}
			case "--dev":
				devMode = true
			case "--outs":
				showOuts = true
			case "--rule", "-r":
				if i+1 < len(os.Args) {
					ruleStr = os.Args[i+2]
				}
			case "--difficulty", "-d":
				if i+1 < len(os.Args) {
					switch os.Args[i+2] {
					case "easy":
						difficulty = engine.DifficultyEasy
					case "medium":
						difficulty = engine.DifficultyMedium
					case "hard":
						difficulty = engine.DifficultyHard
					}
				}
			case "--help", "-h":
				fmt.Println("Usage: go run cmd/server/main.go [options]")
				fmt.Println("Options:")
				fmt.Println("  --port, -p PORT        Server port (default: 8080)")
				fmt.Println("  --dev                  Enable development mode")
				fmt.Println("  --outs                 Show outs for players")
				fmt.Println("  --rule, -r RULE        Game rule (pls7, pls, nlh, plo, plo8)")
				fmt.Println("  --difficulty, -d DIFF  AI difficulty (easy, medium, hard)")
				fmt.Println("  --help, -h             Show this help message")
				os.Exit(0)
			}
		}
	}

	// Load game rules
	rules, err := config.LoadGameRulesFromOptions(ruleStr)
	if err != nil {
		log.Fatalf("Failed to load game rules: %v", err)
	}

	fmt.Printf("Starting multiplayer poker server on port %s\n", port)
	fmt.Printf("Game rule: %s\n", rules.Name)
	fmt.Printf("Difficulty: %s\n", difficulty)
	if devMode {
		fmt.Println("Development mode enabled")
	}

	// Create and start server
	server := NewServer(port, devMode, showOuts, blindUp, initialChips, smallBlind, bigBlind, difficulty, rules)

	// Start the game when we have at least one client
	go func() {
		time.Sleep(2 * time.Second) // Wait a bit for clients to connect
		if err := server.startGame(); err != nil {
			logrus.Errorf("Failed to start game: %v", err)
		}
	}()

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
