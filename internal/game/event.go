package game

// ActionEvent represents a player's action and the data associated with it.
type ActionEvent struct {
	PlayerName string
	Action     ActionType
	Amount     int
}

// BlindEvent represents a blind level update.
type BlindEvent struct {
	SmallBlind int
	BigBlind   int
}
