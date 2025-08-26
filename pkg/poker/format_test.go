package poker

import (
	"testing"
)

func TestFormatCards(t *testing.T) {
	threeCards := []string{"A", "K", "Q"}
	expected := "A-K-Q"
	actual := JoinStrings(threeCards)
	if actual != expected {
		t.Errorf("Expected 'A-K-Q', got '%s'", actual)
	}
}
