package config

import (
	"fmt"
	os "os"
	"pls7-cli/pkg/poker"

	"gopkg.in/yaml.v3"
)

// LoadGameRulesFromFile reads a YAML file from the given path and returns a GameRules struct.
func LoadGameRulesFromFile(filePath string) (*poker.GameRules, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rules poker.GameRules
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

// LoadGameRulesFromBytes unmarshals a byte slice into a GameRules struct.
func LoadGameRulesFromBytes(data []byte) (*poker.GameRules, error) {
	var rules poker.GameRules
	err := yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}
	return &rules, nil
}

// LoadGameRulesFromOptions loads game rules from a YAML string by option value.
// - Available ruleStr: "pls", "pls7"
func LoadGameRulesFromOptions(ruleStr string) (*poker.GameRules, error) {
	filePath := fmt.Sprintf("rules/%s.yml", ruleStr)
	return LoadGameRulesFromFile(filePath)
}
