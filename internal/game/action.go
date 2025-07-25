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

// PlayerAction represents an action taken by a player.
type PlayerAction struct {
	Type   ActionType
	Amount int // Used for Bet or Raise
}
