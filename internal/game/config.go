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

// String returns a human-readable string representation of the Difficulty.
func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyMedium:
		return "Medium"
	case DifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}
