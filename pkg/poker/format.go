package poker

// JoinStrings joins a slice of strings with hyphens such as "A-K-Q-J-10".
func JoinStrings(s []string) string {
	if len(s) == 0 {
		return ""
	}
	var formatted string
	for i, item := range s {
		if i > 0 {
			formatted += "-"
		}
		formatted += item
	}
	return formatted
}
