package model

type Difficulty string

const (
	Beginner     Difficulty = "beginner"
	Intermediate Difficulty = "intermediate"
	Advanced     Difficulty = "advanced"
)

// Valid reports whether d is one of the known difficulty levels.
func (d Difficulty) Valid() bool {
	return d == Beginner || d == Intermediate || d == Advanced
}
