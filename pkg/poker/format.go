package poker

// JoinStrings joins a slice of strings into a single string, with each element
// separated by a hyphen. For example, a slice `[]string{"A", "K", "Q"}` would
// become `"A-K-Q"`. This is a utility function for creating formatted string
// representations of card collections.
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
