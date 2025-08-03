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

// PlayerHoleCardsForDebug is YOU (human player) hole cards for debugging purposes.
const PlayerHoleCardsForDebug = "As Ah Ad" // For testing outs for Four of a Kind and Full House
//const PlayerHoleCardsForDebug = "As Qs Ts" // For testing outs for Flush, Straight, and Skip Straight
