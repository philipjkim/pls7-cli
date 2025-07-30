package game

// Game settings constants
const (
	SmallBlindAmt = 500
	BigBlindAmt   = 1000
)

// Difficulty defines the AI difficulty level.
type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyMedium
	DifficultyHard
)
