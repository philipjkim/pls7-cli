package util

import "strconv"

// FormatNumber takes an integer and returns a string with commas as thousands separators.
func FormatNumber(n int) string {
	s := strconv.Itoa(n)
	length := len(s)
	if length <= 3 {
		return s
	}

	firstGroupLen := length % 3
	if firstGroupLen == 0 {
		firstGroupLen = 3
	}

	result := s[:firstGroupLen]

	for i := firstGroupLen; i < length; i += 3 {
		result += "," + s[i:i+3]
	}

	return result
}
