package game

import "time"

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

// CPUThinkTime is the default thinking time for CPU actions
var CPUThinkTime = 500 * time.Millisecond
