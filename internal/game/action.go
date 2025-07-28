package game

// ActionType defines the type of action a player can take.
type ActionType int

const (
	ActionFold ActionType = iota
	ActionCheck
	ActionCall
	ActionBet
	ActionRaise
)

// String makes ActionType implement the Stringer interface.
func (at ActionType) String() string {
	return []string{"Fold", "Check", "Call", "Bet", "Raise"}[at]
}

// PlayerAction represents an action taken by a player.
type PlayerAction struct {
	Type   ActionType
	Amount int // Used for Bet or Raise
}

// ActionProvider is an interface that defines how to get a player's action.
// This allows us to use the real CLI prompt in the game, and a mock prompter in tests.
type ActionProvider interface {
	GetAction(g *Game) PlayerAction
}
