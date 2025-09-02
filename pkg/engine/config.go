package engine

// Difficulty represents the skill level of the AI opponents in the game.
// It can be used to select different AI profiles or adjust their behavior.
type Difficulty int

// Difficulty level constants.
const (
	DifficultyEasy   Difficulty = iota // DifficultyEasy represents the easiest AI opponents.
	DifficultyMedium                   // DifficultyMedium represents standard AI opponents.
	DifficultyHard                     // DifficultyHard represents the most challenging AI opponents.
)

// String returns a human-readable string representation of the Difficulty level.
// It implements the fmt.Stringer interface.
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
