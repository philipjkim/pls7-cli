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
var PlayerHoleCardsForDebug = map[string]string{
	"3As":        "As Ah Ad", // For testing outs for Four of a Kind and Full House
	"AQT-suited": "As Qs Ts", // For testing outs for Flush, Straight, and Skip Straight
	"AAK":        "As Ah Ks", // For testing outs for Three of a Kind
	"A23-suited": "As 2s 3s", // For testing outs for Straight, Flush, and low hand scenarios
}
