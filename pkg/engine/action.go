// Package engine provides the state machine and core logic for running a poker game.
// It manages the game flow, player actions, betting rounds, and pot distribution,
// using the rules and data structures defined in the `poker` package.
package engine

import "math/rand"

// ActionType defines the type of a player's action during a betting round.
type ActionType int

// ActionType constants represent the set of possible actions a player can take.
const (
	ActionFold  ActionType = iota // ActionFold signifies that the player forfeits their hand and any claim to the pot.
	ActionCheck                   // ActionCheck passes the action to the next player without betting, only possible when there is no open bet or raise.
	ActionCall                    // ActionCall matches the current bet amount.
	ActionBet                     // ActionBet is the first bet made in a betting round.
	ActionRaise                   // ActionRaise increases the size of the current bet.
)

// String returns the string representation of an ActionType (e.g., "Fold", "Check").
// It implements the fmt.Stringer interface.
func (at ActionType) String() string {
	return []string{"Fold", "Check", "Call", "Bet", "Raise"}[at]
}

// PlayerAction represents an action taken by a player, including the type of action
// and the amount for bets or raises.
type PlayerAction struct {
	// Type is the kind of action performed (e.g., Fold, Call, Raise).
	Type ActionType
	// Amount is the size of the bet or raise. It is only applicable for
	// ActionBet and ActionRaise actions. For other actions, it should be 0.
	Amount int
}

// ActionProvider is a crucial interface that decouples the game engine from the
// source of player input. By implementing this interface, different types of
// players (e.g., human CLI users, AI opponents, GUI clients) can be seamlessly
// integrated into the game. The engine calls GetAction when it's a player's
// turn, and the specific implementation provides the chosen action.
type ActionProvider interface {
	// GetAction is called by the game engine to request an action from a player.
	// The implementation should contain the logic for deciding on an action,
	// whether it's prompting a human user, running an AI algorithm, or receiving
	// input from a network client.
	//
	// Parameters:
	//   - g: A pointer to the current Game state, providing full context.
	//   - p: A pointer to the Player whose turn it is to act.
	//   - r: A source of randomness, primarily for use by AI players.
	//
	// Returns a PlayerAction struct representing the player's chosen action.
	GetAction(g *Game, p *Player, r *rand.Rand) PlayerAction
}
