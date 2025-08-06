package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadGameRulesFromFile tests the loading and parsing of a game rule YAML file.
func TestLoadGameRulesFromFile(t *testing.T) {
	// Create a temporary directory and a dummy YAML file for the test.
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	err := os.Mkdir(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp rules dir: %v", err)
	}

	yamlContent := `
name: "Pot-Limit Sampyeong 7-or-Better"
abbreviation: "PLS7"
betting_limit: "pot_limit"
hole_cards:
  count: 3
  use_constraint: "any"
  use_count: 0
hand_rankings:
  enabled:
    - "royal_flush"
    - "skip_straight_flush"
    - "straight_flush"
    - "four_of_a_kind"
    - "full_house"
    - "flush"
    - "skip_straight"
    - "straight"
    - "three_of_a_kind"
    - "two_pair"
    - "one_pair"
    - "high_card"
low_hand:
  enabled: true
  max_rank: 7
`
	filePath := filepath.Join(rulesDir, "pls7.yml")
	err = os.WriteFile(filePath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp yaml file: %v", err)
	}

	// --- This is the function we are testing ---
	rules, err := LoadGameRulesFromFile(filePath)
	// ---

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// --- Assertions ---
	if rules.Name != "Pot-Limit Sampyeong 7-or-Better" {
		t.Errorf("Expected name to be 'Pot-Limit Sampyeong 7-or-Better', but got '%s'", rules.Name)
	}
	if rules.Abbreviation != "PLS7" {
		t.Errorf("Expected abbreviation to be 'PLS7', but got '%s'", rules.Abbreviation)
	}
	if rules.BettingLimit != "pot_limit" {
		t.Errorf("Expected betting_limit to be 'pot_limit', but got '%s'", rules.BettingLimit)
	}
	if rules.HoleCards.Count != 3 {
		t.Errorf("Expected hole_cards.count to be 3, but got %d", rules.HoleCards.Count)
	}
	if rules.HoleCards.UseConstraint != "any" {
		t.Errorf("Expected hole_cards.use_constraint to be 'any', but got '%s'", rules.HoleCards.UseConstraint)
	}
	if len(rules.HandRankings.Enabled) != 12 {
		t.Errorf("Expected 12 enabled hand rankings, but got %d", len(rules.HandRankings.Enabled))
	}
	if rules.HandRankings.Enabled[1] != "skip_straight_flush" {
		t.Errorf("Expected second enabled hand ranking to be 'skip_straight_flush', but got '%s'", rules.HandRankings.Enabled[1])
	}
	if !rules.LowHand.Enabled {
		t.Error("Expected low_hand.enabled to be true, but got false")
	}
	if rules.LowHand.MaxRank != 7 {
		t.Errorf("Expected low_hand.max_rank to be 7, but got %d", rules.LowHand.MaxRank)
	}
}
