package engine

// ActionEvent represents a significant action taken by a player during a betting
// round. It is intended to be used for logging, display, or broadcasting game
// state changes to observers like a UI.
type ActionEvent struct {
	// PlayerName is the name of the player who performed the action.
	PlayerName string
	// Action is the type of action taken (e.g., Fold, Call, Raise).
	Action ActionType
	// Amount is the value associated with the action, such as the size of a
	// bet or raise. It is 0 for actions like Fold and Check.
	Amount int
}

// BlindEvent represents the posting of the small and big blinds at the beginning
// of a hand. It can be used to announce the current blind levels.
type BlindEvent struct {
	// SmallBlind is the size of the small blind.
	SmallBlind int
	// BigBlind is the size of the big blind.
	BigBlind int
}
